package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/elloloop/project-scaffold/packages/go/serverkit"
	"github.com/elloloop/project-scaffold/services/tasks/internal/tasks"
)

func main() {
	log := serverkit.Logger()
	db := openDB(log)
	defer func() { _ = db.Close() }()

	if err := tasks.Migrate(db); err != nil {
		log.Error("tasks_worker_migrate_failed", "err", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	store := tasks.NewPostgresStore(db)
	log.Info("tasks_worker_started")
	for {
		select {
		case <-ctx.Done():
			log.Info("tasks_worker_shutdown")
			return
		default:
		}

		job, ok, err := store.ClaimJob(ctx)
		if err != nil {
			log.Error("tasks_worker_claim_failed", "err", err)
			sleep(ctx, 2*time.Second)
			continue
		}
		if !ok {
			sleep(ctx, 2*time.Second)
			continue
		}

		log.Info("tasks_worker_job_claimed", "job_id", job.ID, "kind", job.Kind, "task_id", job.TaskID, "attempts", job.Attempts)
		if err := store.MarkJobDone(ctx, job.ID); err != nil {
			log.Error("tasks_worker_mark_done_failed", "job_id", job.ID, "err", err)
			continue
		}
		log.Info("tasks_worker_job_done", "job_id", job.ID)
	}
}

func openDB(log *slog.Logger) *sql.DB {
	db, err := sql.Open("pgx", getenv("DATABASE_URL", "postgres://app:app@localhost:5432/app?sslmode=disable"))
	if err != nil {
		log.Error("tasks_worker_db_open_failed", "err", err)
		os.Exit(1)
	}
	for attempt := 1; attempt <= 30; attempt++ {
		if err := db.Ping(); err == nil {
			return db
		}
		log.Warn("tasks_worker_db_wait", "attempt", attempt)
		time.Sleep(time.Second)
	}
	log.Error("tasks_worker_db_unavailable")
	os.Exit(1)
	return nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
