package testkit

import (
	"context"
	"errors"
	"testing"
)

// Store is the persistence contract the repository layer is built on. FakeStore
// (in-memory, for unit tests) and any real database adapter (database today, a
// Postgres adapter tomorrow - for integration tests) both satisfy it. That is
// the "swap the database" guarantee: the domain depends on this interface, not
// on database, so a new backend is a new Store implementation behind the same
// contract.
//
// RunStoreContract is the SINGLE behavioural suite for that contract. Run it
// against FakeStore in unit tests and against a real adapter under
// `-tags=integration`, and both backends are proven to behave identically.
type Store interface {
	Create(ctx context.Context, rec Record) error
	Get(ctx context.Context, tenantID, typ, id string) (Record, error)
	Update(ctx context.Context, rec Record) error
	Delete(ctx context.Context, tenantID, typ, id string) error
	SoftDelete(ctx context.Context, tenantID, typ, id string) error
	List(ctx context.Context, tenantID, typ string, opts ListOptions) (Page, error)
}

var _ Store = (*FakeStore)(nil)

// RunStoreContract exercises the full Store contract against any implementation
// produced by newStore (called fresh per subtest for isolation). A real adapter
// must map its errors to the testkit sentinels (ErrNotFound, ErrUniqueViolation,
// ErrForeignKeyViolation) so the same assertions hold across backends.
func RunStoreContract(t *testing.T, newStore func() Store) {
	t.Helper()
	ctx := context.Background()
	rec := func(tenant, id string, fields map[string]any) Record {
		return Record{TenantID: tenant, Type: "Item", ID: id, Fields: fields}
	}

	t.Run("create/read/update", func(t *testing.T) {
		t.Parallel()
		s := newStore()
		if err := s.Create(ctx, rec("t1", "a", map[string]any{"name": "x"})); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := s.Get(ctx, "t1", "Item", "a")
		if err != nil || got.Fields["name"] != "x" {
			t.Fatalf("get: %v %+v", err, got)
		}
		got.Fields["name"] = "y"
		if err := s.Update(ctx, got); err != nil {
			t.Fatalf("update: %v", err)
		}
		if g, _ := s.Get(ctx, "t1", "Item", "a"); g.Fields["name"] != "y" {
			t.Errorf("update not persisted: %+v", g)
		}
	})

	t.Run("soft delete hides", func(t *testing.T) {
		t.Parallel()
		s := newStore()
		_ = s.Create(ctx, rec("t1", "a", nil))
		if err := s.SoftDelete(ctx, "t1", "Item", "a"); err != nil {
			t.Fatalf("soft-delete: %v", err)
		}
		if _, err := s.Get(ctx, "t1", "Item", "a"); !errors.Is(err, ErrNotFound) {
			t.Errorf("soft-deleted should be NotFound, got %v", err)
		}
	})

	t.Run("tenant isolation", func(t *testing.T) {
		t.Parallel()
		s := newStore()
		_ = s.Create(ctx, rec("tenantA", "a", nil))
		if _, err := s.Get(ctx, "tenantB", "Item", "a"); !errors.Is(err, ErrNotFound) {
			t.Errorf("cross-tenant read leaked: %v", err)
		}
	})

	t.Run("unique + foreign-key constraints", func(t *testing.T) {
		t.Parallel()
		s := newStore()
		_ = s.Create(ctx, Record{TenantID: "t1", Type: "Mailbox", ID: "m1", UniqueKeys: map[string]string{"addr": "a@x.test"}})
		dup := Record{TenantID: "t1", Type: "Mailbox", ID: "m2", UniqueKeys: map[string]string{"addr": "a@x.test"}}
		if err := s.Create(ctx, dup); !errors.Is(err, ErrUniqueViolation) {
			t.Errorf("dup unique key should fail, got %v", err)
		}
		orphan := Record{TenantID: "t1", Type: "Msg", ID: "x", Refs: map[string]Ref{"mbox": {Type: "Mailbox", ID: "missing"}}}
		if err := s.Create(ctx, orphan); !errors.Is(err, ErrForeignKeyViolation) {
			t.Errorf("dangling FK should fail, got %v", err)
		}
	})

	t.Run("list filter/sort/paginate", func(t *testing.T) {
		t.Parallel()
		s := newStore()
		for i, p := range []int{3, 1, 2} {
			_ = s.Create(ctx, rec("t1", string(rune('a'+i)), map[string]any{"prio": p, "open": true}))
		}
		page, err := s.List(ctx, "t1", "Item", ListOptions{Filter: map[string]any{"open": true}, SortBy: "prio", Limit: 2})
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		AssertPage(t, page, 3, 2, true)
		if page.Items[0].Fields["prio"].(int) > page.Items[1].Fields["prio"].(int) {
			t.Errorf("not sorted ascending: %+v", page.Items)
		}
	})
}
