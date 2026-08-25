package queries

import (
	"context"
	"fmt"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// GetLoreHandler retrieves a lore entry by id.
type GetLoreHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetLoreHandler(begin ports.UnitOfWorkFactory) *GetLoreHandler {
	return &GetLoreHandler{begin: begin}
}

func (h *GetLoreHandler) Handle(ctx context.Context, entryID string) (domain.LoreEntry, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)

	entry, err := uow.LoreEntries().Get(ctx, entryID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			return domain.LoreEntry{}, &domain.NotFoundError{
				Message: fmt.Sprintf("lore entry %s not found", entryID),
			}
		}
		return domain.LoreEntry{}, err
	}
	return entry, nil
}
