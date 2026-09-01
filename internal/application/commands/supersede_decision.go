package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// SupersedeDecisionCommand replaces a current Decision with a new human successor.
type SupersedeDecisionCommand struct {
	PredecessorID      string
	Question           string
	Choice             string
	Rationale          string
	Alternatives       []domain.DecisionAlternative
	Consequences       string
	Owner              string
	DecidedAt          time.Time
	AffectedComponents []string
	Evidence           []domain.EvidenceReference
	ActorID            string
}

// SupersedeDecisionHandler creates a successor Decision and supersedes predecessor lore.
type SupersedeDecisionHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewSupersedeDecisionHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *SupersedeDecisionHandler {
	return &SupersedeDecisionHandler{begin: begin, clock: clock}
}

func (h *SupersedeDecisionHandler) Handle(ctx context.Context, cmd SupersedeDecisionCommand) (domain.Decision, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.Decision{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	predID := strings.TrimSpace(cmd.PredecessorID)
	if predID == "" {
		return domain.Decision{}, &domain.ValidationError{Message: "id must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.Decision{}, err
	}
	defer uow.Rollback(ctx)

	predLore, err := uow.LoreEntries().Get(ctx, predID)
	if err != nil {
		return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", predID)}
	}
	stored, storedErr := uow.Decisions().Get(ctx, predID)
	hasRow := storedErr == nil
	if !hasRow {
		var nf *domain.NotFoundError
		if storedErr != nil && !errors.As(storedErr, &nf) {
			return domain.Decision{}, storedErr
		}
		if !domain.CanProjectADRDecision(predLore) {
			return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", predID)}
		}
	} else if !domain.DecisionIsCurrent(stored, predLore) {
		return domain.Decision{}, &domain.ValidationError{Message: "cannot supersede a superseded lore entry"}
	}

	now := h.clock.Now()
	successor, err := domain.NewHumanDecision(domain.NewDecisionInput{
		Scope:              predLore.Scope,
		Question:           cmd.Question,
		Choice:             cmd.Choice,
		Rationale:          cmd.Rationale,
		Alternatives:       cmd.Alternatives,
		Consequences:       cmd.Consequences,
		Owner:              cmd.Owner,
		DecidedAt:          cmd.DecidedAt,
		AffectedComponents: cmd.AffectedComponents,
		Evidence:           cmd.Evidence,
		CreatedBy:          actor,
		Now:                now,
	})
	if err != nil {
		return domain.Decision{}, err
	}
	succLore, err := domain.NewDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: successor.Choice,
		Scope:     successor.Scope,
		CreatedBy: actor,
		Evidence:  successor.Evidence,
		ID:        successor.ID,
		Now:       now,
	})
	if err != nil {
		return domain.Decision{}, err
	}
	result, err := domain.ApplySupersessionWithSuccessor(predLore, succLore, actor, now)
	if err != nil {
		return domain.Decision{}, err
	}
	if err := uow.LoreEntries().Add(ctx, result.Successor); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.LoreEntries().Save(ctx, result.Predecessor); err != nil {
		return domain.Decision{}, err
	}
	if hasRow {
		updated := stored.WithSupersededBy(successor.ID)
		if err := uow.Decisions().Save(ctx, updated); err != nil {
			return domain.Decision{}, err
		}
	}
	if err := uow.Decisions().Add(ctx, successor); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Audits().Add(ctx, result.SupersedeAudit); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Audits().Add(ctx, result.CreateAudit); err != nil {
		return domain.Decision{}, err
	}
	outboxEvent, err := domain.NewEpisodeIngestOutboxEvent(result.Successor, now)
	if err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Outbox().Add(ctx, outboxEvent); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.Decision{}, err
	}
	successor.Evidence = result.Successor.Evidence
	successor.Current = true
	return successor, nil
}
