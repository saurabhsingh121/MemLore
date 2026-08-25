-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox_events (
    id VARCHAR(36) PRIMARY KEY,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    processed_at TIMESTAMPTZ,
    last_error TEXT
);

CREATE UNIQUE INDEX ux_outbox_events_idempotency_key ON outbox_events (idempotency_key);
CREATE INDEX ix_outbox_events_status_created ON outbox_events (status, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ix_outbox_events_status_created;
DROP INDEX IF EXISTS ux_outbox_events_idempotency_key;
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
