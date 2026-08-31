-- +goose Up
-- +goose StatementBegin
CREATE TABLE lore_review_decisions (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    evidence_type VARCHAR(32) NOT NULL,
    evidence_value VARCHAR(2048) NOT NULL,
    statement_checksum VARCHAR(64) NOT NULL,
    lore_entry_id VARCHAR(36) NOT NULL,
    successor_lore_id VARCHAR(36),
    status VARCHAR(32) NOT NULL,
    actor_id VARCHAR(256) NOT NULL,
    decided_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX ux_lore_review_decisions_identity
    ON lore_review_decisions (scope_kind, scope_key, evidence_type, evidence_value, statement_checksum);

CREATE INDEX ix_lore_review_decisions_scope
    ON lore_review_decisions (scope_kind, scope_key, decided_at DESC);

CREATE INDEX ix_lore_review_decisions_lore_entry
    ON lore_review_decisions (lore_entry_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lore_review_decisions;
-- +goose StatementEnd
