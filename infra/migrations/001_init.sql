CREATE TABLE IF NOT EXISTS scaffold_healthcheck (
  id BIGSERIAL PRIMARY KEY,
  service_name TEXT NOT NULL,
  checked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

