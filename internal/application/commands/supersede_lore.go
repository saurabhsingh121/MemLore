package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// SupersedeLoreCommand replaces a lore entry with a successor.
type SupersedeLoreCommand struct {
	EntryID   string
	Statement string
	ActorID   string
	Evidence  []domain.EvidenceReference
}

// SupersedeLoreHandler handles lore supersession.
type SupersedeLoreHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewSupersedeLoreHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *SupersedeLoreHandler {
	return &SupersedeLoreHandler{begin: begin, clock: clock}
}

func (h *SupersedeLoreHandler) Handle(ctx context.Context, cmd SupersedeLoreCommand) (domain.LoreEntry, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)

	predecessor, err := uow.LoreEntries().Get(ctx, cmd.EntryID)
	if err != nil {
		var nf *domain.NotFoundError
		if !isNotFound(err, &nf) {
			return domain.LoreEntry{}, err
		}
		return domain.LoreEntry{}, &domain.NotFoundError{
			Message: fmt.Sprintf("lore entry %s not found", cmd.EntryID),
		}
	}

	now := h.clock.Now()
	result, err := domain.ApplySupersession(predecessor, cmd.Statement, actor, cmd.Evidence, now)
	if err != nil {
		return domain.LoreEntry{}, err
	}

	if err := uow.LoreEntries().Add(ctx, result.Successor); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.LoreEntries().Save(ctx, result.Predecessor); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Audits().Add(ctx, result.SupersedeAudit); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Audits().Add(ctx, result.CreateAudit); err != nil {
		return domain.LoreEntry{}, err
	}
	outboxEvent, err := domain.NewEpisodeIngestOutboxEvent(result.Successor, now)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Outbox().Add(ctx, outboxEvent); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.LoreEntry{}, err
	}
	return result.Successor, nil
}
