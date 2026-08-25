package ports

import (
	"context"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// OutboxRepository persists outbox events within governance transactions.
type OutboxRepository interface {
	Add(ctx context.Context, event domain.OutboxEvent) error
}

// OutboxRunner claims and completes outbox events in transactional batches.
type OutboxRunner interface {
	ProcessBatch(
		ctx context.Context,
		limit int,
		now time.Time,
		handler func(context.Context, domain.OutboxEvent) error,
	) (int, error)
}
