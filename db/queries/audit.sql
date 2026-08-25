-- name: InsertAuditRecord :exec
INSERT INTO audit_records (id, target_id, action, actor_id, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditRecordsByTarget :many
SELECT id, target_id, action, actor_id, created_at
FROM audit_records
WHERE target_id = $1
ORDER BY created_at ASC, id ASC;
