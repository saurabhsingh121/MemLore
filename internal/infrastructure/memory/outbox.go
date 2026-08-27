package memory

import (
	"context"
	"sync"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// OutboxRepository is an in-memory outbox store for tests.
type OutboxRepository struct {
	mu    sync.RWMutex
	items []domain.OutboxEvent
}

func NewOutboxRepository() *OutboxRepository {
	return &OutboxRepository{items: make([]domain.OutboxEvent, 0)}
}

func (r *OutboxRepository) Add(_ context.Context, event domain.OutboxEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, event)
	return nil
}

func (r *OutboxRepository) ListPending() []domain.OutboxEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.OutboxEvent, 0)
	for _, item := range r.items {
		if item.Status == domain.OutboxStatusPending {
			out = append(out, item)
		}
	}
	return out
}

func (r *OutboxRepository) MarkCompleted(id string, processedAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.items {
		if item.ID == id {
			item.Status = domain.OutboxStatusCompleted
			item.ProcessedAt = &processedAt
			item.LastError = ""
			r.items[i] = item
			return
		}
	}
}

func (r *OutboxRepository) MarkFailed(id string, attempts int, lastError string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, item := range r.items {
		if item.ID == id {
			item.Attempts = attempts
			item.LastError = lastError
			r.items[i] = item
			return
		}
	}
}

// OutboxRunner processes in-memory pending events for worker tests.
type OutboxRunner struct {
	repo *OutboxRepository
}

func NewOutboxRunner(repo *OutboxRepository) *OutboxRunner {
	return &OutboxRunner{repo: repo}
}

func (r *OutboxRunner) ProcessBatch(
	_ context.Context,
	limit int,
	now time.Time,
	handler func(context.Context, domain.OutboxEvent) error,
) (int, error) {
	pending := r.repo.ListPending()
	if len(pending) == 0 {
		return 0, nil
	}
	if limit > len(pending) {
		limit = len(pending)
	}
	processed := 0
	for _, event := range pending[:limit] {
		if err := handler(context.Background(), event); err != nil {
			r.repo.MarkFailed(event.ID, event.Attempts+1, err.Error())
			continue
		}
		r.repo.MarkCompleted(event.ID, now)
		processed++
	}
	return processed, nil
}

var _ ports.OutboxRunner = (*OutboxRunner)(nil)
