-- +goose Up
-- +goose StatementBegin
CREATE TABLE pr_ingest_runs (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    actor_id VARCHAR(256) NOT NULL,
    pr_number INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL,
    prs_seen INTEGER NOT NULL DEFAULT 0,
    prs_skipped INTEGER NOT NULL DEFAULT 0,
    candidates_stored INTEGER NOT NULL DEFAULT 0,
    cursor_pr INTEGER,
    cursor_at TIMESTAMPTZ,
    error_summary TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ
);

CREATE INDEX ix_pr_ingest_runs_scope_started
    ON pr_ingest_runs (scope_kind, scope_key, started_at DESC);

CREATE UNIQUE INDEX ux_pr_ingest_runs_one_active
    ON pr_ingest_runs (scope_kind, scope_key)
    WHERE status = 'running';

CREATE TABLE pr_ingest_cursors (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    last_pr INTEGER NOT NULL,
    last_merged_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key)
);

CREATE TABLE pr_ingest_prs (
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    pr_number INTEGER NOT NULL,
    node_id VARCHAR(256),
    lore_entry_id VARCHAR(36),
    skipped BOOLEAN NOT NULL,
    skip_reason VARCHAR(64),
    processed_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope_kind, scope_key, pr_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_ingest_prs;
DROP TABLE IF EXISTS pr_ingest_cursors;
DROP INDEX IF EXISTS ux_pr_ingest_runs_one_active;
DROP INDEX IF EXISTS ix_pr_ingest_runs_scope_started;
DROP TABLE IF EXISTS pr_ingest_runs;
-- +goose StatementEnd
