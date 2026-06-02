package testkit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
	"github.com/elloloop/project-scaffold/packages/go/platform/principal"
)

// Trusted identity headers injected by the public gateway after it verifies the
// JWT. The internal service trusts these because it is not publicly
// reachable; handler tests set them directly via AsUser to simulate an
// authenticated caller without minting a token.
const (
	HeaderUserID   = principal.HeaderUserID
	HeaderUserMail = principal.HeaderEmail
	HeaderTenantID = principal.HeaderTenantID
)

// NewConnectRequest builds a Connect-over-HTTP/JSON POST to a procedure path
// such as "/task.v1.IssueService/CreateIssue" with body marshalled to JSON.
func NewConnectRequest(t testing.TB, procedure string, body any) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, procedure, bytes.NewReader(EncodeJSON(t, body)))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Authenticated adds an Authorization: Bearer header (the public-edge contract,
// for gateway/middleware tests that verify the token themselves).
func Authenticated(req *http.Request, token string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// AsUser sets the trusted identity headers the gateway injects, so a core
// handler test runs as the given user without a token.
func AsUser(req *http.Request, c ports.Claims) *http.Request {
	principal.Write(req.Header, c)
	return req
}

// CrossTenant sets the caller's identity headers but points the tenant header
// at a different tenant - the exact shape of a tenant-isolation negative test.
func CrossTenant(req *http.Request, c ports.Claims, otherTenantID string) *http.Request {
	AsUser(req, c)
	req.Header.Set(HeaderTenantID, otherTenantID)
	return req
}

// Serve runs req through handler and returns the recorded response.
func Serve(handler http.Handler, req *http.Request) *http.Response {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

// EncodeJSON marshals v to JSON, failing the test on error.
func EncodeJSON(t testing.TB, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("testkit: encode JSON: %v", err)
	}
	return b
}

// DecodeJSON unmarshals r into a fresh T, failing the test on error.
func DecodeJSON[T any](t testing.TB, r io.Reader) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		t.Fatalf("testkit: decode JSON: %v", err)
	}
	return v
}

// -- Assertions -------------------------------------------------------------

// AssertStatus fails the test unless resp.StatusCode == want.
func AssertStatus(t testing.TB, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Errorf("status = %d, want %d (body: %s)", resp.StatusCode, want, readBody(resp))
	}
}

// AssertJSONEq fails unless got (raw JSON) is structurally equal to wantJSON,
// ignoring key order and whitespace.
func AssertJSONEq(t testing.TB, got []byte, wantJSON string) {
	t.Helper()
	var a, b any
	if err := json.Unmarshal(got, &a); err != nil {
		t.Fatalf("testkit: got is not JSON: %v (%s)", err, got)
	}
	if err := json.Unmarshal([]byte(wantJSON), &b); err != nil {
		t.Fatalf("testkit: want is not JSON: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("JSON mismatch:\n got:  %s\n want: %s", got, wantJSON)
	}
}

// ConnectError is the JSON error envelope Connect returns for a failed RPC.
type ConnectError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AssertConnectError fails unless resp is an error with the given Connect code
// (e.g. "unauthenticated", "permission_denied", "not_found"). It returns the
// decoded envelope for further message assertions.
func AssertConnectError(t testing.TB, resp *http.Response, wantCode string) ConnectError {
	t.Helper()
	body := readBody(resp)
	var ce ConnectError
	if err := json.Unmarshal(body, &ce); err != nil {
		t.Fatalf("testkit: response is not a Connect error: %v (%s)", err, body)
	}
	if ce.Code != wantCode {
		t.Errorf("connect error code = %q, want %q (message: %q)", ce.Code, wantCode, ce.Message)
	}
	return ce
}

// AssertPage checks pagination metadata: total result count, page length, and
// whether another page follows.
func AssertPage(t testing.TB, p Page, wantTotal, wantLen int, wantHasNext bool) {
	t.Helper()
	if p.Total != wantTotal {
		t.Errorf("page total = %d, want %d", p.Total, wantTotal)
	}
	if len(p.Items) != wantLen {
		t.Errorf("page length = %d, want %d", len(p.Items), wantLen)
	}
	if (p.NextOffset >= 0) != wantHasNext {
		t.Errorf("page hasNext = %v (nextOffset=%d), want %v", p.NextOffset >= 0, p.NextOffset, wantHasNext)
	}
}

func readBody(resp *http.Response) []byte {
	if resp == nil || resp.Body == nil {
		return nil
	}
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return b
}

// -- External service double ------------------------------------------------

// RecordedRequest captures one inbound call to an ExternalService.
type RecordedRequest struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
}

// ExternalService is an httptest.Server that records every request and routes
// by exact path, for testing outbound API clients without real network. Build
// it with the canned responders below (RespondJSON, RespondStatus,
// RespondMalformed, RespondHang) to exercise success, error, malformed, and
// timeout paths.
type ExternalService struct {
	*httptest.Server
	mu       sync.Mutex
	requests []RecordedRequest
}

// NewExternalService starts a server that dispatches by exact request path;
// an unmatched path returns 404. It is closed automatically at test end.
func NewExternalService(t testing.TB, routes map[string]http.HandlerFunc) *ExternalService {
	t.Helper()
	es := &ExternalService{}
	es.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		es.mu.Lock()
		es.requests = append(es.requests, RecordedRequest{
			Method: r.Method, Path: r.URL.Path, Header: r.Header.Clone(), Body: body,
		})
		es.mu.Unlock()
		if h, ok := routes[r.URL.Path]; ok {
			r.Body = io.NopCloser(bytes.NewReader(body))
			h(w, r)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(es.Close)
	return es
}

// Requests returns a copy of all recorded requests.
func (es *ExternalService) Requests() []RecordedRequest {
	es.mu.Lock()
	defer es.mu.Unlock()
	return append([]RecordedRequest(nil), es.requests...)
}

// RespondJSON returns a handler that writes status and v as JSON.
func RespondJSON(status int, v any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
}

// RespondStatus returns a handler that writes only a status code.
func RespondStatus(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }
}

// RespondMalformed returns a handler that writes a 200 with a truncated/invalid
// JSON body, for decode-error tests.
func RespondMalformed() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"truncated":`))
	}
}

// RespondHang returns a handler that blocks until the request context is
// cancelled, so a client with a timeout exercises its timeout/retry path
// deterministically (no sleeps).
func RespondHang() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}
}

// -- No-network guard -------------------------------------------------------

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// ErrNetworkDisabled is returned by the transport installed by NoNetwork.
var ErrNetworkDisabled = errors.New("testkit: network access is disabled in this test")

// NoNetwork swaps http.DefaultTransport for one that refuses every call, so any
// code that reaches for the real network via the default client fails loudly.
// Explicit httptest servers keep working because they use their own client.
func NoNetwork(t testing.TB) {
	t.Helper()
	orig := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, ErrNetworkDisabled
	})
	t.Cleanup(func() { http.DefaultTransport = orig })
}
