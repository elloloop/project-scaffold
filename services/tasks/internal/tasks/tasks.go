package tasks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("tasks: not found")

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Patch struct {
	Title     *string
	Completed *bool
}

type Job struct {
	ID       int64
	Kind     string
	TaskID   string
	Payload  []byte
	Attempts int
}

type Store interface {
	CreateTask(ctx context.Context, userID, title string) (Task, error)
	ListTasks(ctx context.Context, userID string) ([]Task, error)
	UpdateTask(ctx context.Context, userID, id string, patch Patch) (Task, error)
	DeleteTask(ctx context.Context, userID, id string) error
}

type JobStore interface {
	ClaimJob(ctx context.Context) (Job, bool, error)
	MarkJobDone(ctx context.Context, id int64) error
	MarkJobFailed(ctx context.Context, id int64, cause error) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func Migrate(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(720001001)`); err != nil {
		return err
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(720001001)`) }()

	_, err = conn.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS tasks (
  id text PRIMARY KEY,
  title text NOT NULL,
  completed boolean NOT NULL DEFAULT false,
  created_by text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS background_jobs (
  id bigserial PRIMARY KEY,
  kind text NOT NULL,
  task_id text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  status text NOT NULL DEFAULT 'queued',
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tasks_created_by_updated_at_idx ON tasks (created_by, updated_at DESC);
CREATE INDEX IF NOT EXISTS background_jobs_status_created_at_idx ON background_jobs (status, created_at);
`)
	return err
}

func (s *PostgresStore) CreateTask(ctx context.Context, userID, title string) (Task, error) {
	id := newID("task")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var task Task
	err = tx.QueryRowContext(ctx, `
INSERT INTO tasks (id, title, created_by)
VALUES ($1, $2, $3)
RETURNING id, title, completed, created_by, created_at, updated_at
`, id, title, userID).Scan(&task.ID, &task.Title, &task.Completed, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return Task{}, err
	}

	payload, _ := json.Marshal(map[string]string{"task_id": task.ID, "created_by": userID})
	if _, err := tx.ExecContext(ctx, `
INSERT INTO background_jobs (kind, task_id, payload)
VALUES ($1, $2, $3::jsonb)
`, "task.created", task.ID, string(payload)); err != nil {
		return Task{}, err
	}

	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (s *PostgresStore) ListTasks(ctx context.Context, userID string) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, title, completed, created_by, created_at, updated_at
FROM tasks
WHERE created_by = $1
ORDER BY updated_at DESC, created_at DESC
`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Completed, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, task)
	}
	return out, rows.Err()
}

func (s *PostgresStore) UpdateTask(ctx context.Context, userID, id string, patch Patch) (Task, error) {
	var title sql.NullString
	if patch.Title != nil {
		title = sql.NullString{String: *patch.Title, Valid: true}
	}
	var completed sql.NullBool
	if patch.Completed != nil {
		completed = sql.NullBool{Bool: *patch.Completed, Valid: true}
	}

	var task Task
	err := s.db.QueryRowContext(ctx, `
UPDATE tasks
SET title = COALESCE($3, title),
    completed = COALESCE($4, completed),
    updated_at = now()
WHERE id = $1 AND created_by = $2
RETURNING id, title, completed, created_by, created_at, updated_at
`, id, userID, title, completed).Scan(&task.ID, &task.Title, &task.Completed, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return task, err
}

func (s *PostgresStore) DeleteTask(ctx context.Context, userID, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1 AND created_by = $2`, id, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ClaimJob(ctx context.Context) (Job, bool, error) {
	var job Job
	err := s.db.QueryRowContext(ctx, `
WITH next_job AS (
  SELECT id
  FROM background_jobs
  WHERE status = 'queued'
  ORDER BY created_at
  LIMIT 1
  FOR UPDATE SKIP LOCKED
)
UPDATE background_jobs
SET status = 'running',
    attempts = attempts + 1,
    updated_at = now()
WHERE id IN (SELECT id FROM next_job)
RETURNING id, kind, task_id, payload, attempts
`).Scan(&job.ID, &job.Kind, &job.TaskID, &job.Payload, &job.Attempts)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, err
	}
	return job, true, nil
}

func (s *PostgresStore) MarkJobDone(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE background_jobs
SET status = 'done', updated_at = now(), last_error = NULL
WHERE id = $1
`, id)
	return err
}

func (s *PostgresStore) MarkJobFailed(ctx context.Context, id int64, cause error) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE background_jobs
SET status = 'queued', updated_at = now(), last_error = $2
WHERE id = $1
`, id, cause.Error())
	return err
}

func newID(prefix string) string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}
