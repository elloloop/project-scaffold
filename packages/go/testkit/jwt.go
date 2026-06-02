package testkit

import (
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elloloop/project-scaffold/packages/go/platform/ports"
)

// HMACSecret is the default HS256 signing secret used by TokenFactory. It is a
// test constant - never a production key.
const HMACSecret = "test-hs256-secret-do-not-use-in-prod"

// Token verification errors. Handlers/middleware must map each to a 401 and
// must never leak which one fired to the caller.
var (
	ErrMalformedToken    = errors.New("testkit: malformed token")
	ErrBadSignature      = errors.New("testkit: bad signature")
	ErrExpiredToken      = errors.New("testkit: token expired")
	ErrTokenNotYetValid  = errors.New("testkit: token not yet valid")
	ErrRevokedToken      = errors.New("testkit: token revoked")
	ErrUnsupportedAlg    = errors.New("testkit: unsupported alg")
	ErrMissingExpiration = errors.New("testkit: token missing exp")
)

// DefaultTokenTTL matches the production access-token lifetime (15 min).
const DefaultTokenTTL = 15 * time.Minute

// TokenFactory mints and verifies bearer tokens for tests. It supports HS256
// (the gateway's shared-secret development path) and RS256 (the JWKS path),
// can produce every invalid variant a handler must reject, and tracks
// revocations by jti. Expiry is driven by an injected clock, so token-lifetime
// tests are deterministic. The factory itself implements ports.TokenVerifier.
type TokenFactory struct {
	secret []byte
	rsaKey *rsa.PrivateKey
	clock  ports.Clock

	mu      sync.Mutex
	seq     int
	revoked map[string]bool
}

var _ ports.TokenVerifier = (*TokenFactory)(nil)

// NewTokenFactory returns a factory whose clock starts at FixedClockTime.
func NewTokenFactory() *TokenFactory {
	return NewTokenFactoryAt(NewFakeClock(FixedClockTime))
}

// NewTokenFactoryAt returns a factory driven by the given clock, letting a test
// share one clock between token expiry and the code under test.
func NewTokenFactoryAt(clock ports.Clock) *TokenFactory {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("testkit: rsa key generation failed: " + err.Error())
	}
	return &TokenFactory{
		secret:  []byte(HMACSecret),
		rsaKey:  key,
		clock:   clock,
		revoked: map[string]bool{},
	}
}

// RSAPublicKey exposes the RS256 public key so a test can stand up a JWKS-style
// verifier against this factory's tokens.
func (f *TokenFactory) RSAPublicKey() *rsa.PublicKey { return &f.rsaKey.PublicKey }

type mintOpts struct {
	alg  string // "HS256" | "RS256"
	exp  time.Time
	nbf  time.Time
	bunk bool // sign with a throwaway key to force a signature mismatch
}

// Valid mints a correctly-signed, unexpired HS256 token for claims.
func (f *TokenFactory) Valid(claims ports.Claims) string {
	tok, _ := f.mint(claims, mintOpts{alg: "HS256"})
	return tok
}

// ValidRS256 mints a correctly-signed, unexpired RS256 token (identity path).
func (f *TokenFactory) ValidRS256(claims ports.Claims) string {
	tok, _ := f.mint(claims, mintOpts{alg: "RS256"})
	return tok
}

// Expired mints a correctly-signed token whose exp is in the past.
func (f *TokenFactory) Expired(claims ports.Claims) string {
	now := f.clock.Now()
	tok, _ := f.mint(claims, mintOpts{alg: "HS256", exp: now.Add(-time.Minute), nbf: now.Add(-time.Hour)})
	return tok
}

// NotYetValid mints a token whose nbf is in the future.
func (f *TokenFactory) NotYetValid(claims ports.Claims) string {
	now := f.clock.Now()
	tok, _ := f.mint(claims, mintOpts{alg: "HS256", nbf: now.Add(time.Hour), exp: now.Add(2 * time.Hour)})
	return tok
}

// WrongSignature mints a well-formed token signed with the wrong key.
func (f *TokenFactory) WrongSignature(claims ports.Claims) string {
	tok, _ := f.mint(claims, mintOpts{alg: "HS256", bunk: true})
	return tok
}

// Revoked mints a valid token and records its jti as revoked, so Verify
// rejects it even though signature and expiry are fine.
func (f *TokenFactory) Revoked(claims ports.Claims) string {
	tok, jti := f.mint(claims, mintOpts{alg: "HS256"})
	f.mu.Lock()
	f.revoked[jti] = true
	f.mu.Unlock()
	return tok
}

// Malformed returns a syntactically invalid token (not three segments).
func (f *TokenFactory) Malformed() string { return "not.a-valid-jwt" }

// AlgNone returns a token with alg=none and an empty signature - the classic
// downgrade attack. Verify must reject it.
func (f *TokenFactory) AlgNone(claims ports.Claims) string {
	header := b64(`{"alg":"none","typ":"JWT"}`)
	payload := b64(string(f.payloadJSON(claims, f.clock.Now(), f.clock.Now().Add(DefaultTokenTTL), "jti-none")))
	return header + "." + payload + "."
}

func (f *TokenFactory) mint(claims ports.Claims, o mintOpts) (token, jti string) {
	now := f.clock.Now()
	f.mu.Lock()
	f.seq++
	jti = fmt.Sprintf("jti-%d", f.seq)
	f.mu.Unlock()

	exp := o.exp
	if exp.IsZero() {
		exp = now.Add(DefaultTokenTTL)
	}
	nbf := o.nbf
	if nbf.IsZero() {
		nbf = now
	}

	header := fmt.Sprintf(`{"alg":%q,"typ":"JWT"}`, o.alg)
	signingInput := b64(header) + "." + b64(string(f.payloadJSONNBF(claims, now, nbf, exp, jti)))

	var sig []byte
	switch o.alg {
	case "RS256":
		digest := sha256.Sum256([]byte(signingInput))
		key := f.rsaKey
		if o.bunk {
			key, _ = rsa.GenerateKey(rand.Reader, 2048)
		}
		sig, _ = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	default: // HS256
		secret := f.secret
		if o.bunk {
			secret = []byte("a-different-secret-entirely")
		}
		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(signingInput))
		sig = mac.Sum(nil)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), jti
}

func (f *TokenFactory) payloadJSON(c ports.Claims, iat, exp time.Time, jti string) []byte {
	return f.payloadJSONNBF(c, iat, iat, exp, jti)
}

func (f *TokenFactory) payloadJSONNBF(c ports.Claims, iat, nbf, exp time.Time, jti string) []byte {
	m := map[string]any{
		"sub":        c.Subject,
		"email":      c.Email,
		"name":       c.Name,
		"role":       c.Role,
		"tenant":     c.TenantID,
		"avatar_url": c.AvatarURL,
		"iat":        iat.Unix(),
		"nbf":        nbf.Unix(),
		"exp":        exp.Unix(),
		"jti":        jti,
	}
	b, _ := json.Marshal(m)
	return b
}

// Verify implements ports.TokenVerifier: it checks the signature (HS256 or
// RS256), enforces exp/nbf against the factory's clock, rejects revoked jtis,
// and refuses any other alg (including "none"). It returns a typed error for
// each failure so security tests can assert the exact rejection reason.
func (f *TokenFactory) Verify(_ context.Context, token string) (ports.Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ports.Claims{}, ErrMalformedToken
	}
	headerRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ports.Claims{}, ErrMalformedToken
	}
	var hdr struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerRaw, &hdr); err != nil {
		return ports.Claims{}, ErrMalformedToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ports.Claims{}, ErrMalformedToken
	}
	signingInput := parts[0] + "." + parts[1]

	switch hdr.Alg {
	case "HS256":
		mac := hmac.New(sha256.New, f.secret)
		mac.Write([]byte(signingInput))
		if !hmac.Equal(sig, mac.Sum(nil)) {
			return ports.Claims{}, ErrBadSignature
		}
	case "RS256":
		digest := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(&f.rsaKey.PublicKey, crypto.SHA256, digest[:], sig); err != nil {
			return ports.Claims{}, ErrBadSignature
		}
	default:
		return ports.Claims{}, ErrUnsupportedAlg
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ports.Claims{}, ErrMalformedToken
	}
	var p struct {
		Sub       string  `json:"sub"`
		Email     string  `json:"email"`
		Name      string  `json:"name"`
		Role      string  `json:"role"`
		Tenant    string  `json:"tenant"`
		AvatarURL string  `json:"avatar_url"`
		IAT       float64 `json:"iat"`
		NBF       float64 `json:"nbf"`
		EXP       float64 `json:"exp"`
		JTI       string  `json:"jti"`
	}
	if err := json.Unmarshal(payloadRaw, &p); err != nil {
		return ports.Claims{}, ErrMalformedToken
	}
	if p.EXP == 0 {
		return ports.Claims{}, ErrMissingExpiration
	}
	now := f.clock.Now()
	if now.After(time.Unix(int64(p.EXP), 0)) {
		return ports.Claims{}, ErrExpiredToken
	}
	if p.NBF != 0 && now.Before(time.Unix(int64(p.NBF), 0)) {
		return ports.Claims{}, ErrTokenNotYetValid
	}
	f.mu.Lock()
	revoked := f.revoked[p.JTI]
	f.mu.Unlock()
	if revoked {
		return ports.Claims{}, ErrRevokedToken
	}
	return ports.Claims{
		Subject:   p.Sub,
		Email:     p.Email,
		Name:      p.Name,
		Role:      p.Role,
		TenantID:  p.Tenant,
		AvatarURL: p.AvatarURL,
		IssuedAt:  time.Unix(int64(p.IAT), 0).UTC(),
		ExpiresAt: time.Unix(int64(p.EXP), 0).UTC(),
	}, nil
}

func b64(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
