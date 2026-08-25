package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// LoreRepository persists governance-plane lore entries.
type LoreRepository interface {
	Add(ctx context.Context, entry domain.LoreEntry) error
	Get(ctx context.Context, id string) (domain.LoreEntry, error)
	Save(ctx context.Context, entry domain.LoreEntry) error
	ListByScope(ctx context.Context, scope domain.Scope) ([]domain.LoreEntry, error)
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
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
