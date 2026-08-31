package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// AcceptReviewCommand promotes a pending observational extract.
type AcceptReviewCommand struct {
	EntryID   string
	ActorID   string
	Statement *string
}

// AcceptReviewHandler accepts suggested lore with human verification provenance.
type AcceptReviewHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewAcceptReviewHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *AcceptReviewHandler {
	return &AcceptReviewHandler{begin: begin, clock: clock}
}

func (h *AcceptReviewHandler) Handle(ctx context.Context, cmd AcceptReviewCommand) (domain.LoreEntry, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	id := strings.TrimSpace(cmd.EntryID)
	if id == "" {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "id must be non-empty"}
	}
	statement := ""
	if cmd.Statement != nil {
		if strings.TrimSpace(*cmd.Statement) == "" {
			return domain.LoreEntry{}, &domain.ValidationError{Message: "statement must be non-empty"}
		}
		statement = *cmd.Statement
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)

	predecessor, err := uow.LoreEntries().Get(ctx, id)
	if err != nil {
		return domain.LoreEntry{}, &domain.NotFoundError{Message: fmt.Sprintf("review item %s not found", id)}
	}

	identity, err := domain.ExtractIdentityOf(predecessor)
	if err != nil {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "lore entry is not a pending suggested-lore item"}
	}
	existing, ok, err := uow.ReviewDecisions().GetByIdentity(ctx, identity)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	if ok {
		if existing.Status == domain.ReviewStatusRejected {
			return domain.LoreEntry{}, &domain.ValidationError{Message: "cannot accept a rejected extract"}
		}
		if existing.SuccessorLoreID == nil {
			return domain.LoreEntry{}, &domain.ValidationError{Message: "accepted extract is missing successor"}
		}
		succ, err := uow.LoreEntries().Get(ctx, *existing.SuccessorLoreID)
		if err != nil {
			return domain.LoreEntry{}, err
		}
		return succ, nil
	}

	now := h.clock.Now()
	result, err := domain.AcceptSuggestedLore(predecessor, statement, actor, now)
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
	if err := uow.ReviewDecisions().Insert(ctx, result.Decision); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.LoreEntry{}, err
	}
	return result.Successor, nil
}
