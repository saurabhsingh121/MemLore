package mcpadapter

import (
	"context"

	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tools wires lore MCP tool handlers to the application layer.
type Tools struct {
	CreateLore      *commands.CreateLoreHandler
	VerifyLore      *commands.VerifyLoreHandler
	GetLore         *queries.GetLoreHandler
	ListLoreByScope *queries.ListLoreByScopeHandler
	ListAudits      *queries.ListAuditsHandler
	SearchKnowledge *queries.SearchKnowledgeHandler
	CompileContext  *queries.CompileContextHandler
}

// NewTools constructs MCP tool handlers from a unit-of-work factory and clock.
func NewTools(begin ports.UnitOfWorkFactory, clock ports.Clock, graph ports.KnowledgeGraph) *Tools {
	search := queries.NewSearchKnowledgeHandler(begin, graph, nil)
	return &Tools{
		CreateLore:      commands.NewCreateLoreHandler(begin, clock),
		VerifyLore:      commands.NewVerifyLoreHandler(begin, clock),
		GetLore:         queries.NewGetLoreHandler(begin),
		ListLoreByScope: queries.NewListLoreByScopeHandler(begin),
		ListAudits:      queries.NewListAuditsHandler(begin),
		SearchKnowledge: search,
		CompileContext:  queries.NewCompileContextHandler(search),
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
	Statement string          `json:"statement"`
	Scope     scopeInput      `json:"scope"`
	ActorID   string          `json:"actor_id"`
	Evidence  []evidenceInput `json:"evidence,omitempty"`
}

type getInput struct {
	ID string `json:"id"`
}

type verifyInput struct {
	ID      string `json:"id"`
	ActorID string `json:"actor_id"`
}

type explainInput struct {
	ID string `json:"id"`
}

type searchInput struct {
	Scope scopeInput `json:"scope"`
}

type knowledgeSearchInput struct {
	Query   string      `json:"query"`
	Scope   *scopeInput `json:"scope,omitempty"`
	Limit   *int        `json:"limit,omitempty"`
	ActorID string      `json:"actor_id"`
}

type getForTaskInput struct {
	Task        string     `json:"task"`
	Query       string     `json:"query,omitempty"`
	Scope       scopeInput `json:"scope"`
	TokenBudget *int       `json:"token_budget,omitempty"`
	ActorID     string     `json:"actor_id"`
}

func (t *Tools) remember(ctx context.Context, _ *sdkmcp.CallToolRequest, input rememberInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	actor, err := requireActor(input.ActorID)
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
		ActorID:   actor,
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
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) verify(ctx context.Context, _ *sdkmcp.CallToolRequest, input verifyInput) (*sdkmcp.CallToolResult, presenters.LoreEntry, error) {
	actor, err := requireActor(input.ActorID)
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	entry, err := t.VerifyLore.Handle(ctx, commands.VerifyLoreCommand{
		EntryID: input.ID,
		ActorID: actor,
	})
	if err != nil {
		return nil, presenters.LoreEntry{}, mapDomainError(err)
	}
	return nil, presenters.ToLoreEntry(entry), nil
}

func (t *Tools) explain(ctx context.Context, _ *sdkmcp.CallToolRequest, input explainInput) (*sdkmcp.CallToolResult, presenters.ExplainResult, error) {
	entry, err := t.GetLore.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.ExplainResult{}, mapDomainError(err)
	}
	audits, err := t.ListAudits.Handle(ctx, input.ID)
	if err != nil {
		return nil, presenters.ExplainResult{}, mapDomainError(err)
	}
	result := presenters.ExplainResult{
		LoreEntry: presenters.ToLoreEntry(entry),
		Audits:    make([]presenters.AuditRecord, 0, len(audits)),
	}
	for _, record := range audits {
		result.Audits = append(result.Audits, presenters.ToAuditRecord(record))
	}
	return nil, result, nil
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
	items, err := t.ListLoreByScope.Handle(ctx, scope)
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
	if _, err := requireActor(input.ActorID); err != nil {
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
	}
	result, err := t.SearchKnowledge.Handle(ctx, queries.SearchKnowledgeQuery{
		Query: input.Query,
		Scope: scope,
		Limit: derefLimit(input.Limit),
	})
	if err != nil {
		return nil, presenters.KnowledgeSearchResult{}, mapDomainError(err)
	}
	return nil, presenters.ToKnowledgeSearchResult(
		result.Query,
		result.Scope,
		result.Governance,
		result.Graph,
		result.Warnings,
	), nil
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
	if _, err := requireActor(input.ActorID); err != nil {
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
