package queries

import (
	"context"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ListIngestRunsQuery lists ingest runs for a repository scope.
type ListIngestRunsQuery struct {
	Scope domain.Scope
}

// GetIngestRunQuery loads one ingest run by id.
type GetIngestRunQuery struct {
	ID string
}

// ListIngestCandidatesQuery lists current observational lore for a repository.
type ListIngestCandidatesQuery struct {
	Scope domain.Scope
}

// ListIngestRunsHandler lists runs newest first.
type ListIngestRunsHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListIngestRunsHandler(begin ports.UnitOfWorkFactory) *ListIngestRunsHandler {
	return &ListIngestRunsHandler{begin: begin}
}

func (h *ListIngestRunsHandler) Handle(ctx context.Context, q ListIngestRunsQuery) ([]domain.IngestRun, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "ingest scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.Ingest().ListRunsByScope(ctx, q.Scope)
}

// GetIngestRunHandler loads a single run.
type GetIngestRunHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetIngestRunHandler(begin ports.UnitOfWorkFactory) *GetIngestRunHandler {
	return &GetIngestRunHandler{begin: begin}
}

func (h *GetIngestRunHandler) Handle(ctx context.Context, q GetIngestRunQuery) (domain.IngestRun, error) {
	if q.ID == "" {
		return domain.IngestRun{}, &domain.ValidationError{Message: "run id is required"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.IngestRun{}, err
	}
	defer uow.Rollback(ctx)
	return uow.Ingest().GetRun(ctx, q.ID)
}

// ListIngestCandidatesHandler lists current repository_observation lore.
type ListIngestCandidatesHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListIngestCandidatesHandler(begin ports.UnitOfWorkFactory) *ListIngestCandidatesHandler {
	return &ListIngestCandidatesHandler{begin: begin}
}

func (h *ListIngestCandidatesHandler) Handle(ctx context.Context, q ListIngestCandidatesQuery) ([]domain.LoreEntry, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "ingest scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	items, err := uow.LoreEntries().ListByScope(ctx, q.Scope)
	if err != nil {
		return nil, err
	}
	current := appcontext.FilterCurrent(items)
	out := make([]domain.LoreEntry, 0, len(current))
	for _, e := range current {
		if e.Origin == domain.KnowledgeOriginRepositoryObservation {
			out = append(out, e)
		}
	}
	return out, nil
}
