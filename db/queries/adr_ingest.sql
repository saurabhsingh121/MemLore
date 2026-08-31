-- name: InsertADRIngestRun :exec
INSERT INTO adr_ingest_runs (
    id, scope_kind, scope_key, actor_id, local_path, extra_dirs, status,
    files_seen, files_skipped, lore_stored, lore_superseded,
    error_summary, started_at, finished_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14
);

-- name: UpdateADRIngestRun :exec
UPDATE adr_ingest_runs SET
    status = $2,
    files_seen = $3,
    files_skipped = $4,
    lore_stored = $5,
    lore_superseded = $6,
    error_summary = $7,
    finished_at = $8
WHERE id = $1;

-- name: GetADRIngestRun :one
SELECT
    id, scope_kind, scope_key, actor_id, local_path, extra_dirs, status,
    files_seen, files_skipped, lore_stored, lore_superseded,
    error_summary, started_at, finished_at
FROM adr_ingest_runs
WHERE id = $1;

-- name: ListADRIngestRunsByScope :many
SELECT
    id, scope_kind, scope_key, actor_id, local_path, extra_dirs, status,
    files_seen, files_skipped, lore_stored, lore_superseded,
    error_summary, started_at, finished_at
FROM adr_ingest_runs
WHERE scope_kind = $1 AND scope_key = $2
ORDER BY started_at DESC;

-- name: GetADRIngestCursor :one
SELECT scope_kind, scope_key, last_path, last_checksum, updated_at
FROM adr_ingest_cursors
WHERE scope_kind = $1 AND scope_key = $2;

-- name: UpsertADRIngestCursor :exec
INSERT INTO adr_ingest_cursors (scope_kind, scope_key, last_path, last_checksum, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (scope_kind, scope_key) DO UPDATE SET
    last_path = EXCLUDED.last_path,
    last_checksum = EXCLUDED.last_checksum,
    updated_at = EXCLUDED.updated_at;

-- name: GetADRIngestFile :one
SELECT scope_kind, scope_key, relative_path, checksum, adr_id, lore_entry_id, skipped, skip_reason, processed_at
FROM adr_ingest_files
WHERE scope_kind = $1 AND scope_key = $2 AND relative_path = $3 AND checksum = $4;

-- name: InsertADRIngestFile :exec
INSERT INTO adr_ingest_files (
    scope_kind, scope_key, relative_path, checksum, adr_id, lore_entry_id, skipped, skip_reason, processed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: LatestStoredADRByPath :one
SELECT scope_kind, scope_key, relative_path, checksum, adr_id, lore_entry_id, skipped, skip_reason, processed_at
FROM adr_ingest_files
WHERE scope_kind = $1 AND scope_key = $2 AND relative_path = $3
  AND skipped = FALSE AND lore_entry_id IS NOT NULL AND lore_entry_id <> ''
ORDER BY processed_at DESC
LIMIT 1;

-- name: LatestStoredADRByID :one
SELECT scope_kind, scope_key, relative_path, checksum, adr_id, lore_entry_id, skipped, skip_reason, processed_at
FROM adr_ingest_files
WHERE scope_kind = $1 AND scope_key = $2
  AND skipped = FALSE AND lore_entry_id IS NOT NULL AND lore_entry_id <> ''
  AND adr_id IS NOT NULL
  AND (
    adr_id = sqlc.arg(adr_id)
    OR adr_id LIKE sqlc.arg(adr_id) || '-%'
  )
ORDER BY processed_at DESC
LIMIT 1;
