package testkit

import (
	"context"
	"errors"
	"testing"
)

func newRec(tenant, id string, fields map[string]any) Record {
	return Record{TenantID: tenant, Type: "Issue", ID: id, Fields: fields}
}

func TestFakeStore_CRUDAndSoftDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := NewFakeStore(NewFakeClock(FixedClockTime))

	if err := s.Create(ctx, newRec("t1", "i1", map[string]any{"title": "a"})); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.Get(ctx, "t1", "Issue", "i1")
	if err != nil || got.Fields["title"] != "a" {
		t.Fatalf("get: %v %+v", err, got)
	}

	got.Fields["title"] = "b"
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if g, _ := s.Get(ctx, "t1", "Issue", "i1"); g.Fields["title"] != "b" {
		t.Errorf("update not persisted: %+v", g)
	}

	if err := s.SoftDelete(ctx, "t1", "Issue", "i1"); err != nil {
		t.Fatalf("soft-delete: %v", err)
	}
	if _, err := s.Get(ctx, "t1", "Issue", "i1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("soft-deleted record should be NotFound, got %v", err)
	}
}

func TestFakeStore_TenantIsolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := NewFakeStore(NewFakeClock(FixedClockTime))
	_ = s.Create(ctx, newRec("tenantA", "i1", nil))

	// A record created in tenantA is invisible to tenantB - the core isolation
	// guarantee every multi-tenant test relies on.
	if _, err := s.Get(ctx, "tenantB", "Issue", "i1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("cross-tenant read leaked: err = %v", err)
	}
}

func TestFakeStore_Constraints(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := NewFakeStore(NewFakeClock(FixedClockTime))

	mb := Record{TenantID: "t1", Type: "Mailbox", ID: "m1", UniqueKeys: map[string]string{"address": "a@x.test"}}
	if err := s.Create(ctx, mb); err != nil {
		t.Fatalf("create mailbox: %v", err)
	}
	dup := Record{TenantID: "t1", Type: "Mailbox", ID: "m2", UniqueKeys: map[string]string{"address": "a@x.test"}}
	if err := s.Create(ctx, dup); !errors.Is(err, ErrUniqueViolation) {
		t.Errorf("duplicate unique key should fail, got %v", err)
	}

	orphan := Record{TenantID: "t1", Type: "Message", ID: "msg1", Refs: map[string]Ref{"mailbox": {Type: "Mailbox", ID: "missing"}}}
	if err := s.Create(ctx, orphan); !errors.Is(err, ErrForeignKeyViolation) {
		t.Errorf("dangling FK should fail, got %v", err)
	}
}

func TestFakeStore_ListFilterSortPaginate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := NewFakeStore(NewFakeClock(FixedClockTime))
	for i, p := range []int{3, 1, 2, 1} {
		_ = s.Create(ctx, newRec("t1", string(rune('a'+i)), map[string]any{"priority": p, "open": i != 3}))
	}

	// Filter open issues, sort by priority ascending, first page of 2.
	page, _ := s.List(ctx, "t1", "Issue", ListOptions{
		Filter: map[string]any{"open": true},
		SortBy: "priority",
		Limit:  2,
		Offset: 0,
	})
	AssertPage(t, page, 3, 2, true)
	if page.Items[0].Fields["priority"].(int) > page.Items[1].Fields["priority"].(int) {
		t.Errorf("not sorted ascending: %+v", page.Items)
	}

	second, _ := s.List(ctx, "t1", "Issue", ListOptions{Filter: map[string]any{"open": true}, SortBy: "priority", Limit: 2, Offset: 2})
	AssertPage(t, second, 3, 1, false)
}
