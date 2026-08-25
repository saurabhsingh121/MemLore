package queries

import (
	"context"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ListLoreByScopeHandler lists lore entries for an exact scope.
type ListLoreByScopeHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListLoreByScopeHandler(begin ports.UnitOfWorkFactory) *ListLoreByScopeHandler {
	return &ListLoreByScopeHandler{begin: begin}
}

func (h *ListLoreByScopeHandler) Handle(ctx context.Context, scope domain.Scope) ([]domain.LoreEntry, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.LoreEntries().ListByScope(ctx, scope)
}
