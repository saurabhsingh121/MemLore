package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// LoreRepository is an in-memory lore store for tests.
type LoreRepository struct {
	mu    sync.RWMutex
	items map[string]domain.LoreEntry
}

func NewLoreRepository() *LoreRepository {
	return &LoreRepository{items: make(map[string]domain.LoreEntry)}
}

func (r *LoreRepository) Add(_ context.Context, entry domain.LoreEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[entry.ID] = entry
	return nil
}

func (r *LoreRepository) Get(_ context.Context, id string) (domain.LoreEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.items[id]
	if !ok {
		return domain.LoreEntry{}, &domain.NotFoundError{
			Message: fmt.Sprintf("lore entry %s not found", id),
		}
	}
	return entry, nil
}

func (r *LoreRepository) Save(_ context.Context, entry domain.LoreEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[entry.ID] = entry
	return nil
}

// Len returns the number of stored lore entries (test helper).
func (r *LoreRepository) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}

func (r *LoreRepository) ListByScope(_ context.Context, scope domain.Scope) ([]domain.LoreEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.LoreEntry, 0)
	for _, entry := range r.items {
		if entry.Scope.Kind == scope.Kind && entry.Scope.Key == scope.Key {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *LoreRepository) SearchRelevant(_ context.Context, opts ports.SearchRelevantOpts) ([]domain.LoreEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.LoreEntry, 0)
	for _, entry := range r.items {
		if opts.Scope != nil {
			if entry.Scope.Kind != opts.Scope.Kind || entry.Scope.Key != opts.Scope.Key {
				continue
			}
		}
		if !domain.StatementMatchesQuery(entry.Statement, opts.Query) {
			continue
		}
		out = append(out, entry)
	}
	domain.SortLoreByRelevance(out, opts.Query)
	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[:opts.Limit]
	}
	return out, nil
}

// AuditRepository is an in-memory audit store for tests.
type AuditRepository struct {
	mu    sync.RWMutex
	items []domain.AuditRecord
}

func NewAuditRepository() *AuditRepository {
	return &AuditRepository{items: make([]domain.AuditRecord, 0)}
}

func (r *AuditRepository) Add(_ context.Context, record domain.AuditRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, record)
	return nil
}

func (r *AuditRepository) ListByTarget(_ context.Context, targetID string) ([]domain.AuditRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.AuditRecord, 0)
	for _, record := range r.items {
		if record.TargetID == targetID {
			out = append(out, record)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

// UnitOfWork shares repositories across HTTP requests in contract tests.
type UnitOfWork struct {
	lore      *LoreRepository
	audits    *AuditRepository
	outbox    *OutboxRepository
	ingest    *IngestRepository
	prIngest  *PRIngestRepository
	adrIngest *ADRIngestRepository
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{
		lore:      NewLoreRepository(),
		audits:    NewAuditRepository(),
		outbox:    NewOutboxRepository(),
		ingest:    NewIngestRepository(),
		prIngest:  NewPRIngestRepository(),
		adrIngest: NewADRIngestRepository(),
	}
}

func (u *UnitOfWork) LoreEntries() ports.LoreRepository { return u.lore }

func (u *UnitOfWork) Audits() ports.AuditRepository { return u.audits }

func (u *UnitOfWork) Outbox() ports.OutboxRepository { return u.outbox }

func (u *UnitOfWork) Ingest() ports.IngestRepository { return u.ingest }

func (u *UnitOfWork) PRIngest() ports.PRIngestRepository { return u.prIngest }

func (u *UnitOfWork) ADRIngest() ports.ADRIngestRepository { return u.adrIngest }

func (u *UnitOfWork) Commit(context.Context) error   { return nil }
func (u *UnitOfWork) Rollback(context.Context) error { return nil }

// BeginFactory returns a factory that always yields the shared unit of work.
func BeginFactory(uow *UnitOfWork) ports.UnitOfWorkFactory {
	return func(context.Context) (ports.UnitOfWork, error) {
		return uow, nil
	}
}
