// Package example is a self-contained reference showing how to structure and
// test service code against shared seams. It is not wired into a binary; it is a
// copyable pattern for a real domain.
package example

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
	"github.com/elloloop/project-scaffold/packages/go/platform/principal"
)

// NormalizeTag canonicalizes a label: trimmed, lower-cased, with internal
// whitespace collapsed to single dashes.
func NormalizeTag(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), "-")
}

// ErrNoteNotFound is returned when a note does not exist in the caller's tenant.
var ErrNoteNotFound = errors.New("example: note not found")

// Note is a tiny domain entity.
type Note struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Author   string `json:"author"`
	Body     string `json:"body"`
}

// NoteRepo is the persistence seam. Tests wrap testkit.FakeStore; production
// code would provide a concrete adapter.
type NoteRepo interface {
	Save(ctx context.Context, n Note) error
	Get(ctx context.Context, tenantID, id string) (Note, error)
}

// NoteService is the business layer. Its dependencies are interfaces, so it is
// unit-testable with fakes and deterministic with fake clocks and ID generators.
type NoteService struct {
	repo NoteRepo
	ids  ports.IDGenerator
}

// NewNoteService wires the service.
func NewNoteService(repo NoteRepo, ids ports.IDGenerator) *NoteService {
	return &NoteService{repo: repo, ids: ids}
}

// Create stores a note authored by author in tenantID.
func (s *NoteService) Create(ctx context.Context, tenantID, author, body string) (Note, error) {
	n := Note{ID: s.ids.NewID(), TenantID: tenantID, Author: author, Body: body}
	if err := s.repo.Save(ctx, n); err != nil {
		return Note{}, err
	}
	return n, nil
}

// Get returns a note scoped to tenantID.
func (s *NoteService) Get(ctx context.Context, tenantID, id string) (Note, error) {
	return s.repo.Get(ctx, tenantID, id)
}

// NoteHandler adapts HTTP/JSON requests to the service.
type NoteHandler struct {
	svc *NoteService
}

// NewNoteHandler wires the handler.
func NewNoteHandler(svc *NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

// CreateNote handles POST {body} -> Note.
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	claims, ok := identity(r)
	if !ok {
		writeConnectError(w, http.StatusUnauthorized, "unauthenticated", "missing identity")
		return
	}
	var req struct {
		Body string `json:"body"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	note, err := h.svc.Create(r.Context(), claims.TenantID, claims.Subject, req.Body)
	if err != nil {
		writeConnectError(w, http.StatusInternalServerError, "internal", "could not create note")
		return
	}
	writeJSON(w, note)
}

// GetNote handles POST {id} -> Note, scoped to the caller's tenant.
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	claims, ok := identity(r)
	if !ok {
		writeConnectError(w, http.StatusUnauthorized, "unauthenticated", "missing identity")
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	note, err := h.svc.Get(r.Context(), claims.TenantID, req.ID)
	if errors.Is(err, ErrNoteNotFound) {
		writeConnectError(w, http.StatusNotFound, "not_found", "note not found")
		return
	}
	if err != nil {
		writeConnectError(w, http.StatusInternalServerError, "internal", "lookup failed")
		return
	}
	writeJSON(w, note)
}

func identity(r *http.Request) (ports.Claims, bool) {
	claims := principal.Read(r.Header)
	return claims, claims.Subject != "" && claims.TenantID != ""
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeConnectError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": msg})
}

// RegistrarClient calls an external availability API. It is constructed with an
// *http.Client so tests can inject an httptest server and timeout settings.
type RegistrarClient struct {
	baseURL string
	http    *http.Client
}

// NewRegistrarClient wires the client.
func NewRegistrarClient(baseURL string, c *http.Client) *RegistrarClient {
	return &RegistrarClient{baseURL: baseURL, http: c}
}

// CheckAvailability returns whether domain is available, surfacing errors for
// non-200 responses, malformed bodies, and transport failures.
func (rc *RegistrarClient) CheckAvailability(ctx context.Context, domain string) (bool, error) {
	u := rc.baseURL + "/availability?domain=" + url.QueryEscape(domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return false, err
	}
	resp, err := rc.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("registrar: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("registrar: unexpected status %d", resp.StatusCode)
	}
	var out struct {
		Available bool `json:"available"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, fmt.Errorf("registrar: decode: %w", err)
	}
	return out.Available, nil
}
