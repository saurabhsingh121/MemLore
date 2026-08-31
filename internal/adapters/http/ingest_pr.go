package httpadapter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type ingestPRRequest struct {
	Scope  scopeDTO `json:"scope"`
	PR     int      `json:"pr"`
	MaxPRs int      `json:"max_prs"`
}

type prIngestRunResponse struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	Scope            scopeDTO   `json:"scope"`
	ActorID          string     `json:"actor_id"`
	PR               int        `json:"pr,omitempty"`
	PRsSeen          int        `json:"prs_seen"`
	PRsSkipped       int        `json:"prs_skipped"`
	CandidatesStored int        `json:"candidates_stored"`
	CursorPR         int        `json:"cursor_pr,omitempty"`
	CursorAt         *time.Time `json:"cursor_at,omitempty"`
	ErrorSummary     string     `json:"error_summary,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
}

type prIngestRunListResponse struct {
	Items []prIngestRunResponse `json:"items"`
}

func toPRIngestRunResponse(run domain.PRIngestRun) prIngestRunResponse {
	return prIngestRunResponse{
		ID:               run.ID,
		Status:           string(run.Status),
		Scope:            scopeDTO{Kind: run.Scope.Kind, Key: run.Scope.Key},
		ActorID:          run.ActorID,
		PR:               run.PRNumber,
		PRsSeen:          run.PRsSeen,
		PRsSkipped:       run.PRsSkipped,
		CandidatesStored: run.CandidatesStored,
		CursorPR:         run.CursorPR,
		CursorAt:         run.CursorAt,
		ErrorSummary:     run.ErrorSummary,
		StartedAt:        run.StartedAt.UTC(),
		FinishedAt:       run.FinishedAt,
	}
}

func (h *Handlers) ingestPR(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body ingestPRRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	kind, err := domain.ParseScopeKind(string(body.Scope.Kind))
	if err != nil {
		handleDomainError(w, err)
		return
	}
	scope, err := domain.NewScope(kind, body.Scope.Key)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	if h.IngestPR == nil {
		writeError(w, http.StatusBadRequest, "validation_error", "pull request reader is not configured")
		return
	}
	run, err := h.IngestPR.Handle(r.Context(), commands.IngestPullRequestsCommand{
		Scope:   scope,
		ActorID: actor,
		PR:      body.PR,
		MaxPRs:  body.MaxPRs,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPRIngestRunResponse(run))
}

func (h *Handlers) listPRIngestRuns(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	scope, err := h.parseIngestScope(r)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	if h.ListPRIngestRuns == nil {
		writeJSON(w, http.StatusOK, prIngestRunListResponse{Items: []prIngestRunResponse{}})
		return
	}
	runs, err := h.ListPRIngestRuns.Handle(r.Context(), queries.ListPRIngestRunsQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := prIngestRunListResponse{Items: make([]prIngestRunResponse, 0, len(runs))}
	for _, run := range runs {
		resp.Items = append(resp.Items, toPRIngestRunResponse(run))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) getPRIngestRun(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if h.GetPRIngestRun == nil {
		writeError(w, http.StatusNotFound, "not_found", "pr ingest run not found")
		return
	}
	id := chi.URLParam(r, "id")
	run, err := h.GetPRIngestRun.Handle(r.Context(), queries.GetPRIngestRunQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, run.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toPRIngestRunResponse(run))
}
