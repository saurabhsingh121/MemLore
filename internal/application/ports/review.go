package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// ReviewDecisionRepository persists Accept/Reject overlays for extract identities.
type ReviewDecisionRepository interface {
	Insert(ctx context.Context, decision domain.ReviewDecision) error
	GetByIdentity(ctx context.Context, identity domain.ExtractIdentity) (domain.ReviewDecision, bool, error)
	GetByLoreID(ctx context.Context, loreEntryID string) (domain.ReviewDecision, bool, error)
	ListByScope(ctx context.Context, scope domain.Scope) ([]domain.ReviewDecision, error)
}
