-- +goose Up
-- +goose StatementBegin
CREATE TABLE git_ingest_runs (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    actor_id VARCHAR(256) NOT NULL,
    local_path TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    commits_seen INTEGER NOT NULL DEFAULT 0,
    commits_skipped INTEGER NOT NULL DEFAULT 0,
    candidates_stored INTEGER NOT NULL DEFAULT 0,
    cursor_sha VARCHAR(64),
    cursor_at TIMESTAMPTZ,
    error_summary TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ
);

CREATE INDEX ix_git_ingest_runs_scope_started
    ON git_ingest_runs (scope_kind, scope_key, started_at DESC);

CREATE UNIQUE INDEX ux_git_ingest_runs_one_active
    ON git_ingest_runs (scope_kind, scope_key)
    WHERE status = 'running';

CREATE TABLE git_ingest_cursors (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    last_sha VARCHAR(64) NOT NULL,
    last_committed_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key)
);

CREATE TABLE git_ingest_shas (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    sha VARCHAR(64) NOT NULL,
    lore_entry_id VARCHAR(36),
    skipped BOOLEAN NOT NULL,
    skip_reason VARCHAR(64),
    processed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key, sha)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS git_ingest_shas;
DROP TABLE IF EXISTS git_ingest_cursors;
DROP INDEX IF EXISTS ux_git_ingest_runs_one_active;
DROP INDEX IF EXISTS ix_git_ingest_runs_scope_started;
DROP TABLE IF EXISTS git_ingest_runs;
-- +goose StatementEnd
