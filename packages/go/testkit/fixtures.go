package testkit

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// Role values used across authorization tests.
const (
	RoleUser     = "user"
	RoleAdmin    = "admin"
	RoleReadOnly = "readonly"
)

// RecordTypeUser is the FakeStore record type used by SeedUser.
const RecordTypeUser = "User"

var uniqueSeq atomic.Uint64

// Unique returns a process-unique value with the given prefix, so fixtures
// never collide even across parallel tests.
func Unique(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, uniqueSeq.Add(1))
}

// Tenant is a minimal tenant fixture.
type Tenant struct {
	ID   string
	Slug string
	Name string
}

// NewTenant returns a fresh tenant with unique id/slug.
func NewTenant() Tenant {
	id := Unique("tenant")
	return Tenant{ID: id, Slug: id, Name: "Test " + id}
}

// User is a test principal. It maps to ports.Claims for handler tests and can
// be seeded into a FakeStore as a record.
type User struct {
	ID        string
	Email     string
	Name      string
	Role      string
	TenantID  string
	Suspended bool
	Deleted   bool
}

// Claims renders the user as verified identity claims.
func (u User) Claims() ports.Claims {
	return ports.Claims{
		Subject:  u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Role:     u.Role,
		TenantID: u.TenantID,
	}
}

// Actor is the database actor string convention, "user:{id}".
func (u User) Actor() string { return "user:" + u.ID }

// NewUser builds a normal user in tenantID. Apply mods to tweak fields.
func NewUser(tenantID string, mods ...func(*User)) User {
	id := Unique("user")
	u := User{
		ID:       id,
		Email:    id + "@example.test",
		Name:     "Test " + id,
		Role:     RoleUser,
		TenantID: tenantID,
	}
	for _, m := range mods {
		m(&u)
	}
	return u
}

// NewAdmin builds an admin user in tenantID.
func NewAdmin(tenantID string, mods ...func(*User)) User {
	return NewUser(tenantID, append([]func(*User){func(u *User) { u.Role = RoleAdmin }}, mods...)...)
}

// NewReadOnlyUser builds a read-only user in tenantID.
func NewReadOnlyUser(tenantID string, mods ...func(*User)) User {
	return NewUser(tenantID, append([]func(*User){func(u *User) { u.Role = RoleReadOnly }}, mods...)...)
}

// NewSuspendedUser builds a suspended user in tenantID.
func NewSuspendedUser(tenantID string, mods ...func(*User)) User {
	return NewUser(tenantID, append([]func(*User){func(u *User) { u.Suspended = true }}, mods...)...)
}

// NewDeletedUser builds a soft-deleted user in tenantID.
func NewDeletedUser(tenantID string, mods ...func(*User)) User {
	return NewUser(tenantID, append([]func(*User){func(u *User) { u.Deleted = true }}, mods...)...)
}

// OwnerAndNonOwner returns two distinct normal users in the same tenant, for
// object-level authorization tests (owner may act, non-owner may not).
func OwnerAndNonOwner(tenantID string) (owner, nonOwner User) {
	return NewUser(tenantID), NewUser(tenantID)
}

// SeedUser inserts u into store as a User record with a unique email key and a
// suspended/deleted-aware state, and returns the stored record.
func SeedUser(t testing.TB, store *FakeStore, u User) Record {
	t.Helper()
	rec := Record{
		TenantID:   u.TenantID,
		Type:       RecordTypeUser,
		ID:         u.ID,
		UniqueKeys: map[string]string{"email": u.Email},
		Fields: map[string]any{
			"email":     u.Email,
			"name":      u.Name,
			"role":      u.Role,
			"suspended": u.Suspended,
		},
		Deleted: u.Deleted,
	}
	if err := store.Create(context.Background(), rec); err != nil {
		t.Fatalf("testkit: seed user %s: %v", u.ID, err)
	}
	if u.Deleted {
		if err := store.SoftDelete(context.Background(), u.TenantID, RecordTypeUser, u.ID); err != nil {
			t.Fatalf("testkit: soft-delete seeded user %s: %v", u.ID, err)
		}
	}
	return rec
}
