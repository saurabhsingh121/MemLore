package commands

import (
	"context"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// CreateLoreCommand stores a human-authored lore entry.
type CreateLoreCommand struct {
	Statement string
	Scope     domain.Scope
	ActorID   string
	Evidence  []domain.EvidenceReference
}

// CreateLoreHandler handles lore creation.
type CreateLoreHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
}

func NewCreateLoreHandler(begin ports.UnitOfWorkFactory, clock ports.Clock) *CreateLoreHandler {
	return &CreateLoreHandler{begin: begin, clock: clock}
}

func (h *CreateLoreHandler) Handle(ctx context.Context, cmd CreateLoreCommand) (domain.LoreEntry, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.LoreEntry{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}

	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)

	now := h.clock.Now()
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: cmd.Statement,
		Scope:     cmd.Scope,
		CreatedBy: actor,
		Evidence:  cmd.Evidence,
		Now:       now,
	})
	if err != nil {
		return domain.LoreEntry{}, err
	}
	audit, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID:  entry.ID,
		Action:    domain.AuditActionCreate,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return domain.LoreEntry{}, err
	}

	if err := uow.LoreEntries().Add(ctx, entry); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Audits().Add(ctx, audit); err != nil {
		return domain.LoreEntry{}, err
	}
	if err := uow.Commit(ctx); err != nil {
		return domain.LoreEntry{}, err
	}
	return entry, nil
}
