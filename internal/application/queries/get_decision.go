package queries

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// GetDecisionQuery loads one Decision by id (human row or ADR projection).
type GetDecisionQuery struct {
	ID string
}

// GetDecisionHandler loads a Decision from the decisions table or ADR lore.
type GetDecisionHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetDecisionHandler(begin ports.UnitOfWorkFactory) *GetDecisionHandler {
	return &GetDecisionHandler{begin: begin}
}

func (h *GetDecisionHandler) Handle(ctx context.Context, q GetDecisionQuery) (domain.Decision, error) {
	id := strings.TrimSpace(q.ID)
	if id == "" {
		return domain.Decision{}, &domain.ValidationError{Message: "id must be non-empty"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.Decision{}, err
	}
	defer uow.Rollback(ctx)

	decision, err := loadDecision(ctx, uow, id)
	if err != nil {
		return domain.Decision{}, err
	}
	return decision, nil
}

func loadDecision(ctx context.Context, uow ports.UnitOfWork, id string) (domain.Decision, error) {
	stored, err := uow.Decisions().Get(ctx, id)
	if err == nil {
		lore, loreErr := uow.LoreEntries().Get(ctx, id)
		if loreErr != nil {
			return domain.Decision{}, loreErr
		}
		return hydrateHumanDecision(stored, lore), nil
	}
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		return domain.Decision{}, err
	}
	lore, err := uow.LoreEntries().Get(ctx, id)
	if err != nil {
		return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", id)}
	}
	if !domain.CanProjectADRDecision(lore) {
		return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", id)}
	}
	return domain.ProjectADRDecision(lore)
}

func hydrateHumanDecision(d domain.Decision, lore domain.LoreEntry) domain.Decision {
	d.Evidence = lore.Evidence
	if d.Evidence == nil {
		d.Evidence = []domain.EvidenceReference{}
	}
	d.Current = domain.DecisionIsCurrent(d, lore)
	if d.SupersededByID == nil && lore.SupersededByID != nil {
		d.SupersededByID = lore.SupersededByID
	}
	return d
}
