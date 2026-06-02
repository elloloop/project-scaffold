package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/elloloop/project-scaffold/services/tasks/internal/auth"
	"github.com/elloloop/project-scaffold/services/tasks/internal/tasks"
)

type Config struct {
	Store          tasks.Store
	Auth           *auth.Service
	Logger         *slog.Logger
	AllowedOrigins []string
}

type Server struct {
	store          tasks.Store
	auth           *auth.Service
	log            *slog.Logger
	allowedOrigins map[string]bool
}

func New(cfg Config) http.Handler {
	s := &Server{
		store:          cfg.Store,
		auth:           cfg.Auth,
		log:            cfg.Logger,
		allowedOrigins: map[string]bool{},
	}
	if s.log == nil {
		s.log = slog.Default()
	}
	for _, origin := range cfg.AllowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			s.allowedOrigins[origin] = true
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("GET /api/session", s.withUser(s.session))
	mux.HandleFunc("GET /api/tasks", s.withUser(s.listTasks))
	mux.HandleFunc("POST /api/tasks", s.withUser(s.createTask))
	mux.HandleFunc("PATCH /api/tasks/{id}", s.withUser(s.updateTask))
	mux.HandleFunc("DELETE /api/tasks/{id}", s.withUser(s.deleteTask))
	return s.cors(mux)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if s.allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type userHandler func(http.ResponseWriter, *http.Request, auth.User)

func (s *Server) withUser(next userHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Bearer(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "missing bearer token")
			return
		}
		user, err := s.auth.Verify(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "invalid bearer token")
			return
		}
		next(w, r, user)
	}
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	session, err := s.auth.Login(req.Email, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}
	if err != nil {
		s.log.ErrorContext(r.Context(), "login_failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal", "login failed")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) session(w http.ResponseWriter, _ *http.Request, user auth.User) {
	writeJSON(w, http.StatusOK, map[string]auth.User{"user": user})
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request, user auth.User) {
	items, err := s.store.ListTasks(r.Context(), user.ID)
	if err != nil {
		s.log.ErrorContext(r.Context(), "list_tasks_failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal", "could not list tasks")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]tasks.Task{"items": items})
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request, user auth.User) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "title is required")
		return
	}
	task, err := s.store.CreateTask(r.Context(), user.ID, title)
	if err != nil {
		s.log.ErrorContext(r.Context(), "create_task_failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal", "could not create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, user auth.User) {
	var req struct {
		Title     *string `json:"title"`
		Completed *bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "title cannot be empty")
			return
		}
		req.Title = &trimmed
	}
	task, err := s.store.UpdateTask(r.Context(), user.ID, r.PathValue("id"), tasks.Patch{
		Title:     req.Title,
		Completed: req.Completed,
	})
	if errors.Is(err, tasks.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "task not found")
		return
	}
	if err != nil {
		s.log.ErrorContext(r.Context(), "update_task_failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal", "could not update task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request, user auth.User) {
	err := s.store.DeleteTask(r.Context(), user.ID, r.PathValue("id"))
	if errors.Is(err, tasks.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "task not found")
		return
	}
	if err != nil {
		s.log.ErrorContext(r.Context(), "delete_task_failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal", "could not delete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"code": code, "message": message})
}
