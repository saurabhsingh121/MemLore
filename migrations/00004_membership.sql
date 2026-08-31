-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    subject VARCHAR(256) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE teams (
    id VARCHAR(36) PRIMARY KEY,
    key VARCHAR(512) NOT NULL UNIQUE,
    name VARCHAR(512) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE projects (
    id VARCHAR(36) PRIMARY KEY,
    key VARCHAR(512) NOT NULL UNIQUE,
    name VARCHAR(512) NOT NULL,
    team_id VARCHAR(36) REFERENCES teams (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX ix_projects_team_id ON projects (team_id);

CREATE TABLE memberships (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    target_kind VARCHAR(32) NOT NULL,
    target_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, target_kind, target_id),
    CONSTRAINT memberships_target_kind_check CHECK (target_kind IN ('team', 'project'))
);

CREATE INDEX ix_memberships_user_id ON memberships (user_id);
CREATE INDEX ix_memberships_target ON memberships (target_kind, target_id);

CREATE TABLE scope_bindings (
    id VARCHAR(36) PRIMARY KEY,
    scope_kind VARCHAR(64) NOT NULL,
    scope_key VARCHAR(512) NOT NULL,
    project_id VARCHAR(36) NOT NULL REFERENCES projects (id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (scope_kind, scope_key),
    CONSTRAINT scope_bindings_kind_check CHECK (scope_kind IN ('repository', 'feature', 'task'))
);

CREATE INDEX ix_scope_bindings_project_id ON scope_bindings (project_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ix_scope_bindings_project_id;
DROP TABLE IF EXISTS scope_bindings;
DROP INDEX IF EXISTS ix_memberships_target;
DROP INDEX IF EXISTS ix_memberships_user_id;
DROP TABLE IF EXISTS memberships;
DROP INDEX IF EXISTS ix_projects_team_id;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
