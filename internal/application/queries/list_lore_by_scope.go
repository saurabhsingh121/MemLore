package queries

import (
	"context"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ListLoreByScopeQuery lists lore for an exact scope.
type ListLoreByScopeQuery struct {
	Scope        domain.Scope
	IncludeStale bool
}

// ListLoreByScopeHandler lists lore entries for an exact scope.
type ListLoreByScopeHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListLoreByScopeHandler(begin ports.UnitOfWorkFactory) *ListLoreByScopeHandler {
	return &ListLoreByScopeHandler{begin: begin}
}

func (h *ListLoreByScopeHandler) Handle(ctx context.Context, query ListLoreByScopeQuery) ([]domain.LoreEntry, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	items, err := uow.LoreEntries().ListByScope(ctx, query.Scope)
	if err != nil {
		return nil, err
	}
	if query.IncludeStale {
		return items, nil
	}
	return appcontext.FilterCurrent(items), nil
}
