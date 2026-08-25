-- +goose Up
-- +goose StatementBegin
CREATE TABLE lore_entries (
    id VARCHAR(36) PRIMARY KEY,
    statement TEXT NOT NULL,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    origin VARCHAR(64) NOT NULL,
    verification_status VARCHAR(32) NOT NULL,
    evidence JSONB NOT NULL,
    created_by VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    verified_by VARCHAR(256),
    verified_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_lore_entries_scope_created
    ON lore_entries (scope_kind, scope_key, created_at);

CREATE TABLE audit_records (
    id VARCHAR(36) PRIMARY KEY,
    target_id VARCHAR(36) NOT NULL,
    action VARCHAR(32) NOT NULL,
    actor_id VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_audit_records_target_id ON audit_records (target_id);

CREATE INDEX ix_audit_records_target_created
    ON audit_records (target_id, created_at, id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ix_audit_records_target_created;
DROP INDEX IF EXISTS ix_audit_records_target_id;
DROP TABLE IF EXISTS audit_records;
DROP INDEX IF EXISTS ix_lore_entries_scope_created;
DROP TABLE IF EXISTS lore_entries;
-- +goose StatementEnd
