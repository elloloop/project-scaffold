package testkit

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// Repository/DB errors surfaced by FakeStore. They mirror the failure modes a
// real database-backed repository must handle, so service-layer tests can assert
// the same behaviour without a database.
var (
	ErrNotFound            = errors.New("testkit: record not found")
	ErrUniqueViolation     = errors.New("testkit: unique constraint violation")
	ErrForeignKeyViolation = errors.New("testkit: foreign key violation")
	ErrTenantRequired      = errors.New("testkit: tenant id required")
)

// Ref is a foreign-key reference to another record in the same tenant.
type Ref struct {
	Type string
	ID   string
}

// Record is a generic, tenant-scoped stored entity. It models the parts of a
// row that matter for testing persistence rules: identity, unique keys,
// foreign keys, soft-delete, timestamps, and arbitrary fields.
type Record struct {
	TenantID   string
	Type       string
	ID         string
	UniqueKeys map[string]string // constraint name -> value (scoped per tenant+type)
	Refs       map[string]Ref    // fk name -> target
	Fields     map[string]any
	Deleted    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r Record) clone() Record {
	cp := r
	cp.UniqueKeys = cloneStrMap(r.UniqueKeys)
	cp.Fields = cloneAnyMap(r.Fields)
	if r.Refs != nil {
		cp.Refs = make(map[string]Ref, len(r.Refs))
		for k, v := range r.Refs {
			cp.Refs[k] = v
		}
	}
	return cp
}

// FakeStore is an in-memory, tenant-scoped record store. Every operation is
// tenant-qualified; a read for the wrong tenant returns ErrNotFound, which is
// what makes multi-tenant isolation tests trivial to write. It is safe for
// concurrent use.
type FakeStore struct {
	mu    sync.Mutex
	clock ports.Clock
	// data[tenant][type][id] = record
	data map[string]map[string]map[string]Record
	uniq map[string]bool // "tenant|type|constraint|value"
}

// NewFakeStore builds an empty store using clock for timestamps.
func NewFakeStore(clock ports.Clock) *FakeStore {
	return &FakeStore{
		clock: clock,
		data:  map[string]map[string]map[string]Record{},
		uniq:  map[string]bool{},
	}
}

// Create inserts rec, enforcing unique and foreign-key constraints. It sets
// CreatedAt/UpdatedAt from the clock.
func (s *FakeStore) Create(_ context.Context, rec Record) error {
	if rec.TenantID == "" {
		return ErrTenantRequired
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lookup(rec.TenantID, rec.Type, rec.ID); exists {
		return ErrUniqueViolation
	}
	if err := s.checkUnique(rec); err != nil {
		return err
	}
	if err := s.checkRefs(rec); err != nil {
		return err
	}
	now := s.clock.Now()
	rec.CreatedAt, rec.UpdatedAt = now, now
	s.put(rec)
	s.takeUnique(rec)
	return nil
}

// Get returns the live record identified by (tenant, typ, id). A soft-deleted
// record or a wrong-tenant read both return ErrNotFound.
func (s *FakeStore) Get(_ context.Context, tenant, typ, id string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.lookup(tenant, typ, id)
	if !ok || rec.Deleted {
		return Record{}, ErrNotFound
	}
	return rec.clone(), nil
}

// Update replaces an existing record, re-checking unique and FK constraints.
func (s *FakeStore) Update(_ context.Context, rec Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.lookup(rec.TenantID, rec.Type, rec.ID)
	if !ok {
		return ErrNotFound
	}
	s.freeUnique(old)
	if err := s.checkUnique(rec); err != nil {
		s.takeUnique(old) // roll back
		return err
	}
	if err := s.checkRefs(rec); err != nil {
		s.takeUnique(old)
		return err
	}
	rec.CreatedAt = old.CreatedAt
	rec.UpdatedAt = s.clock.Now()
	s.put(rec)
	s.takeUnique(rec)
	return nil
}

// Delete hard-deletes a record and frees its unique keys.
func (s *FakeStore) Delete(_ context.Context, tenant, typ, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.lookup(tenant, typ, id)
	if !ok {
		return ErrNotFound
	}
	s.freeUnique(rec)
	delete(s.data[tenant][typ], id)
	return nil
}

// SoftDelete marks a record deleted without removing it; its unique keys are
// freed so a replacement can be created.
func (s *FakeStore) SoftDelete(_ context.Context, tenant, typ, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.lookup(tenant, typ, id)
	if !ok || rec.Deleted {
		return ErrNotFound
	}
	s.freeUnique(rec)
	rec.Deleted = true
	rec.UpdatedAt = s.clock.Now()
	s.put(rec)
	return nil
}

// ListOptions controls List: exact-match field filters, sort field/direction,
// and offset/limit pagination. Soft-deleted records are excluded unless
// IncludeDeleted is set.
type ListOptions struct {
	Filter         map[string]any
	SortBy         string
	Desc           bool
	Limit          int
	Offset         int
	IncludeDeleted bool
}

// Page is a slice of a result set plus the metadata a paginated API returns.
type Page struct {
	Items      []Record
	Total      int // total matching the filter, before pagination
	Limit      int
	Offset     int
	NextOffset int // -1 when there are no more items
}

// List returns records of typ in tenant, filtered, sorted, and paginated.
func (s *FakeStore) List(_ context.Context, tenant, typ string, opts ListOptions) (Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var matched []Record
	for _, rec := range s.data[tenant][typ] {
		if rec.Deleted && !opts.IncludeDeleted {
			continue
		}
		if matchesFilter(rec, opts.Filter) {
			matched = append(matched, rec.clone())
		}
	}

	if opts.SortBy != "" {
		sort.SliceStable(matched, func(i, j int) bool {
			less := fieldLess(matched[i].Fields[opts.SortBy], matched[j].Fields[opts.SortBy])
			if opts.Desc {
				return fieldLess(matched[j].Fields[opts.SortBy], matched[i].Fields[opts.SortBy])
			}
			return less
		})
	} else {
		// Stable, deterministic default order by id.
		sort.SliceStable(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })
	}

	total := len(matched)
	lo := clamp(opts.Offset, 0, total)
	hi := total
	if opts.Limit > 0 {
		hi = clamp(lo+opts.Limit, lo, total)
	}
	page := matched[lo:hi]
	next := -1
	if hi < total {
		next = hi
	}
	return Page{Items: page, Total: total, Limit: opts.Limit, Offset: opts.Offset, NextOffset: next}, nil
}

// Count returns the number of live records of typ in tenant.
func (s *FakeStore) Count(ctx context.Context, tenant, typ string) int {
	p, _ := s.List(ctx, tenant, typ, ListOptions{})
	return p.Total
}

// -- internals --------------------------------------------------------------

func (s *FakeStore) lookup(tenant, typ, id string) (Record, bool) {
	rec, ok := s.data[tenant][typ][id]
	return rec, ok
}

func (s *FakeStore) put(rec Record) {
	if s.data[rec.TenantID] == nil {
		s.data[rec.TenantID] = map[string]map[string]Record{}
	}
	if s.data[rec.TenantID][rec.Type] == nil {
		s.data[rec.TenantID][rec.Type] = map[string]Record{}
	}
	s.data[rec.TenantID][rec.Type][rec.ID] = rec.clone()
}

func uniqKey(tenant, typ, name, value string) string {
	return tenant + "|" + typ + "|" + name + "|" + value
}

func (s *FakeStore) checkUnique(rec Record) error {
	for name, value := range rec.UniqueKeys {
		if s.uniq[uniqKey(rec.TenantID, rec.Type, name, value)] {
			return fmt.Errorf("%w: %s=%s", ErrUniqueViolation, name, value)
		}
	}
	return nil
}

func (s *FakeStore) takeUnique(rec Record) {
	for name, value := range rec.UniqueKeys {
		s.uniq[uniqKey(rec.TenantID, rec.Type, name, value)] = true
	}
}

func (s *FakeStore) freeUnique(rec Record) {
	for name, value := range rec.UniqueKeys {
		delete(s.uniq, uniqKey(rec.TenantID, rec.Type, name, value))
	}
}

func (s *FakeStore) checkRefs(rec Record) error {
	for name, ref := range rec.Refs {
		target, ok := s.lookup(rec.TenantID, ref.Type, ref.ID)
		if !ok || target.Deleted {
			return fmt.Errorf("%w: %s -> %s/%s", ErrForeignKeyViolation, name, ref.Type, ref.ID)
		}
	}
	return nil
}

func matchesFilter(rec Record, filter map[string]any) bool {
	for k, want := range filter {
		if !reflect.DeepEqual(rec.Fields[k], want) {
			return false
		}
	}
	return true
}

// fieldLess orders the common comparable field types; anything else falls back
// to string comparison so sorting is always defined.
func fieldLess(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, _ := b.(string)
		return av < bv
	case int:
		bv, _ := b.(int)
		return av < bv
	case int64:
		bv, _ := b.(int64)
		return av < bv
	case float64:
		bv, _ := b.(float64)
		return av < bv
	case time.Time:
		bv, _ := b.(time.Time)
		return av.Before(bv)
	default:
		return fmt.Sprint(a) < fmt.Sprint(b)
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func cloneStrMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
