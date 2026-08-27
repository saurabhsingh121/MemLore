package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

// OutboxRepository implements outbox persistence via sqlc.
type OutboxRepository struct {
	q *sqlc.Queries
}

func NewOutboxRepository(q *sqlc.Queries) *OutboxRepository {
	return &OutboxRepository{q: q}
}

func (r *OutboxRepository) Add(ctx context.Context, event domain.OutboxEvent) error {
	params, err := outboxEventToInsertParams(event)
	if err != nil {
		return err
	}
	return r.q.InsertOutboxEvent(ctx, params)
}

// OutboxProcessor runs outbox work inside a single database transaction.
type OutboxProcessor struct {
	pool *pgxpool.Pool
}

func NewOutboxProcessor(pool *pgxpool.Pool) *OutboxProcessor {
	return &OutboxProcessor{pool: pool}
}

// ProcessBatch claims pending events and runs the handler for each within one transaction.
func (p *OutboxProcessor) ProcessBatch(
	ctx context.Context,
	limit int,
	now time.Time,
	handler func(context.Context, domain.OutboxEvent) error,
) (int, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	rows, err := q.ClaimPendingOutboxEvents(ctx, int32(limit))
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	processed := 0
	for _, row := range rows {
		event, convErr := outboxEventFromRow(row)
		if convErr != nil {
			return processed, convErr
		}
		if err := handler(ctx, event); err != nil {
			attempts := event.Attempts + 1
			if markErr := q.MarkOutboxEventFailed(ctx, sqlc.MarkOutboxEventFailedParams{
				ID:        event.ID,
				Attempts:  int32(attempts),
				LastError: textFromString(err.Error()),
			}); markErr != nil {
				return processed, markErr
			}
			continue
		}
		if err := q.MarkOutboxEventCompleted(ctx, sqlc.MarkOutboxEventCompletedParams{
			ID:          event.ID,
			ProcessedAt: timestamptzFromTime(now),
		}); err != nil {
			return processed, err
		}
		processed++
	}
	if err := tx.Commit(ctx); err != nil {
		return processed, err
	}
	return processed, nil
}

// Ensure interfaces.
var _ ports.OutboxRepository = (*OutboxRepository)(nil)
var _ ports.OutboxRunner = (*OutboxProcessor)(nil)
