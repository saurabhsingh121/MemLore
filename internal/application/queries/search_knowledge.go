package queries

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"golang.org/x/sync/errgroup"
)

const defaultSearchLimit = 10

const warningGraphServiceUnavailable = "graph_service_unavailable"

// SearchKnowledgeQuery is input for dual-plane knowledge search.
type SearchKnowledgeQuery struct {
	Query string
	Scope *domain.Scope
	Limit int
}

// SearchKnowledgeResult is the merged orchestration output.
type SearchKnowledgeResult struct {
	Query      string
	Scope      *domain.Scope
	Governance []domain.LoreEntry
	Graph      []ports.GraphFact
	Warnings   []string
}

// SearchKnowledgeHandler orchestrates governance scope list and graph search.
type SearchKnowledgeHandler struct {
	listByScope *ListLoreByScopeHandler
	graph       ports.KnowledgeGraph
	logger      *slog.Logger
}

// NewSearchKnowledgeHandler wires the orchestrator with injected dependencies.
func NewSearchKnowledgeHandler(begin ports.UnitOfWorkFactory, graph ports.KnowledgeGraph, logger *slog.Logger) *SearchKnowledgeHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &SearchKnowledgeHandler{
		listByScope: NewListLoreByScopeHandler(begin),
		graph:       graph,
		logger:      logger,
	}
}

// Handle runs parallel governance and graph retrieval and merges results.
func (h *SearchKnowledgeHandler) Handle(ctx context.Context, query SearchKnowledgeQuery) (SearchKnowledgeResult, error) {
	trimmed := strings.TrimSpace(query.Query)
	if trimmed == "" {
		return SearchKnowledgeResult{}, &domain.ValidationError{Message: "query is required"}
	}

	limit := query.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	result := SearchKnowledgeResult{
		Query: trimmed,
		Scope: query.Scope,
	}

	var (
		governance []domain.LoreEntry
		graphFacts []ports.GraphFact
		warnings   []string
	)

	g, ctx := errgroup.WithContext(ctx)

	if query.Scope != nil {
		scope := *query.Scope
		g.Go(func() error {
			items, err := h.listByScope.Handle(ctx, scope)
			if err != nil {
				return err
			}
			governance = items
			return nil
		})
	}

	g.Go(func() error {
		var graphScope *ports.GraphScope
		if query.Scope != nil {
			graphScope = &ports.GraphScope{
				Kind: string(query.Scope.Kind),
				Key:  query.Scope.Key,
			}
		}
		facts, err := h.graph.Search(ctx, ports.SearchRequest{
			Query: trimmed,
			Scope: graphScope,
			Limit: limit,
		})
		if err != nil {
			h.logger.Warn("graph search failed", "error", err)
			warnings = append(warnings, warningGraphServiceUnavailable)
			return nil
		}
		graphFacts = facts
		return nil
	})

	if err := g.Wait(); err != nil {
		return SearchKnowledgeResult{}, err
	}

	sort.Slice(graphFacts, func(i, j int) bool {
		return graphFacts[i].Score > graphFacts[j].Score
	})

	result.Governance = governance
	if result.Governance == nil {
		result.Governance = []domain.LoreEntry{}
	}
	result.Graph = graphFacts
	if result.Graph == nil {
		result.Graph = []ports.GraphFact{}
	}
	result.Warnings = warnings
	if result.Warnings == nil {
		result.Warnings = []string{}
	}
	return result, nil
}
