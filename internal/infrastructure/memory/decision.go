package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.DecisionRepository = (*DecisionRepository)(nil)

// DecisionRepository is an in-memory Decision store for tests.
type DecisionRepository struct {
	mu    sync.RWMutex
	items map[string]domain.Decision
}

func NewDecisionRepository() *DecisionRepository {
	return &DecisionRepository{items: make(map[string]domain.Decision)}
}

func cloneDecision(d domain.Decision) domain.Decision {
	out := d
	if d.Alternatives != nil {
		out.Alternatives = append([]domain.DecisionAlternative{}, d.Alternatives...)
	} else {
		out.Alternatives = []domain.DecisionAlternative{}
	}
	if d.AffectedComponents != nil {
		out.AffectedComponents = append([]string{}, d.AffectedComponents...)
	} else {
		out.AffectedComponents = []string{}
	}
	if d.Evidence != nil {
		out.Evidence = append([]domain.EvidenceReference{}, d.Evidence...)
	} else {
		out.Evidence = []domain.EvidenceReference{}
	}
	if d.SupersededByID != nil {
		id := *d.SupersededByID
		out.SupersededByID = &id
	}
	return out
}

func (r *DecisionRepository) Add(_ context.Context, decision domain.Decision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[decision.ID]; ok {
		return &domain.ConflictError{Message: "decision already exists"}
	}
	r.items[decision.ID] = cloneDecision(decision)
	return nil
}

func (r *DecisionRepository) Get(_ context.Context, id string) (domain.Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	got, ok := r.items[id]
	if !ok {
		return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", id)}
	}
	return cloneDecision(got), nil
}

func (r *DecisionRepository) Save(_ context.Context, decision domain.Decision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[decision.ID]; !ok {
		return &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", decision.ID)}
	}
	r.items[decision.ID] = cloneDecision(decision)
	return nil
}

func (r *DecisionRepository) ListByScope(_ context.Context, scope domain.Scope) ([]domain.Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Decision, 0)
	for _, d := range r.items {
		if d.Scope.Kind == scope.Kind && d.Scope.Key == scope.Key {
			out = append(out, cloneDecision(d))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].DecidedAt.After(out[j].DecidedAt)
	})
	return out, nil
}
