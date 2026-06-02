package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elloloop/project-scaffold/services/tasks/internal/auth"
	"github.com/elloloop/project-scaffold/services/tasks/internal/tasks"
)

type fakeStore struct {
	items []tasks.Task
}

func (s *fakeStore) CreateTask(_ context.Context, userID, title string) (tasks.Task, error) {
	task := tasks.Task{ID: "task_1", Title: title, CreatedBy: userID, CreatedAt: time.Unix(1, 0), UpdatedAt: time.Unix(1, 0)}
	s.items = append(s.items, task)
	return task, nil
}

func (s *fakeStore) ListTasks(_ context.Context, userID string) ([]tasks.Task, error) {
	var out []tasks.Task
	for _, item := range s.items {
		if item.CreatedBy == userID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *fakeStore) UpdateTask(_ context.Context, userID, id string, patch tasks.Patch) (tasks.Task, error) {
	for i, item := range s.items {
		if item.ID == id && item.CreatedBy == userID {
			if patch.Title != nil {
				item.Title = *patch.Title
			}
			if patch.Completed != nil {
				item.Completed = *patch.Completed
			}
			s.items[i] = item
			return item, nil
		}
	}
	return tasks.Task{}, tasks.ErrNotFound
}

func (s *fakeStore) DeleteTask(_ context.Context, userID, id string) error {
	for i, item := range s.items {
		if item.ID == id && item.CreatedBy == userID {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return nil
		}
	}
	return tasks.ErrNotFound
}

func TestAuthenticatedTaskFlow(t *testing.T) {
	store := &fakeStore{}
	authSvc := auth.NewService(auth.Config{Secret: "secret", Now: func() time.Time { return time.Unix(1000, 0) }})
	handler := New(Config{Store: store, Auth: authSvc, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})

	login := request(t, handler, http.MethodPost, "/api/auth/login", map[string]string{
		"email":    "demo@example.com",
		"password": "demo",
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", login.Code, login.Body.String())
	}
	var session auth.Session
	if err := json.Unmarshal(login.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}

	created := request(t, handler, http.MethodPost, "/api/tasks", map[string]string{"title": "Ship scaffold"}, session.Token)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", created.Code, created.Body.String())
	}

	list := request(t, handler, http.MethodGet, "/api/tasks", nil, session.Token)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", list.Code, list.Body.String())
	}
	if !bytes.Contains(list.Body.Bytes(), []byte("Ship scaffold")) {
		t.Fatalf("created task missing from list: %s", list.Body.String())
	}
}

func TestTasksRequireAuth(t *testing.T) {
	handler := New(Config{
		Store:  &fakeStore{},
		Auth:   auth.NewService(auth.Config{Secret: "secret"}),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	resp := request(t, handler, http.MethodGet, "/api/tasks", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.Code)
	}
}

func request(t *testing.T, h http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
