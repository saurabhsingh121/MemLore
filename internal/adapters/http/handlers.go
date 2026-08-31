package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/memlore/memlore/internal/adapters/presenters"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/gitcli"
	"github.com/memlore/memlore/internal/infrastructure/githubhttp"
)

// Handlers exposes lore REST endpoints.
type Handlers struct {
	CreateLore        *commands.CreateLoreHandler
	VerifyLore        *commands.VerifyLoreHandler
	InvalidateLore    *commands.InvalidateLoreHandler
	SupersedeLore     *commands.SupersedeLoreHandler
	GetLore           *queries.GetLoreHandler
	ListLoreByScope   *queries.ListLoreByScopeHandler
	ListAudits        *queries.ListAuditsHandler
	SearchKnowledge   *queries.SearchKnowledgeHandler
	CompileContext    *queries.CompileContextHandler
	RepositoryProfile *queries.RepositoryProfileHandler
	ExplainLore       *queries.ExplainLoreHandler
	IngestGit         *commands.IngestGitHandler
	IngestPR          *commands.IngestPullRequestsHandler
	ListIngestRuns    *queries.ListIngestRunsHandler
	GetIngestRun      *queries.GetIngestRunHandler
	ListPRIngestRuns  *queries.ListPRIngestRunsHandler
	GetPRIngestRun    *queries.GetPRIngestRunHandler
	ListIngestCands   *queries.ListIngestCandidatesHandler
	Auth              *appauth.Service
	Authz             *authz.Gate
	Membership        ports.MembershipDirectory
	Version           string
}

// NewHandlers wires application handlers from a unit-of-work factory and clock.
func NewHandlers(begin ports.UnitOfWorkFactory, clock ports.Clock, graph ports.KnowledgeGraph, version string) *Handlers {
	search := queries.NewSearchKnowledgeHandler(begin, graph, nil)
	list := queries.NewListLoreByScopeHandler(begin)
	return &Handlers{
		CreateLore:        commands.NewCreateLoreHandler(begin, clock),
		VerifyLore:        commands.NewVerifyLoreHandler(begin, clock),
		InvalidateLore:    commands.NewInvalidateLoreHandler(begin, clock),
		SupersedeLore:     commands.NewSupersedeLoreHandler(begin, clock),
		GetLore:           queries.NewGetLoreHandler(begin),
		ListLoreByScope:   list,
		ListAudits:        queries.NewListAuditsHandler(begin),
		SearchKnowledge:   search,
		CompileContext:    queries.NewCompileContextHandler(search, list),
		RepositoryProfile: queries.NewRepositoryProfileHandler(list, search),
		ExplainLore:       queries.NewExplainLoreHandler(begin),
		IngestGit:         commands.NewIngestGitHandler(begin, clock, gitcli.NewReader()),
		IngestPR:          commands.NewIngestPullRequestsHandler(begin, clock, githubhttp.NewReader("", githubhttp.TokenFromEnv(), nil)),
		ListIngestRuns:    queries.NewListIngestRunsHandler(begin),
		GetIngestRun:      queries.NewGetIngestRunHandler(begin),
		ListPRIngestRuns:  queries.NewListPRIngestRunsHandler(begin),
		GetPRIngestRun:    queries.NewGetPRIngestRunHandler(begin),
		ListIngestCands:   queries.NewListIngestCandidatesHandler(begin),
		Auth:              appauth.NewService(appauth.Config{}, nil),
		Version:           version,
	}
}

// Router returns the HTTP router for MemLore REST.
func (h *Handlers) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.health)
	r.Route("/v1", func(r chi.Router) {
		r.Use(h.oidcMiddleware)
		r.Post("/lore-entries", h.createLoreEntry)
		r.Get("/lore-entries", h.listLoreEntries)
		r.Get("/lore-entries/{id}", h.getLoreEntry)
		r.Get("/lore-entries/{id}/explain", h.explainLoreEntry)
		r.Post("/lore-entries/{id}/verify", h.verifyLoreEntry)
		r.Post("/lore-entries/{id}/invalidate", h.invalidateLoreEntry)
		r.Post("/lore-entries/{id}/supersede", h.supersedeLoreEntry)
		r.Get("/lore-entries/{id}/audits", h.listLoreAudits)
		r.Post("/knowledge-search", h.knowledgeSearch)
		r.Post("/context/compile", h.compileContext)
		r.Post("/repository-profile", h.repositoryProfile)
		r.Post("/ingest/git", h.ingestGit)
		r.Post("/ingest/pr", h.ingestPR)
		r.Get("/ingest/runs", h.listIngestRuns)
		r.Get("/ingest/runs/{id}", h.getIngestRun)
		r.Get("/ingest/pr-runs", h.listPRIngestRuns)
		r.Get("/ingest/pr-runs/{id}", h.getPRIngestRun)
		r.Get("/ingest/candidates", h.listIngestCandidates)

		r.Post("/admin/teams", h.adminCreateTeam)
		r.Post("/admin/projects", h.adminCreateProject)
		r.Post("/admin/teams/{key}/members", h.adminAddTeamMember)
		r.Delete("/admin/teams/{key}/members/{subject}", h.adminRemoveTeamMember)
		r.Post("/admin/projects/{key}/members", h.adminAddProjectMember)
		r.Delete("/admin/projects/{key}/members/{subject}", h.adminRemoveProjectMember)
		r.Post("/admin/scope-bindings", h.adminBindScope)
		r.Delete("/admin/scope-bindings", h.adminUnbindScope)
	})
	return r
}

func (h *Handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "memlore",
		"version": h.Version,
	})
}

func (h *Handlers) createLoreEntry(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body createLoreRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	statement, scope, evidence, err := parseCreateRequest(body)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	entry, err := h.CreateLore.Handle(r.Context(), commands.CreateLoreCommand{
		Statement: statement,
		Scope:     scope,
		ActorID:   actor,
		Evidence:  evidence,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toLoreResponse(entry))
}

func (h *Handlers) getLoreEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := h.GetLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLoreResponse(entry))
}

func (h *Handlers) explainLoreEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.ExplainLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, result.Entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presenters.ToExplainResult(result.Entry, result.Audits, result.Evaluation))
}

func (h *Handlers) verifyLoreEntry(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermVerify)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	existing, err := h.GetLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, existing.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	entry, err := h.VerifyLore.Handle(r.Context(), commands.VerifyLoreCommand{
		EntryID: id,
		ActorID: actor,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLoreResponse(entry))
}

func (h *Handlers) invalidateLoreEntry(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermInvalidate)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	id := chi.URLParam(r, "id")
	existing, err := h.GetLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, existing.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	entry, err := h.InvalidateLore.Handle(r.Context(), commands.InvalidateLoreCommand{
		EntryID: id,
		ActorID: actor,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLoreResponse(entry))
}

func (h *Handlers) supersedeLoreEntry(w http.ResponseWriter, r *http.Request) {
	actor, err := h.actorFor(r, domain.PermWrite)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	var body supersedeLoreRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	evidence := make([]domain.EvidenceReference, 0, len(body.Evidence))
	for _, item := range body.Evidence {
		ref, refErr := domain.NewEvidenceReference(item.Type, item.Value)
		if refErr != nil {
			handleDomainError(w, refErr)
			return
		}
		evidence = append(evidence, ref)
	}
	id := chi.URLParam(r, "id")
	existing, err := h.GetLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, existing.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	entry, err := h.SupersedeLore.Handle(r.Context(), commands.SupersedeLoreCommand{
		EntryID:   id,
		Statement: body.Statement,
		ActorID:   actor,
		Evidence:  evidence,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toLoreResponse(entry))
}

func (h *Handlers) listLoreEntries(w http.ResponseWriter, r *http.Request) {
	kindRaw := strings.TrimSpace(r.URL.Query().Get("scope_kind"))
	key := strings.TrimSpace(r.URL.Query().Get("scope_key"))
	if kindRaw == "" || key == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "scope_kind and scope_key are required")
		return
	}
	kind, err := domain.ParseScopeKind(kindRaw)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	scope, err := domain.NewScope(kind, key)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, scope, false); err != nil {
		handleDomainError(w, err)
		return
	}
	includeStale := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_stale")), "true")
	items, err := h.ListLoreByScope.Handle(r.Context(), queries.ListLoreByScopeQuery{
		Scope:        scope,
		IncludeStale: includeStale,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := loreEntryListResponse{Items: make([]loreEntryResponse, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toLoreResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) listLoreAudits(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := h.GetLore.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if err := h.requireScopeAccess(r, entry.Scope, true); err != nil {
		handleDomainError(w, err)
		return
	}
	records, err := h.ListAudits.Handle(r.Context(), id)
	if err != nil {
		handleDomainError(w, err)
		return
	}
	resp := auditListResponse{Items: make([]auditRecordResponse, 0, len(records))}
	for _, record := range records {
		resp.Items = append(resp.Items, toAuditResponse(record))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) knowledgeSearch(w http.ResponseWriter, r *http.Request) {
	var body knowledgeSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	var scope *domain.Scope
	if body.Scope != nil {
		kind, err := domain.ParseScopeKind(string(body.Scope.Kind))
		if err != nil {
			handleDomainError(w, err)
			return
		}
		parsed, err := domain.NewScope(kind, body.Scope.Key)
		if err != nil {
			handleDomainError(w, err)
			return
		}
		scope = &parsed
		if err := h.requireScopeAccess(r, parsed, false); err != nil {
			handleDomainError(w, err)
			return
		}
	}
	result, err := h.SearchKnowledge.Handle(r.Context(), queries.SearchKnowledgeQuery{
		Query:        body.Query,
		Scope:        scope,
		Limit:        body.Limit,
		IncludeStale: body.IncludeStale,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	if p, ok := h.currentPrincipal(r); ok {
		filtered, ferr := h.gate().FilterAccessible(r.Context(), p, result.LoreEntries())
		if ferr != nil {
			handleDomainError(w, ferr)
			return
		}
		allowed := make(map[string]struct{}, len(filtered))
		for _, e := range filtered {
			allowed[e.ID] = struct{}{}
		}
		kept := make([]queries.GovernanceHit, 0, len(result.Governance))
		for _, hit := range result.Governance {
			if _, ok := allowed[hit.Entry.ID]; ok {
				kept = append(kept, hit)
			}
		}
		result.Governance = kept
	}
	writeJSON(w, http.StatusOK, presenters.ToKnowledgeSearchResult(result))
}

func (h *Handlers) compileContext(w http.ResponseWriter, r *http.Request) {
	var body compileContextRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	if body.Scope == nil {
		writeError(w, http.StatusBadRequest, "validation_error", "scope is required")
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
	result, err := h.CompileContext.Handle(r.Context(), queries.CompileContextQuery{
		Task:         body.Task,
		Query:        body.Query,
		Scope:        scope,
		TokenBudget:  body.TokenBudget,
		Branch:       body.Branch,
		Ticket:       body.Ticket,
		ChangedFiles: body.ChangedFiles,
		WorkingFiles: body.WorkingFiles,
		AgentID:      body.AgentID,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presenters.ToContextPacket(result))
}

func (h *Handlers) repositoryProfile(w http.ResponseWriter, r *http.Request) {
	var body repositoryProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "invalid JSON body")
		return
	}
	if body.Scope == nil {
		writeError(w, http.StatusBadRequest, "validation_error", "scope is required")
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
	result, err := h.RepositoryProfile.Handle(r.Context(), queries.RepositoryProfileQuery{
		Scope:       scope,
		TokenBudget: body.TokenBudget,
	})
	if err != nil {
		handleDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presenters.ToRepositoryProfile(result))
}

// Serve starts the HTTP server until context is cancelled.
func Serve(ctx context.Context, addr string, handler http.Handler) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
