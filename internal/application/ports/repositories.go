package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// SearchRelevantOpts selects lore by statement relevance.
type SearchRelevantOpts struct {
	Query string
	Scope *domain.Scope // nil searches all scopes
	Limit int           // max candidates to return (after match); <=0 means no cap
}

// LoreRepository persists governance-plane lore entries.
type LoreRepository interface {
	Add(ctx context.Context, entry domain.LoreEntry) error
	Get(ctx context.Context, id string) (domain.LoreEntry, error)
	Save(ctx context.Context, entry domain.LoreEntry) error
	ListByScope(ctx context.Context, scope domain.Scope) ([]domain.LoreEntry, error)
	SearchRelevant(ctx context.Context, opts SearchRelevantOpts) ([]domain.LoreEntry, error)
}

// AuditRepository persists append-only audit records.
type AuditRepository interface {
	Add(ctx context.Context, record domain.AuditRecord) error
	ListByTarget(ctx context.Context, targetID string) ([]domain.AuditRecord, error)
}

// UnitOfWork coordinates repositories within a transaction.
type UnitOfWork interface {
	LoreEntries() LoreRepository
	Audits() AuditRepository
	Outbox() OutboxRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
