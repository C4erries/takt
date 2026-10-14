-- 000001_init_schema.up.sql
-- Base schema placeholder for Sprint 0.1
CREATE TABLE IF NOT EXISTS schema_version_check (
    id SERIAL PRIMARY KEY,
    initialized_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

