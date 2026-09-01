package commands

import (
	"context"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// CreateDecisionCommand records a human engineering decision.
type CreateDecisionCommand struct {
	Scope              domain.Scope
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

// CreateDecisionHandler dual-writes a Decision and linked lore.
type CreateDecisionHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewCreateDecisionHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *CreateDecisionHandler {
	return &CreateDecisionHandler{begin: begin, clock: clock}
}

func (h *CreateDecisionHandler) Handle(ctx context.Context, cmd CreateDecisionCommand) (domain.Decision, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.Decision{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.Decision{}, err
	}
	defer uow.Rollback(ctx)

	now := h.clock.Now()
	decision, err := domain.NewHumanDecision(domain.NewDecisionInput{
		Scope:              cmd.Scope,
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
	lore, err := domain.NewDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: decision.Choice,
		Scope:     decision.Scope,
		CreatedBy: actor,
		Evidence:  decision.Evidence,
		ID:        decision.ID,
		Now:       now,
	})
	if err != nil {
		return domain.Decision{}, err
	}
	audit, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID:  lore.ID,
		Action:    domain.AuditActionCreate,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return domain.Decision{}, err
	}
	if err := uow.LoreEntries().Add(ctx, lore); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Decisions().Add(ctx, decision); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Audits().Add(ctx, audit); err != nil {
		return domain.Decision{}, err
	}
	outboxEvent, err := domain.NewEpisodeIngestOutboxEvent(lore, now)
	if err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Outbox().Add(ctx, outboxEvent); err != nil {
		return domain.Decision{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.Decision{}, err
	}
	decision.Evidence = lore.Evidence
	decision.Current = true
	return decision, nil
}
