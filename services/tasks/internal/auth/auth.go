package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInvalidToken       = errors.New("auth: invalid token")
	ErrExpiredToken       = errors.New("auth: token expired")
)

type Config struct {
	Secret string
	TTL    time.Duration
	Now    func() time.Time
}

type Service struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	Role     string `json:"role"`
}

type Session struct {
	Token     string    `json:"token"`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

type claims struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	TenantID  string `json:"tenant"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

func NewService(cfg Config) *Service {
	if cfg.Secret == "" {
		cfg.Secret = "local-dev-secret-change-me"
	}
	if cfg.TTL <= 0 {
		cfg.TTL = time.Hour
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Service{secret: []byte(cfg.Secret), ttl: cfg.TTL, now: cfg.Now}
}

func (s *Service) Login(email, password string) (Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password != "demo" {
		return Session{}, ErrInvalidCredentials
	}
	user := User{
		ID:       "user:" + email,
		Email:    email,
		TenantID: "local",
		Role:     "member",
	}
	return s.Issue(user)
}

func (s *Service) Issue(user User) (Session, error) {
	expiresAt := s.now().Add(s.ttl).UTC()
	tokenClaims := claims{
		Subject:   user.ID,
		Email:     user.Email,
		TenantID:  user.TenantID,
		Role:      user.Role,
		ExpiresAt: expiresAt.Unix(),
	}
	token, err := s.sign(tokenClaims)
	if err != nil {
		return Session{}, err
	}
	return Session{Token: token, User: user, ExpiresAt: expiresAt}, nil
}

func (s *Service) Verify(token string) (User, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return User{}, ErrInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(s.signature(signingInput))) {
		return User{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return User{}, ErrInvalidToken
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return User{}, ErrInvalidToken
	}
	if c.ExpiresAt <= s.now().Unix() {
		return User{}, ErrExpiredToken
	}
	return User{ID: c.Subject, Email: c.Email, TenantID: c.TenantID, Role: c.Role}, nil
}

func (s *Service) sign(c claims) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	signingInput := encode(header) + "." + encode(payload)
	return signingInput + "." + s.signature(signingInput), nil
}

func (s *Service) signature(signingInput string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(signingInput))
	return encode(mac.Sum(nil))
}

func encode(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func Bearer(header string) (string, error) {
	if !strings.HasPrefix(header, "Bearer ") {
		return "", fmt.Errorf("%w: missing bearer token", ErrInvalidToken)
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" {
		return "", ErrInvalidToken
	}
	return token, nil
}
