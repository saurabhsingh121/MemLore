-- name: InsertPRIngestRun :exec
INSERT INTO pr_ingest_runs (
    id, scope_kind, scope_key, actor_id, pr_number, status,
    prs_seen, prs_skipped, candidates_stored,
    cursor_pr, cursor_at, error_summary, started_at, finished_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14
);

-- name: UpdatePRIngestRun :exec
UPDATE pr_ingest_runs SET
    status = $2,
    prs_seen = $3,
    prs_skipped = $4,
    candidates_stored = $5,
    cursor_pr = $6,
    cursor_at = $7,
    error_summary = $8,
    finished_at = $9
WHERE id = $1;

-- name: GetPRIngestRun :one
SELECT
    id, scope_kind, scope_key, actor_id, pr_number, status,
    prs_seen, prs_skipped, candidates_stored,
    cursor_pr, cursor_at, error_summary, started_at, finished_at
FROM pr_ingest_runs
WHERE id = $1;

-- name: ListPRIngestRunsByScope :many
SELECT
    id, scope_kind, scope_key, actor_id, pr_number, status,
    prs_seen, prs_skipped, candidates_stored,
    cursor_pr, cursor_at, error_summary, started_at, finished_at
FROM pr_ingest_runs
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY started_at DESC;

-- name: GetPRIngestCursor :one
SELECT scope_kind, scope_key, last_pr, last_merged_at, updated_at
FROM pr_ingest_cursors
WHERE scope_kind = $1 AND scope_key = $2;

-- name: UpsertPRIngestCursor :exec
INSERT INTO pr_ingest_cursors (scope_kind, scope_key, last_pr, last_merged_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (scope_kind, scope_key) DO UPDATE SET
    last_pr = EXCLUDED.last_pr,
    last_merged_at = EXCLUDED.last_merged_at,
    updated_at = EXCLUDED.updated_at;

-- name: GetPRIngestPR :one
SELECT scope_kind, scope_key, pr_number, node_id, lore_entry_id, skipped, skip_reason, processed_at
FROM pr_ingest_prs
WHERE scope_kind = $1 AND scope_key = $2 AND pr_number = $3;

-- name: InsertPRIngestPR :exec
INSERT INTO pr_ingest_prs (
    scope_kind, scope_key, pr_number, node_id, lore_entry_id, skipped, skip_reason, processed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
