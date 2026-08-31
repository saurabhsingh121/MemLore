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
	Scope        domain.Scope
	EvidenceType domain.EvidenceType // optional filter (pr or commit)
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
		if q.EvidenceType == domain.EvidenceTypeADR {
			if e.Origin != domain.KnowledgeOriginArchitectureDecision {
				continue
			}
			if !hasEvidenceType(e, domain.EvidenceTypeADR) {
				continue
			}
			out = append(out, e)
			continue
		}
		if e.Origin == domain.KnowledgeOriginRepositoryObservation {
			if q.EvidenceType != "" && !hasEvidenceType(e, q.EvidenceType) {
				continue
			}
			out = append(out, e)
		}
	}
	return out, nil
}

func hasEvidenceType(e domain.LoreEntry, want domain.EvidenceType) bool {
	for _, ev := range e.Evidence {
		if ev.Type == want {
			return true
		}
	}
	return false
}

// ListPRIngestRunsQuery lists PR ingest runs for a repository scope.
type ListPRIngestRunsQuery struct {
	Scope domain.Scope
}

// GetPRIngestRunQuery loads one PR ingest run by id.
type GetPRIngestRunQuery struct {
	ID string
}

// ListPRIngestRunsHandler lists PR runs newest first.
type ListPRIngestRunsHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListPRIngestRunsHandler(begin ports.UnitOfWorkFactory) *ListPRIngestRunsHandler {
	return &ListPRIngestRunsHandler{begin: begin}
}

func (h *ListPRIngestRunsHandler) Handle(ctx context.Context, q ListPRIngestRunsQuery) ([]domain.PRIngestRun, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "ingest scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.PRIngest().ListRunsByScope(ctx, q.Scope)
}

// GetPRIngestRunHandler loads a single PR ingest run.
type GetPRIngestRunHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetPRIngestRunHandler(begin ports.UnitOfWorkFactory) *GetPRIngestRunHandler {
	return &GetPRIngestRunHandler{begin: begin}
}

func (h *GetPRIngestRunHandler) Handle(ctx context.Context, q GetPRIngestRunQuery) (domain.PRIngestRun, error) {
	if q.ID == "" {
		return domain.PRIngestRun{}, &domain.ValidationError{Message: "run id is required"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.PRIngestRun{}, err
	}
	defer uow.Rollback(ctx)
	return uow.PRIngest().GetRun(ctx, q.ID)
}

// ListADRIngestRunsQuery lists ADR ingest runs for a repository scope.
type ListADRIngestRunsQuery struct {
	Scope domain.Scope
}

// GetADRIngestRunQuery loads one ADR ingest run by id.
type GetADRIngestRunQuery struct {
	ID string
}

// ListADRIngestRunsHandler lists ADR runs newest first.
type ListADRIngestRunsHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewListADRIngestRunsHandler(begin ports.UnitOfWorkFactory) *ListADRIngestRunsHandler {
	return &ListADRIngestRunsHandler{begin: begin}
}

func (h *ListADRIngestRunsHandler) Handle(ctx context.Context, q ListADRIngestRunsQuery) ([]domain.ADRIngestRun, error) {
	if q.Scope.Kind != domain.ScopeKindRepository {
		return nil, &domain.ValidationError{Message: "ingest scope kind must be repository"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer uow.Rollback(ctx)
	return uow.ADRIngest().ListRunsByScope(ctx, q.Scope)
}

// GetADRIngestRunHandler loads a single ADR ingest run.
type GetADRIngestRunHandler struct {
	begin ports.UnitOfWorkFactory
}

func NewGetADRIngestRunHandler(begin ports.UnitOfWorkFactory) *GetADRIngestRunHandler {
	return &GetADRIngestRunHandler{begin: begin}
}

func (h *GetADRIngestRunHandler) Handle(ctx context.Context, q GetADRIngestRunQuery) (domain.ADRIngestRun, error) {
	if q.ID == "" {
		return domain.ADRIngestRun{}, &domain.ValidationError{Message: "run id is required"}
	}
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.ADRIngestRun{}, err
	}
	defer uow.Rollback(ctx)
	return uow.ADRIngest().GetRun(ctx, q.ID)
}
