-- name: InsertLoreReviewDecision :exec
INSERT INTO lore_review_decisions (
    id, scope_kind, scope_key, evidence_type, evidence_value, statement_checksum,
    lore_entry_id, successor_lore_id, status, actor_id, decided_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11
);

-- name: GetLoreReviewDecisionByIdentity :one
SELECT
    id, scope_kind, scope_key, evidence_type, evidence_value, statement_checksum,
    lore_entry_id, successor_lore_id, status, actor_id, decided_at
FROM lore_review_decisions
WHERE scope_kind = $1 AND scope_key = $2
  AND evidence_type = $3 AND evidence_value = $4 AND statement_checksum = $5;

-- name: GetLoreReviewDecisionByLoreID :one
SELECT
    id, scope_kind, scope_key, evidence_type, evidence_value, statement_checksum,
    lore_entry_id, successor_lore_id, status, actor_id, decided_at
FROM lore_review_decisions
WHERE lore_entry_id = $1
ORDER BY decided_at DESC
LIMIT 1;

-- name: ListLoreReviewDecisionsByScope :many
SELECT
    id, scope_kind, scope_key, evidence_type, evidence_value, statement_checksum,
    lore_entry_id, successor_lore_id, status, actor_id, decided_at
FROM lore_review_decisions
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY decided_at DESC;
