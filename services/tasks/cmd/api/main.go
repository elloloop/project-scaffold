package main

import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/elloloop/project-scaffold/packages/go/serverkit"
	"github.com/elloloop/project-scaffold/services/tasks/internal/auth"
	"github.com/elloloop/project-scaffold/services/tasks/internal/httpapi"
	"github.com/elloloop/project-scaffold/services/tasks/internal/tasks"
)

func main() {
	log := serverkit.Logger()
	db := openDB(log)
	defer func() { _ = db.Close() }()

	if err := tasks.Migrate(db); err != nil {
		log.Error("tasks_api_migrate_failed", "err", err)
		os.Exit(1)
	}

	store := tasks.NewPostgresStore(db)
	authService := auth.NewService(auth.Config{
		Secret: getenv("AUTH_HMAC_SECRET", "local-dev-secret-change-me"),
		TTL:    12 * time.Hour,
	})
	handler := httpapi.New(httpapi.Config{
		Store:          store,
		Auth:           authService,
		Logger:         log,
		AllowedOrigins: splitCSV(getenv("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")),
	})

	port := serverkit.EnvInt("TASK_API_PORT", serverkit.EnvInt("PORT", 8080))
	serverkit.Serve(log, "tasks_api", port, handler, serverkit.WithCleartextGRPC(false))
}

func openDB(log *slog.Logger) *sql.DB {
	db, err := sql.Open("pgx", getenv("DATABASE_URL", "postgres://app:app@localhost:5432/app?sslmode=disable"))
	if err != nil {
		log.Error("tasks_api_db_open_failed", "err", err)
		os.Exit(1)
	}
	for attempt := 1; attempt <= 30; attempt++ {
		if err := db.Ping(); err == nil {
			return db
		}
		log.Warn("tasks_api_db_wait", "attempt", attempt)
		time.Sleep(time.Second)
	}
	log.Error("tasks_api_db_unavailable")
	os.Exit(1)
	return nil
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ',' {
			part := trim(value[start:i])
			if part != "" {
				out = append(out, part)
			}
			start = i + 1
		}
	}
	return out
}

func trim(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t' || value[0] == '\n') {
		value = value[1:]
	}
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != ' ' && last != '\t' && last != '\n' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}
