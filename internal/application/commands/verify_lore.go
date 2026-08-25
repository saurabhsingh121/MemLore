package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// VerifyLoreCommand marks a lore entry verified.
type VerifyLoreCommand struct {
	EntryID string
	ActorID string
}

// VerifyLoreHandler handles lore verification.
type VerifyLoreHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewVerifyLoreHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *VerifyLoreHandler {
	return &VerifyLoreHandler{begin: begin, clock: clock}
}

func (h *VerifyLoreHandler) Handle(ctx context.Context, cmd VerifyLoreCommand) (domain.LoreEntry, error) {
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

	updated, audit, err := domain.ApplyVerification(entry, actor, h.clock.Now())
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

func isNotFound(err error, target **domain.NotFoundError) bool {
	if err == nil {
		return false
	}
	if nf, ok := err.(*domain.NotFoundError); ok {
		*target = nf
		return true
	}
	return false
}
