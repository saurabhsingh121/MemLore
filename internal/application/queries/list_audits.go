package queries

import (
	"context"
	"fmt"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ListAuditsHandler lists audit records for a lore entry.
type ListAuditsHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListAuditsHandler(begin ports.UnitOfWorkFactory) *ListAuditsHandler {
	return &ListAuditsHandler{begin: begin}
}

func (h *ListAuditsHandler) Handle(ctx context.Context, entryID string) ([]domain.AuditRecord, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)

	if _, err := uow.LoreEntries().Get(ctx, entryID); err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			return nil, &domain.NotFoundError{
				Message: fmt.Sprintf("lore entry %s not found", entryID),
			}
		}
		return nil, err
	}
	return uow.Audits().ListByTarget(ctx, entryID)
}
