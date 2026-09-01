-- name: InsertDecision :exec
INSERT INTO decisions (
    id, scope_kind, scope_key, question, choice, rationale, consequences,
    owner, decided_at, source_kind, superseded_by_id, created_by, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13, $14
);

-- name: GetDecision :one
SELECT
    id, scope_kind, scope_key, question, choice, rationale, consequences,
    owner, decided_at, source_kind, superseded_by_id, created_by, created_at, updated_at
FROM decisions
WHERE id = $1;

-- name: UpdateDecision :exec
UPDATE decisions SET
    question = $2,
    choice = $3,
    rationale = $4,
    consequences = $5,
    owner = $6,
    decided_at = $7,
    source_kind = $8,
    superseded_by_id = $9,
    updated_at = $10
WHERE id = $1;

-- name: ListDecisionsByScope :many
SELECT
    id, scope_kind, scope_key, question, choice, rationale, consequences,
    owner, decided_at, source_kind, superseded_by_id, created_by, created_at, updated_at
FROM decisions
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY decided_at DESC;

-- name: InsertDecisionAlternative :exec
INSERT INTO decision_alternatives (decision_id, position, label, note)
VALUES ($1, $2, $3, $4);

-- name: DeleteDecisionAlternatives :exec
DELETE FROM decision_alternatives WHERE decision_id = $1;

-- name: ListDecisionAlternatives :many
SELECT decision_id, position, label, note
FROM decision_alternatives
WHERE decision_id = $1
ORDER BY position ASC;

-- name: InsertDecisionComponent :exec
INSERT INTO decision_components (decision_id, position, name)
VALUES ($1, $2, $3);

-- name: DeleteDecisionComponents :exec
DELETE FROM decision_components WHERE decision_id = $1;

-- name: ListDecisionComponents :many
SELECT decision_id, position, name
FROM decision_components
WHERE decision_id = $1
ORDER BY position ASC;
