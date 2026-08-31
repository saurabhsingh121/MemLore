-- +goose Up
-- +goose StatementBegin
ALTER TABLE lore_entries
    ADD COLUMN superseded_by_id VARCHAR(36) NULL REFERENCES lore_entries (id),
    ADD COLUMN invalidated_by VARCHAR(256) NULL,
    ADD COLUMN invalidated_at TIMESTAMPTZ NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE lore_entries
    DROP COLUMN IF EXISTS superseded_by_id,
    DROP COLUMN IF EXISTS invalidated_by,
    DROP COLUMN IF EXISTS invalidated_at;
-- +goose StatementEnd
