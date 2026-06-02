package example_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/example"
	"github.com/elloloop/project-scaffold/packages/go/testkit"
)

func TestNormalizeTag(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"trims and lowercases", "  Hello  ", "hello"},
		{"collapses whitespace to dashes", "High Priority Bug", "high-priority-bug"},
		{"already normal", "done", "done"},
		{"empty", "   ", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := example.NormalizeTag(tc.in); got != tc.want {
				t.Errorf("NormalizeTag(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

type noteRepo struct {
	store *testkit.FakeStore
}

func (r noteRepo) Save(ctx context.Context, n example.Note) error {
	return r.store.Create(ctx, testkit.Record{
		TenantID: n.TenantID,
		Type:     "Note",
		ID:       n.ID,
		Fields:   map[string]any{"author": n.Author, "body": n.Body},
	})
}

func (r noteRepo) Get(ctx context.Context, tenantID, id string) (example.Note, error) {
	rec, err := r.store.Get(ctx, tenantID, "Note", id)
	if errors.Is(err, testkit.ErrNotFound) {
		return example.Note{}, example.ErrNoteNotFound
	}
	if err != nil {
		return example.Note{}, err
	}
	return example.Note{
		ID:       rec.ID,
		TenantID: rec.TenantID,
		Author:   rec.Fields["author"].(string),
		Body:     rec.Fields["body"].(string),
	}, nil
}

func newService(t *testing.T) (*example.NoteService, *testkit.FakeStore) {
	t.Helper()
	store := testkit.NewFakeStore(testkit.NewFakeClock(testkit.FixedClockTime))
	svc := example.NewNoteService(noteRepo{store: store}, testkit.NewSeqIDGen("note"))
	return svc, store
}

func TestNoteService_CreateGet_TenantScoped(t *testing.T) {
	t.Parallel()
	ctx := testkit.Context(t)
	svc, _ := newService(t)

	created, err := svc.Create(ctx, "tenantA", "user:1", "hello")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := svc.Get(ctx, "tenantA", created.ID)
	if err != nil || got.Body != "hello" {
		t.Fatalf("get in-tenant: %v %+v", err, got)
	}
	if _, err := svc.Get(ctx, "tenantB", created.ID); !errors.Is(err, example.ErrNoteNotFound) {
		t.Errorf("cross-tenant read should be NotFound, got %v", err)
	}
}

func TestNoteHandler_CreateAndGet(t *testing.T) {
	t.Parallel()
	svc, _ := newService(t)
	h := example.NewNoteHandler(svc)
	user := testkit.NewUser("tenantA")

	req := testkit.AsUser(testkit.NewConnectRequest(t, "/example.NoteService/CreateNote", map[string]any{"body": "hi"}), user.Claims())
	resp := testkit.Serve(http.HandlerFunc(h.CreateNote), req)
	testkit.AssertStatus(t, resp, http.StatusOK)
	note := testkit.DecodeJSON[example.Note](t, resp.Body)
	if note.Body != "hi" || note.TenantID != "tenantA" {
		t.Fatalf("unexpected note: %+v", note)
	}

	getReq := testkit.AsUser(testkit.NewConnectRequest(t, "/example.NoteService/GetNote", map[string]any{"id": note.ID}), user.Claims())
	getResp := testkit.Serve(http.HandlerFunc(h.GetNote), getReq)
	testkit.AssertStatus(t, getResp, http.StatusOK)
}

func TestNoteHandler_AnonymousRejected(t *testing.T) {
	t.Parallel()
	svc, _ := newService(t)
	h := example.NewNoteHandler(svc)
	resp := testkit.Serve(http.HandlerFunc(h.CreateNote),
		testkit.NewConnectRequest(t, "/example.NoteService/CreateNote", map[string]any{"body": "x"}))
	testkit.AssertStatus(t, resp, http.StatusUnauthorized)
	testkit.AssertConnectError(t, resp, "unauthenticated")
}

func TestNoteHandler_GetIsTenantIsolated(t *testing.T) {
	t.Parallel()
	svc, _ := newService(t)
	h := example.NewNoteHandler(svc)

	owner := testkit.NewUser("tenantA")
	createResp := testkit.Serve(http.HandlerFunc(h.CreateNote),
		testkit.AsUser(testkit.NewConnectRequest(t, "/example.NoteService/CreateNote", map[string]any{"body": "secret"}), owner.Claims()))
	note := testkit.DecodeJSON[example.Note](t, createResp.Body)

	attacker := testkit.NewUser("tenantB")
	resp := testkit.Serve(http.HandlerFunc(h.GetNote),
		testkit.AsUser(testkit.NewConnectRequest(t, "/example.NoteService/GetNote", map[string]any{"id": note.ID}), attacker.Claims()))
	testkit.AssertStatus(t, resp, http.StatusNotFound)
	testkit.AssertConnectError(t, resp, "not_found")
}

func TestRegistrarClient(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		es := testkit.NewExternalService(t, map[string]http.HandlerFunc{
			"/availability": testkit.RespondJSON(http.StatusOK, map[string]any{"available": true}),
		})
		c := example.NewRegistrarClient(es.URL, es.Client())
		ok, err := c.CheckAvailability(testkit.Context(t), "example.dev")
		if err != nil || !ok {
			t.Fatalf("success: ok=%v err=%v", ok, err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		es := testkit.NewExternalService(t, map[string]http.HandlerFunc{
			"/availability": testkit.RespondStatus(http.StatusInternalServerError),
		})
		c := example.NewRegistrarClient(es.URL, es.Client())
		if _, err := c.CheckAvailability(testkit.Context(t), "x"); err == nil {
			t.Error("expected error on 500")
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		es := testkit.NewExternalService(t, map[string]http.HandlerFunc{
			"/availability": testkit.RespondMalformed(),
		})
		c := example.NewRegistrarClient(es.URL, es.Client())
		if _, err := c.CheckAvailability(testkit.Context(t), "x"); err == nil {
			t.Error("expected decode error on malformed body")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		es := testkit.NewExternalService(t, map[string]http.HandlerFunc{
			"/availability": testkit.RespondHang(),
		})
		client := es.Client()
		client.Timeout = 50 * time.Millisecond
		c := example.NewRegistrarClient(es.URL, client)
		if _, err := c.CheckAvailability(context.Background(), "x"); err == nil {
			t.Error("expected timeout error")
		}
	})
}

func TestNoNetworkGuard(t *testing.T) {
	testkit.NoNetwork(t)
	_, err := http.Get("http://should-not-resolve.invalid")
	if !errors.Is(err, testkit.ErrNetworkDisabled) {
		t.Errorf("expected network to be blocked, got %v", err)
	}
}
