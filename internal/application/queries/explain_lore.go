package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/memlore/memlore/internal/application/authority"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ExplainLoreResult is a lore entry plus audits and an ephemeral authority evaluation.
type ExplainLoreResult struct {
	Entry      domain.LoreEntry
	Audits     []domain.AuditRecord
	Evaluation domain.Evaluation
}

// ExplainLoreHandler returns explain payload inputs for one lore entry.
type ExplainLoreHandler struct {
	begin ports.UnitOfWorkFactory
	now   func() time.Time
}

// NewExplainLoreHandler wires get + audits + authority evaluation.
func NewExplainLoreHandler(begin ports.UnitOfWorkFactory) *ExplainLoreHandler {
	return &ExplainLoreHandler{begin: begin, now: time.Now}
}

// Handle loads the entry (including stale), audits, and evaluates authority.
func (h *ExplainLoreHandler) Handle(ctx context.Context, entryID string) (ExplainLoreResult, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return ExplainLoreResult{}, err
	}
	defer uow.Rollback(ctx)

	entry, err := uow.LoreEntries().Get(ctx, entryID)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			return ExplainLoreResult{}, &domain.NotFoundError{
				Message: fmt.Sprintf("lore entry %s not found", entryID),
			}
		}
		return ExplainLoreResult{}, err
	}
	audits, err := uow.Audits().ListByTarget(ctx, entryID)
	if err != nil {
		return ExplainLoreResult{}, err
	}
	scope := entry.Scope
	eval := authority.EvaluateGovernance(entry, &scope, h.now())
	return ExplainLoreResult{Entry: entry, Audits: audits, Evaluation: eval}, nil
}
