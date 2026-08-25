-- name: InsertOutboxEvent :exec
INSERT INTO outbox_events (
    id,
    aggregate_type,
    aggregate_id,
    event_type,
    payload,
    status,
    attempts,
    idempotency_key,
    created_at,
    processed_at,
    last_error
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: ClaimPendingOutboxEvents :many
SELECT
    id,
    aggregate_type,
    aggregate_id,
    event_type,
    payload,
    status,
    attempts,
    idempotency_key,
    created_at,
    processed_at,
    last_error
FROM outbox_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEventCompleted :exec
UPDATE outbox_events
SET status = 'completed',
    processed_at = $2,
    last_error = NULL
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET status = 'pending',
    attempts = $2,
    last_error = $3
WHERE id = $1;
