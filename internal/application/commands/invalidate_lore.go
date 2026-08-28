package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// InvalidateLoreCommand marks a lore entry invalidated.
type InvalidateLoreCommand struct {
	EntryID string
	ActorID string
}

// InvalidateLoreHandler handles lore invalidation.
type InvalidateLoreHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewInvalidateLoreHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *InvalidateLoreHandler {
	return &InvalidateLoreHandler{begin: begin, clock: clock}
}

func (h *InvalidateLoreHandler) Handle(ctx context.Context, cmd InvalidateLoreCommand) (domain.LoreEntry, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)

	entry, err := uow.LoreEntries().Get(ctx, cmd.EntryID)
	if err != nil {
		var nf *domain.NotFoundError
		if !isNotFound(err, &nf) {
			return domain.LoreEntry{}, err
		}
		return domain.LoreEntry{}, &domain.NotFoundError{
			Message: fmt.Sprintf("lore entry %s not found", cmd.EntryID),
		}
	}

	updated, audit, err := domain.ApplyInvalidation(entry, actor, h.clock.Now())
	if err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.LoreEntries().Save(ctx, updated); err != nil {
		return domain.LoreEntry{}, err
	}
	if audit != nil {
		if err := uow.Audits().Add(ctx, *audit); err != nil {
			return domain.LoreEntry{}, err
		}
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.LoreEntry{}, err
	}
	return updated, nil
}
