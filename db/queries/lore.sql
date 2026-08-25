-- name: InsertLoreEntry :exec
INSERT INTO lore_entries (
    id, statement, scope_kind, scope_key, origin, verification_status,
    evidence, created_by, created_at, verified_by, verified_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);

-- name: GetLoreEntry :one
SELECT
    id, statement, scope_kind, scope_key, origin, verification_status,
    evidence, created_by, created_at, verified_by, verified_at, updated_at
FROM lore_entries
WHERE id = $1;

-- name: UpdateLoreEntry :exec
UPDATE lore_entries SET
    statement = $2,
    scope_kind = $3,
    scope_key = $4,
    origin = $5,
    verification_status = $6,
    evidence = $7,
    created_by = $8,
    created_at = $9,
    verified_by = $10,
    verified_at = $11,
    updated_at = $12
WHERE id = $1;

-- name: ListLoreEntriesByScope :many
SELECT
    id, statement, scope_kind, scope_key, origin, verification_status,
    evidence, created_by, created_at, verified_by, verified_at, updated_at
FROM lore_entries
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY created_at DESC;
