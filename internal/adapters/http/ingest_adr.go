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

type ingestADRRequest struct {
	Scope   scopeDTO `json:"scope"`
	Path    string   `json:"path"`
	ADRDirs []string `json:"adr_dirs"`
}

type adrIngestRunResponse struct {
	ID             string     `json:"id"`
	Status         string     `json:"status"`
	Scope          scopeDTO   `json:"scope"`
	ActorID        string     `json:"actor_id"`
	Path           string     `json:"path"`
	FilesSeen      int        `json:"files_seen"`
	FilesSkipped   int        `json:"files_skipped"`
	LoreStored     int        `json:"lore_stored"`
	LoreSuperseded int        `json:"lore_superseded"`
	ErrorSummary   string     `json:"error_summary,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
}

type adrIngestRunListResponse struct {
	Items []adrIngestRunResponse `json:"items"`
}

func toADRIngestRunResponse(run domain.ADRIngestRun) adrIngestRunResponse {
	return adrIngestRunResponse{
		ID:             run.ID,
		Status:         string(run.Status),
		Scope:          scopeDTO{Kind: run.Scope.Kind, Key: run.Scope.Key},
		ActorID:        run.ActorID,
		Path:           run.LocalPath,
		FilesSeen:      run.FilesSeen,
		FilesSkipped:   run.FilesSkipped,
		LoreStored:     run.LoreStored,
		LoreSuperseded: run.LoreSuperseded,
		ErrorSummary:   run.ErrorSummary,
		StartedAt:      run.StartedAt.UTC(),
		FinishedAt:     run.FinishedAt,
	}
}

func (h *Handlers) ingestADR(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body ingestADRRequest
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
	if h.IngestADR == nil {
		writeError(w, http.StatusBadRequest, "validation_error", "adr reader is not configured")
		return
	}
	run, err := h.IngestADR.Handle(r.Context(), commands.IngestADRsCommand{
		Scope:     scope,
		Path:      body.Path,
		ActorID:   actor,
		ExtraDirs: body.ADRDirs,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toADRIngestRunResponse(run))
}

func (h *Handlers) listADRIngestRuns(w http.ResponseWriter, r *http.Request) {
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
	if h.ListADRIngestRuns == nil {
		writeJSON(w, http.StatusOK, adrIngestRunListResponse{Items: []adrIngestRunResponse{}})
		return
	}
	runs, err := h.ListADRIngestRuns.Handle(r.Context(), queries.ListADRIngestRunsQuery{Scope: scope})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := adrIngestRunListResponse{Items: make([]adrIngestRunResponse, 0, len(runs))}
	for _, run := range runs {
		resp.Items = append(resp.Items, toADRIngestRunResponse(run))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) getADRIngestRun(w http.ResponseWriter, r *http.Request) {
	if err := h.ensureReadAuthorized(r); err != nil {
		handleDomainError(w, err)
		return
	}
	if h.GetADRIngestRun == nil {
		writeError(w, http.StatusNotFound, "not_found", "adr ingest run not found")
		return
	}
	id := chi.URLParam(r, "id")
	run, err := h.GetADRIngestRun.Handle(r.Context(), queries.GetADRIngestRunQuery{ID: id})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, run.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toADRIngestRunResponse(run))
}
