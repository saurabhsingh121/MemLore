package memory

import (
	"context"
	"sync"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.ReviewDecisionRepository = (*ReviewDecisionRepository)(nil)

// ReviewDecisionRepository is an in-memory review-decision store for tests.
type ReviewDecisionRepository struct {
	mu    sync.RWMutex
	items map[string]domain.ReviewDecision
}

func NewReviewDecisionRepository() *ReviewDecisionRepository {
	return &ReviewDecisionRepository{items: make(map[string]domain.ReviewDecision)}
}

func reviewIdentityKey(d domain.ReviewDecision) string {
	return domain.ExtractIdentity{
		Scope:             d.Scope,
		EvidenceType:      d.EvidenceType,
		EvidenceValue:     d.EvidenceValue,
		StatementChecksum: d.StatementChecksum,
	}.Key()
}

func (r *ReviewDecisionRepository) Insert(_ context.Context, decision domain.ReviewDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := reviewIdentityKey(decision)
	if _, ok := r.items[key]; ok {
		return &domain.ConflictError{Message: "review decision already exists for extract"}
	}
	r.items[key] = decision
	return nil
}

func (r *ReviewDecisionRepository) GetByIdentity(_ context.Context, identity domain.ExtractIdentity) (domain.ReviewDecision, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	got, ok := r.items[identity.Key()]
	return got, ok, nil
}

func (r *ReviewDecisionRepository) GetByLoreID(_ context.Context, loreEntryID string) (domain.ReviewDecision, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var found domain.ReviewDecision
	ok := false
	for _, d := range r.items {
		if d.LoreEntryID == loreEntryID {
			if !ok || d.DecidedAt.After(found.DecidedAt) {
				found = d
				ok = true
			}
		}
	}
	return found, ok, nil
}

func (r *ReviewDecisionRepository) ListByScope(_ context.Context, scope domain.Scope) ([]domain.ReviewDecision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ReviewDecision, 0)
	for _, d := range r.items {
		if d.Scope.Kind == scope.Kind && d.Scope.Key == scope.Key {
			out = append(out, d)
		}
	}
	return out, nil
}
