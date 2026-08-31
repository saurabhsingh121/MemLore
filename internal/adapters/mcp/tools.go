package mcpadapter

import (
	"context"

	"github.com/memlore/memlore/internal/adapters/presenters"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tools wires lore MCP tool handlers to the application layer.
type Tools struct {
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
	Auth              *appauth.Service
	Authz             *authz.Gate
	Membership        ports.MembershipDirectory
}

// NewTools constructs MCP tool handlers from a unit-of-work factory and clock.
func NewTools(begin ports.UnitOfWorkFactory, clock ports.Clock, graph ports.KnowledgeGraph) *Tools {
	search := queries.NewSearchKnowledgeHandler(begin, graph, nil)
	list := queries.NewListLoreByScopeHandler(begin)
	return &Tools{
		CreateLore:        commands.NewCreateLoreHandler(begin, clock),
		VerifyLore:        commands.NewVerifyLoreHandler(begin, clock),
		InvalidateLore:    commands.NewInvalidateLoreHandler(begin, clock),
		SupersedeLore:     commands.NewSupersedeLoreHandler(begin, clock),
		GetLore:           queries.NewGetLoreHandler(begin),
		ListLoreByScope:   list,
		ListAudits:        queries.NewListAuditsHandler(begin),
		SearchKnowledge:   search,
		CompileContext:    queries.NewCompileContextHandler(search),
		RepositoryProfile: queries.NewRepositoryProfileHandler(list, search),
		ExplainLore:       queries.NewExplainLoreHandler(begin),
		Auth:              appauth.NewService(appauth.Config{}, nil),
	}
}

type scopeInput struct {
	Kind domain.ScopeKind `json:"kind"`
	Key  string           `json:"key"`
}

type evidenceInput struct {
	Type  domain.EvidenceType `json:"type"`
	Value string              `json:"value"`
}

type rememberInput struct {
	Statement   string          `json:"statement"`
	Scope       scopeInput      `json:"scope"`
	ActorID     string          `json:"actor_id"`
	AccessToken string          `json:"access_token,omitempty"`
	Evidence    []evidenceInput `json:"evidence,omitempty"`
}

type getInput struct {
	ID          string `json:"id"`
	AccessToken string `json:"access_token,omitempty"`
}

type verifyInput struct {
	ID          string `json:"id"`
	ActorID     string `json:"actor_id"`
	AccessToken string `json:"access_token,omitempty"`
}

type invalidateInput struct {
	ID          string `json:"id"`
	ActorID     string `json:"actor_id"`
	AccessToken string `json:"access_token,omitempty"`
}

type supersedeInput struct {
	ID          string          `json:"id"`
	Statement   string          `json:"statement"`
	ActorID     string          `json:"actor_id"`
	AccessToken string          `json:"access_token,omitempty"`
	Evidence    []evidenceInput `json:"evidence,omitempty"`
}

type explainInput struct {
	ID          string `json:"id"`
	AccessToken string `json:"access_token,omitempty"`
}

type searchInput struct {
	Scope        scopeInput `json:"scope"`
	IncludeStale bool       `json:"include_stale,omitempty"`
	AccessToken  string     `json:"access_token,omitempty"`
}

type knowledgeSearchInput struct {
	Query        string      `json:"query"`
	Scope        *scopeInput `json:"scope,omitempty"`
	Limit        *int        `json:"limit,omitempty"`
	IncludeStale bool        `json:"include_stale,omitempty"`
	ActorID      string      `json:"actor_id"`
	AccessToken  string      `json:"access_token,omitempty"`
}

type getForTaskInput struct {
	Task        string     `json:"task"`
	Query       string     `json:"query,omitempty"`
	Scope       scopeInput `json:"scope"`
	TokenBudget *int       `json:"token_budget,omitempty"`
	ActorID     string     `json:"actor_id"`
	AccessToken string     `json:"access_token,omitempty"`
}

type repoProfileInput struct {
	Scope       scopeInput `json:"scope"`
	TokenBudget *int       `json:"token_budget,omitempty"`
	ActorID     string     `json:"actor_id"`
	AccessToken string     `json:"access_token,omitempty"`
}

func (t *Tools) remember(ctx context.Context, _ *sdkmcp.CallToolRequest, input rememberInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	p, err := t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermWrite)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	kind, err := domain.ParseScopeKind(string(input.Scope.Kind))
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	scope, err := domain.NewScope(kind, input.Scope.Key)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	if err := t.requireScope(ctx, p, scope, false); err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	evidence := make([]domain.EvidenceReference, 0, len(input.Evidence))
	for _, item := range input.Evidence {
		ref, refErr := domain.NewEvidenceReference(item.Type, item.Value)
		if refErr != nil {
			return nil, presenters.LoreEntry{}, mapDomainError(refErr)
		}
		evidence = append(evidence, ref)
	}
	entry, err := t.CreateLore.Handle(ctx, commands.CreateLoreCommand{
		Statement: input.Statement,
		Scope:     scope,
		ActorID:   p.Subject,
		Evidence:  evidence,
	})
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) get(ctx context.Context, _ *sdkmcp.CallToolRequest, input getInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	entry, err := t.GetLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	if t.authEnabled() {
		p, err := t.resolvePrincipal(ctx, "", input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.LoreEntry{}, mapDomainError(err)
		}
		if err := t.requireScope(ctx, p, entry.Scope, true); err != nil {
			return nil, presenters.LoreEntry{}, mapDomainError(err)
		}
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) verify(ctx context.Context, _ *sdkmcp.CallToolRequest, input verifyInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	p, err := t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermVerify)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	existing, err := t.GetLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	if err := t.requireScope(ctx, p, existing.Scope, true); err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	entry, err := t.VerifyLore.Handle(ctx, commands.VerifyLoreCommand{
		EntryID: input.ID,
		ActorID: p.Subject,
	})
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) invalidate(ctx context.Context, _ *sdkmcp.CallToolRequest, input invalidateInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	p, err := t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermInvalidate)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	existing, err := t.GetLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	if err := t.requireScope(ctx, p, existing.Scope, true); err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	entry, err := t.InvalidateLore.Handle(ctx, commands.InvalidateLoreCommand{
		EntryID: input.ID,
		ActorID: p.Subject,
	})
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) supersede(ctx context.Context, _ *sdkmcp.CallToolRequest, input supersedeInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	p, err := t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermWrite)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	existing, err := t.GetLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	if err := t.requireScope(ctx, p, existing.Scope, true); err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	evidence := make([]domain.EvidenceReference, 0, len(input.Evidence))
	for _, item := range input.Evidence {
		ref, refErr := domain.NewEvidenceReference(item.Type, item.Value)
		if refErr != nil {
			return nil, presenters.LoreEntry{}, mapDomainError(refErr)
		}
		evidence = append(evidence, ref)
	}
	entry, err := t.SupersedeLore.Handle(ctx, commands.SupersedeLoreCommand{
		EntryID:   input.ID,
		Statement: input.Statement,
		ActorID:   p.Subject,
		Evidence:  evidence,
	})
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) explain(ctx context.Context, _ *sdkmcp.CallToolRequest, input explainInput) (*sdkmcp.CallToolResult, presenters.ExplainResult, error) {
	result, err := t.ExplainLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.ExplainResult{}, mapDomainError(err)
	}
	if t.authEnabled() {
		p, err := t.resolvePrincipal(ctx, "", input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.ExplainResult{}, mapDomainError(err)
		}
		if err := t.requireScope(ctx, p, result.Entry.Scope, true); err != nil {
			return nil, presenters.ExplainResult{}, mapDomainError(err)
		}
	}
	return nil, presenters.ToExplainResult(result.Entry, result.Audits, result.Evaluation), nil
}

func (t *Tools) search(ctx context.Context, _ *sdkmcp.CallToolRequest, input searchInput) (*sdkmcp.CallToolResult, presenters.LoreEntryList, error) {
	kind, err := domain.ParseScopeKind(string(input.Scope.Kind))
	if err != nil {
		return nil, presenters.LoreEntryList{}, mapDomainError(err)
	}
	scope, err := domain.NewScope(kind, input.Scope.Key)
	if err != nil {
		return nil, presenters.LoreEntryList{}, mapDomainError(err)
	}
	if t.authEnabled() {
		p, err := t.resolvePrincipal(ctx, "", input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.LoreEntryList{}, mapDomainError(err)
		}
		if err := t.requireScope(ctx, p, scope, false); err != nil {
			return nil, presenters.LoreEntryList{}, mapDomainError(err)
		}
	}
	items, err := t.ListLoreByScope.Handle(ctx, queries.ListLoreByScopeQuery{
		Scope:        scope,
		IncludeStale: input.IncludeStale,
	})
	if err != nil {
		return nil, presenters.LoreEntryList{}, mapDomainError(err)
	}
	resp := presenters.LoreEntryList{Items: make([]presenters.LoreEntry, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, presenters.ToLoreEntry(item))
	}
	return nil, resp, nil
}

func (t *Tools) knowledgeSearch(ctx context.Context, _ *sdkmcp.CallToolRequest, input knowledgeSearchInput) (*sdkmcp.CallToolResult, presenters.KnowledgeSearchResult, error) {
	var p domain.Principal
	if t.authEnabled() {
		var err error
		p, err = t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
		}
	} else if _, err := requireActor(input.ActorID); err != nil {
		return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
	}
	var scope *domain.Scope
	if input.Scope != nil {
		kind, err := domain.ParseScopeKind(string(input.Scope.Kind))
		if err != nil {
			return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
		}
		parsed, err := domain.NewScope(kind, input.Scope.Key)
		if err != nil {
			return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
		}
		scope = &parsed
		if t.authEnabled() {
			if err := t.requireScope(ctx, p, parsed, false); err != nil {
				return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
			}
		}
	}
	result, err := t.SearchKnowledge.Handle(ctx, queries.SearchKnowledgeQuery{
		Query:        input.Query,
		Scope:        scope,
		Limit:        derefLimit(input.Limit),
		IncludeStale: input.IncludeStale,
	})
	if err != nil {
		return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
	}
	if t.authEnabled() {
		filtered, ferr := t.gate().FilterAccessible(ctx, p, result.LoreEntries())
		if ferr != nil {
			return nil, presenters.KnowledgeSearchResult{}, mapDomainError(ferr)
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
	return nil, presenters.ToKnowledgeSearchResult(result), nil
}

func derefLimit(limit *int) int {
	if limit == nil {
		return 0
	}
	return *limit
}

func derefTokenBudget(budget *int) int {
	if budget == nil {
		return 0
	}
	return *budget
}

func (t *Tools) getForTask(ctx context.Context, _ *sdkmcp.CallToolRequest, input getForTaskInput) (*sdkmcp.CallToolResult, presenters.ContextPacket, error) {
	var p domain.Principal
	if t.authEnabled() {
		var err error
		p, err = t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.ContextPacket{}, mapDomainError(err)
		}
	} else if _, err := requireActor(input.ActorID); err != nil {
		return nil, presenters.ContextPacket{}, mapDomainError(err)
	}
	kind, err := domain.ParseScopeKind(string(input.Scope.Kind))
	if err != nil {
		return nil, presenters.ContextPacket{}, mapDomainError(err)
	}
	scope, err := domain.NewScope(kind, input.Scope.Key)
	if err != nil {
		return nil, presenters.ContextPacket{}, mapDomainError(err)
	}
	if t.authEnabled() {
		if err := t.requireScope(ctx, p, scope, false); err != nil {
			return nil, presenters.ContextPacket{}, mapDomainError(err)
		}
	}
	result, err := t.CompileContext.Handle(ctx, queries.CompileContextQuery{
		Task:        input.Task,
		Query:       input.Query,
		Scope:       scope,
		TokenBudget: derefTokenBudget(input.TokenBudget),
	})
	if err != nil {
		return nil, presenters.ContextPacket{}, mapDomainError(err)
	}
	return nil, presenters.ToContextPacket(result), nil
}

func (t *Tools) repoProfile(ctx context.Context, _ *sdkmcp.CallToolRequest, input repoProfileInput) (*sdkmcp.CallToolResult, presenters.RepositoryProfile, error) {
	var p domain.Principal
	if t.authEnabled() {
		var err error
		p, err = t.resolvePrincipal(ctx, input.ActorID, input.AccessToken, domain.PermRead)
		if err != nil {
			return nil, presenters.RepositoryProfile{}, mapDomainError(err)
		}
	} else if _, err := requireActor(input.ActorID); err != nil {
		return nil, presenters.RepositoryProfile{}, mapDomainError(err)
	}
	kind, err := domain.ParseScopeKind(string(input.Scope.Kind))
	if err != nil {
		return nil, presenters.RepositoryProfile{}, mapDomainError(err)
	}
	scope, err := domain.NewScope(kind, input.Scope.Key)
	if err != nil {
		return nil, presenters.RepositoryProfile{}, mapDomainError(err)
	}
	if t.authEnabled() {
		if err := t.requireScope(ctx, p, scope, false); err != nil {
			return nil, presenters.RepositoryProfile{}, mapDomainError(err)
		}
	}
	result, err := t.RepositoryProfile.Handle(ctx, queries.RepositoryProfileQuery{
		Scope:       scope,
		TokenBudget: derefTokenBudget(input.TokenBudget),
	})
	if err != nil {
		return nil, presenters.RepositoryProfile{}, mapDomainError(err)
	}
	return nil, presenters.ToRepositoryProfile(result), nil
}
