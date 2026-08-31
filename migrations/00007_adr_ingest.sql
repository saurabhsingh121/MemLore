-- +goose Up
-- +goose StatementBegin
CREATE TABLE adr_ingest_runs (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    actor_id VARCHAR(256) NOT NULL,
    local_path TEXT NOT NULL,
    extra_dirs TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL,
    files_seen INTEGER NOT NULL DEFAULT 0,
    files_skipped INTEGER NOT NULL DEFAULT 0,
    lore_stored INTEGER NOT NULL DEFAULT 0,
    lore_superseded INTEGER NOT NULL DEFAULT 0,
    error_summary TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ
);

CREATE INDEX ix_adr_ingest_runs_scope_started
    ON adr_ingest_runs (scope_kind, scope_key, started_at DESC);

CREATE UNIQUE INDEX ux_adr_ingest_runs_one_active
    ON adr_ingest_runs (scope_kind, scope_key)
    WHERE status = 'running';

CREATE TABLE adr_ingest_cursors (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    last_path TEXT,
    last_checksum VARCHAR(64),
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key)
);

CREATE TABLE adr_ingest_files (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    relative_path TEXT NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    adr_id VARCHAR(256),
    lore_entry_id VARCHAR(36),
    skipped BOOLEAN NOT NULL,
    skip_reason VARCHAR(64),
    processed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key, relative_path, checksum)
);

CREATE INDEX ix_adr_ingest_files_path
    ON adr_ingest_files (scope_kind, scope_key, relative_path, processed_at DESC);

CREATE INDEX ix_adr_ingest_files_adr_id
    ON adr_ingest_files (scope_kind, scope_key, adr_id, processed_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS adr_ingest_files;
DROP TABLE IF EXISTS adr_ingest_cursors;
DROP INDEX IF EXISTS ux_adr_ingest_runs_one_active;
DROP INDEX IF EXISTS ix_adr_ingest_runs_scope_started;
DROP TABLE IF EXISTS adr_ingest_runs;
-- +goose StatementEnd
