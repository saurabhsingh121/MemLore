-- +goose Up
-- +goose StatementBegin
CREATE TABLE decisions (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    question TEXT NOT NULL,
    choice TEXT NOT NULL,
    rationale TEXT NOT NULL DEFAULT '',
    consequences TEXT NOT NULL DEFAULT '',
    owner TEXT NOT NULL,
    decided_at TIMESTAMPTZ NOT NULL,
    source_kind VARCHAR(32) NOT NULL,
    superseded_by_id VARCHAR(36),
    created_by VARCHAR(256) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_decisions_scope
    ON decisions (scope_kind, scope_key, decided_at DESC);

CREATE INDEX ix_decisions_superseded_by
    ON decisions (superseded_by_id);

CREATE TABLE decision_alternatives (
    decision_id VARCHAR(36) NOT NULL REFERENCES decisions (id) ON DELETE CASCADE,
    position INT NOT NULL,
    label TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (decision_id, position)
);

CREATE TABLE decision_components (
    decision_id VARCHAR(36) NOT NULL REFERENCES decisions (id) ON DELETE CASCADE,
    position INT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY (decision_id, position)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS decision_components;
DROP TABLE IF EXISTS decision_alternatives;
DROP TABLE IF EXISTS decisions;
-- +goose StatementEnd
