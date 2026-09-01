package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// DecisionRepository persists human-recorded Decisions (not ADR projections).
type DecisionRepository interface {
	Add(ctx context.Context, decision domain.Decision) error
	Get(ctx context.Context, id string) (domain.Decision, error)
	Save(ctx context.Context, decision domain.Decision) error
	ListByScope(ctx context.Context, scope domain.Scope) ([]domain.Decision, error)
}
