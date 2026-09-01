package queries

import (
	"context"
	"sort"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ListDecisionsQuery lists current Decisions for a repository.
type ListDecisionsQuery struct {
	Scope domain.Scope
}

// ListDecisionsHandler unions human-recorded Decisions and ADR projections.
type ListDecisionsHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListDecisionsHandler(begin ports.UnitOfWorkFactory) *ListDecisionsHandler {
	return &ListDecisionsHandler{begin: begin}
}

func (h *ListDecisionsHandler) Handle(ctx context.Context, q ListDecisionsQuery) ([]domain.Decision, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "decision scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	stored, err := uow.Decisions().ListByScope(ctx, q.Scope)
	if err != nil {
		return nil, err
	}
	entries, err := uow.LoreEntries().ListByScope(ctx, q.Scope)
	if err != nil {
		return nil, err
	}
	loreByID := make(map[string]domain.LoreEntry, len(entries))
	for _, entry := range entries {
		loreByID[entry.ID] = entry
	}

	out := make([]domain.Decision, 0)
	humanIDs := make(map[string]struct{}, len(stored))
	for _, d := range stored {
		humanIDs[d.ID] = struct{}{}
		lore, ok := loreByID[d.ID]
		if !ok {
			continue
		}
		got := hydrateHumanDecision(d, lore)
		if !got.Current {
			continue
		}
		out = append(out, got)
	}
	for _, entry := range entries {
		if _, ok := humanIDs[entry.ID]; ok {
			continue
		}
		if !domain.CanProjectADRDecision(entry) || !domain.IsCurrent(entry) {
			continue
		}
		proj, err := domain.ProjectADRDecision(entry)
		if err != nil {
			return nil, err
		}
		out = append(out, proj)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].DecidedAt.After(out[j].DecidedAt)
	})
	return out, nil
}
