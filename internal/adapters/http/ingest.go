package httpadapter

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

type ingestGitRequest struct {
	Scope      scopeDTO `json:"scope"`
	Path       string   `json:"path"`
	MaxCommits int      `json:"max_commits"`
}

type ingestRunResponse struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	Scope            scopeDTO   `json:"scope"`
	LocalPath        string     `json:"local_path"`
	ActorID          string     `json:"actor_id"`
	CommitsSeen      int        `json:"commits_seen"`
	CommitsSkipped   int        `json:"commits_skipped"`
	CandidatesStored int        `json:"candidates_stored"`
	CursorSHA        string     `json:"cursor_sha,omitempty"`
	CursorAt         *time.Time `json:"cursor_at,omitempty"`
	ErrorSummary     string     `json:"error_summary,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
}

type ingestRunListResponse struct {
	Items []ingestRunResponse `json:"items"`
}

func toIngestRunResponse(run domain.IngestRun) ingestRunResponse {
	return ingestRunResponse{
		ID:               run.ID,
		Status:           string(run.Status),
		Scope:            scopeDTO{Kind: run.Scope.Kind, Key: run.Scope.Key},
		LocalPath:        run.LocalPath,
		ActorID:          run.ActorID,
		CommitsSeen:      run.CommitsSeen,
		CommitsSkipped:   run.CommitsSkipped,
		CandidatesStored: run.CandidatesStored,
		CursorSHA:        run.CursorSHA,
		CursorAt:         run.CursorAt,
		ErrorSummary:     run.ErrorSummary,
		StartedAt:        run.StartedAt.UTC(),
		FinishedAt:       run.FinishedAt,
	}
}

func (h *Handlers) ingestGit(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body ingestGitRequest
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
	run, err := h.IngestGit.Handle(r.Context(), commands.IngestGitCommand{
		Scope:      scope,
		Path:       body.Path,
		ActorID:    actor,
		MaxCommits: body.MaxCommits,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toIngestRunResponse(run))
}

func (h *Handlers) parseIngestScope(r *http.Request) (domain.Scope, error) {
	kindRaw := strings.TrimSpace(r.URL.Query().Get("scope_kind"))
	key := strings.TrimSpace(r.URL.Query().Get("scope_key"))
	if kindRaw == "" || key == "" {
		return domain.Scope{}, &domain.ValidationError{Message: "scope_kind and scope_key are required"}
	}
	kind, err := domain.ParseScopeKind(kindRaw)
	if err != nil {
		return domain.Scope{}, err
	}
	return domain.NewScope(kind, key)
}

func (h *Handlers) listIngestRuns(w http.ResponseWriter, r *http.Request) {
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
	runs, err := h.ListIngestRuns.Handle(r.Context(), queries.ListIngestRunsQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := ingestRunListResponse{Items: make([]ingestRunResponse, 0, len(runs))}
	for _, run := range runs {
		resp.Items = append(resp.Items, toIngestRunResponse(run))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) getIngestRun(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	run, err := h.GetIngestRun.Handle(r.Context(), queries.GetIngestRunQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, run.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toIngestRunResponse(run))
}

func (h *Handlers) listIngestCandidates(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.ListIngestCands.Handle(r.Context(), queries.ListIngestCandidatesQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := loreEntryListResponse{Items: make([]loreEntryResponse, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, presenters.ToLoreEntry(item))
	}
	writeJSON(w, http.StatusOK, resp)
}
