-- name: InsertGitIngestRun :exec
INSERT INTO git_ingest_runs (
    id, scope_kind, scope_key, actor_id, local_path, status,
    commits_seen, commits_skipped, candidates_stored,
    cursor_sha, cursor_at, error_summary, started_at, finished_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14
);

-- name: UpdateGitIngestRun :exec
UPDATE git_ingest_runs SET
    status = $2,
    commits_seen = $3,
    commits_skipped = $4,
    candidates_stored = $5,
    cursor_sha = $6,
    cursor_at = $7,
    error_summary = $8,
    finished_at = $9
WHERE id = $1;

-- name: GetGitIngestRun :one
SELECT
    id, scope_kind, scope_key, actor_id, local_path, status,
    commits_seen, commits_skipped, candidates_stored,
    cursor_sha, cursor_at, error_summary, started_at, finished_at
FROM git_ingest_runs
WHERE id = $1;

-- name: ListGitIngestRunsByScope :many
SELECT
    id, scope_kind, scope_key, actor_id, local_path, status,
    commits_seen, commits_skipped, candidates_stored,
    cursor_sha, cursor_at, error_summary, started_at, finished_at
FROM git_ingest_runs
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY started_at DESC;

-- name: GetGitIngestCursor :one
SELECT scope_kind, scope_key, last_sha, last_committed_at, updated_at
FROM git_ingest_cursors
WHERE scope_kind = $1 AND scope_key = $2;

-- name: UpsertGitIngestCursor :exec
INSERT INTO git_ingest_cursors (scope_kind, scope_key, last_sha, last_committed_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (scope_kind, scope_key) DO UPDATE SET
    last_sha = EXCLUDED.last_sha,
    last_committed_at = EXCLUDED.last_committed_at,
    updated_at = EXCLUDED.updated_at;

-- name: GetGitIngestSHA :one
SELECT scope_kind, scope_key, sha, lore_entry_id, skipped, skip_reason, processed_at
FROM git_ingest_shas
WHERE scope_kind = $1 AND scope_key = $2 AND sha = $3;

-- name: InsertGitIngestSHA :exec
INSERT INTO git_ingest_shas (
    scope_kind, scope_key, sha, lore_entry_id, skipped, skip_reason, processed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7);
