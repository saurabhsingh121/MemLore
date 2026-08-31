package queries

import (
	"context"
	"strings"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/domain"
)

const defaultCompileRetrievalLimit = appcontext.DefaultRetrievalLimit

// CompileContextQuery is input for context compilation.
type CompileContextQuery struct {
	Task        string
	Query       string
	Scope       domain.Scope
	TokenBudget int
}

// CompileContextResult is the compiled context packet.
type CompileContextResult struct {
	Task      string
	Query     string
	Scope     domain.Scope
	Items     []appcontext.RankedItem
	Meta      appcontext.Meta
	Warnings  []string
	Conflicts []appcontext.ConflictGroup
}

type knowledgeSearcher interface {
	Handle(ctx context.Context, query SearchKnowledgeQuery) (SearchKnowledgeResult, error)
}

// CompileContextHandler compiles a token-budgeted context packet for agents.
type CompileContextHandler struct {
	search knowledgeSearcher
	now    func() time.Time
}

// NewCompileContextHandler wires the compiler with a search handler.
func NewCompileContextHandler(search knowledgeSearcher) *CompileContextHandler {
	return &CompileContextHandler{
		search: search,
		now:    time.Now,
	}
}

// Handle retrieves knowledge and compiles ranked, budgeted context.
func (h *CompileContextHandler) Handle(ctx context.Context, query CompileContextQuery) (CompileContextResult, error) {
	task := strings.TrimSpace(query.Task)
	if task == "" {
		return CompileContextResult{}, &domain.ValidationError{Message: "task is required"}
	}
	searchQuery := strings.TrimSpace(query.Query)
	if searchQuery == "" {
		searchQuery = task
	}
	if strings.TrimSpace(query.Scope.Key) == "" {
		return CompileContextResult{}, &domain.ValidationError{Message: "scope is required"}
	}

	budget := query.TokenBudget
	if budget <= 0 {
		budget = appcontext.DefaultTokenBudget
	}

	scope := query.Scope
	searchResult, err := h.search.Handle(ctx, SearchKnowledgeQuery{
		Query:        searchQuery,
		Scope:        &scope,
		Limit:        defaultCompileRetrievalLimit,
		IncludeStale: false,
	})
	if err != nil {
		return CompileContextResult{}, err
	}

	// retrieve → temporal filter → conflict detect → evaluate+rank → budget.
	current := appcontext.FilterCurrent(searchResult.Governance)
	conflicts := appcontext.DetectConflicts(current)

	now := h.now()
	ranked := appcontext.RankAndDedup(current, searchResult.Graph, scope, now)
	selected, used := appcontext.ApplyTokenBudget(ranked, budget)

	warnings := searchResult.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	if conflicts == nil {
		conflicts = []appcontext.ConflictGroup{}
	}

	return CompileContextResult{
		Task:  task,
		Query: searchQuery,
		Scope: scope,
		Items: selected,
		Meta: appcontext.Meta{
			TokenBudget:      budget,
			EstimatedTokens:  used,
			ItemsIncluded:    len(selected),
			ItemsTotalRanked: len(ranked),
		},
		Warnings:  warnings,
		Conflicts: conflicts,
	}, nil
}
