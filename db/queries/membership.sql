-- name: EnsureUser :one
INSERT INTO users (id, subject, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (subject) DO UPDATE SET subject = EXCLUDED.subject
RETURNING id, subject, created_at;

-- name: GetUserBySubject :one
SELECT id, subject, created_at FROM users WHERE subject = $1;

-- name: InsertTeam :exec
INSERT INTO teams (id, key, name, created_at) VALUES ($1, $2, $3, $4);

-- name: GetTeamByKey :one
SELECT id, key, name, created_at FROM teams WHERE key = $1;

-- name: InsertProject :exec
INSERT INTO projects (id, key, name, team_id, created_at) VALUES ($1, $2, $3, $4, $5);

-- name: GetProjectByKey :one
SELECT id, key, name, team_id, created_at FROM projects WHERE key = $1;

-- name: GetProjectByID :one
SELECT id, key, name, team_id, created_at FROM projects WHERE id = $1;

-- name: InsertMembership :exec
INSERT INTO memberships (id, user_id, target_kind, target_id, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteMembership :exec
DELETE FROM memberships
WHERE user_id = $1 AND target_kind = $2 AND target_id = $3;

-- name: HasMembership :one
SELECT EXISTS(
    SELECT 1 FROM memberships
    WHERE user_id = $1 AND target_kind = $2 AND target_id = $3
) AS ok;

-- name: InsertScopeBinding :exec
INSERT INTO scope_bindings (id, scope_kind, scope_key, project_id, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteScopeBinding :exec
DELETE FROM scope_bindings WHERE scope_kind = $1 AND scope_key = $2;

-- name: GetScopeBinding :one
SELECT id, scope_kind, scope_key, project_id, created_at
FROM scope_bindings
WHERE scope_kind = $1 AND scope_key = $2;

-- name: GetTeamKeyByID :one
SELECT key FROM teams WHERE id = $1;
