package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// RejectReviewCommand records a negative decision for a pending extract.
type RejectReviewCommand struct {
	EntryID string
	ActorID string
}

// RejectReviewResult is returned to CLI/REST after reject.
type RejectReviewResult struct {
	ID     string
	Status domain.ReviewStatus
	Actor  string
}

// RejectReviewHandler records Reject without mutating observational lore.
type RejectReviewHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewRejectReviewHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *RejectReviewHandler {
	return &RejectReviewHandler{begin: begin, clock: clock}
}

func (h *RejectReviewHandler) Handle(ctx context.Context, cmd RejectReviewCommand) (RejectReviewResult, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return RejectReviewResult{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	id := strings.TrimSpace(cmd.EntryID)
	if id == "" {
		return RejectReviewResult{}, &domain.ValidationError{Message: "id must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return RejectReviewResult{}, err
	}
	defer uow.Rollback(ctx)

	entry, err := uow.LoreEntries().Get(ctx, id)
	if err != nil {
		return RejectReviewResult{}, &domain.NotFoundError{Message: fmt.Sprintf("review item %s not found", id)}
	}

	identity, err := domain.ExtractIdentityOf(entry)
	if err != nil {
		return RejectReviewResult{}, &domain.ValidationError{Message: "lore entry is not a pending suggested-lore item"}
	}
	existing, ok, err := uow.ReviewDecisions().GetByIdentity(ctx, identity)
	if err != nil {
		return RejectReviewResult{}, err
	}
	if ok {
		if existing.Status == domain.ReviewStatusAccepted {
			return RejectReviewResult{}, &domain.ValidationError{Message: "cannot reject an accepted extract"}
		}
		return RejectReviewResult{ID: entry.ID, Status: existing.Status, Actor: existing.ActorID}, nil
	}

	decision, err := domain.RejectSuggestedLore(entry, actor, h.clock.Now())
	if err != nil {
		return RejectReviewResult{}, err
	}
	if err := uow.ReviewDecisions().Insert(ctx, decision); err != nil {
		return RejectReviewResult{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return RejectReviewResult{}, err
	}
	return RejectReviewResult{ID: entry.ID, Status: decision.Status, Actor: decision.ActorID}, nil
}
