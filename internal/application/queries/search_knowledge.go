package queries

import (
	"context"
	"log/slog"
	"sort"
	"strings"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"golang.org/x/sync/errgroup"
)

const defaultSearchLimit = 10

const warningGraphServiceUnavailable = "graph_service_unavailable"

// SearchKnowledgeQuery is input for dual-plane knowledge search.
type SearchKnowledgeQuery struct {
	Query        string
	Scope        *domain.Scope
	Limit        int
	IncludeStale bool
}

// GraphReceipt is a collapsed graph fact attached to a governance hit.
type GraphReceipt struct {
	ID             string
	Statement      string
	Score          float64
	ProvenanceRefs []string
}

// GovernanceHit is a query-relevant lore entry with optional graph receipt.
type GovernanceHit struct {
	Entry   domain.LoreEntry
	Receipt *GraphReceipt
}

// SearchKnowledgeResult is the merged orchestration output.
type SearchKnowledgeResult struct {
	Query      string
	Scope      *domain.Scope
	Governance []GovernanceHit
	Graph      []ports.GraphFact
	Warnings   []string
}

// LoreEntries returns governance entries without receipts (compile/authz helpers).
func (r SearchKnowledgeResult) LoreEntries() []domain.LoreEntry {
	out := make([]domain.LoreEntry, 0, len(r.Governance))
	for _, hit := range r.Governance {
		out = append(out, hit.Entry)
	}
	return out
}

// HitsFromEntries wraps lore entries as governance hits without receipts.
func HitsFromEntries(entries []domain.LoreEntry) []GovernanceHit {
	out := make([]GovernanceHit, 0, len(entries))
	for _, e := range entries {
		out = append(out, GovernanceHit{Entry: e})
	}
	return out
}

// SearchKnowledgeHandler orchestrates governance relevance search and graph search.
type SearchKnowledgeHandler struct {
	begin  ports.UnitOfWorkFactory
	graph  ports.KnowledgeGraph
	logger *slog.Logger
}

// NewSearchKnowledgeHandler wires the orchestrator with injected dependencies.
func NewSearchKnowledgeHandler(begin ports.UnitOfWorkFactory, graph ports.KnowledgeGraph, logger *slog.Logger) *SearchKnowledgeHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &SearchKnowledgeHandler{
		begin:  begin,
		graph:  graph,
		logger: logger,
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

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		uow, err := h.begin(gctx)
		if err != nil {
			return err
		}
		defer uow.Rollback(gctx)
		items, err := uow.LoreEntries().SearchRelevant(gctx, ports.SearchRelevantOpts{
			Query: trimmed,
			Scope: query.Scope,
			Limit: limit * 5,
		})
		if err != nil {
			return err
		}
		if !query.IncludeStale {
			items = appcontext.FilterCurrent(items)
		}
		domain.SortLoreByRelevance(items, trimmed)
		if len(items) > limit {
			items = items[:limit]
		}
		governance = items
		return nil
	})

	g.Go(func() error {
		var graphScope *ports.GraphScope
		if query.Scope != nil {
			graphScope = &ports.GraphScope{
				Kind: string(query.Scope.Kind),
				Key:  query.Scope.Key,
			}
		}
		facts, err := h.graph.Search(gctx, ports.SearchRequest{
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

	hits, remainingGraph, err := h.mergePlanes(ctx, trimmed, query.IncludeStale, governance, graphFacts, limit)
	if err != nil {
		return SearchKnowledgeResult{}, err
	}

	result.Governance = hits
	if result.Governance == nil {
		result.Governance = []GovernanceHit{}
	}
	result.Graph = remainingGraph
	if result.Graph == nil {
		result.Graph = []ports.GraphFact{}
	}
	result.Warnings = warnings
	if result.Warnings == nil {
		result.Warnings = []string{}
	}
	return result, nil
}

func (h *SearchKnowledgeHandler) mergePlanes(
	ctx context.Context,
	query string,
	includeStale bool,
	governance []domain.LoreEntry,
	graphFacts []ports.GraphFact,
	limit int,
) ([]GovernanceHit, []ports.GraphFact, error) {
	byID := make(map[string]int, len(governance))
	hits := make([]GovernanceHit, 0, len(governance))
	for _, entry := range governance {
		byID[entry.ID] = len(hits)
		hits = append(hits, GovernanceHit{Entry: entry})
	}

	remaining := make([]ports.GraphFact, 0, len(graphFacts))
	for _, fact := range graphFacts {
		loreID := linkedLoreID(fact.ProvenanceRefs)
		if loreID == "" {
			remaining = append(remaining, fact)
			continue
		}
		if idx, ok := byID[loreID]; ok {
			if hits[idx].Receipt == nil || fact.Score > hits[idx].Receipt.Score {
				hits[idx].Receipt = receiptFromFact(fact)
			}
			continue
		}
		// Hydrate accessible lore missed by text match.
		entry, err := h.getLore(ctx, loreID)
		if err != nil {
			if _, ok := err.(*domain.NotFoundError); ok {
				// Treat as graph-only if lore gone.
				remaining = append(remaining, fact)
				continue
			}
			return nil, nil, err
		}
		if !includeStale && !domain.IsCurrent(entry) {
			remaining = append(remaining, fact)
			continue
		}
		byID[entry.ID] = len(hits)
		hits = append(hits, GovernanceHit{Entry: entry, Receipt: receiptFromFact(fact)})
	}

	// Re-rank hits (hydrated may not match query text — keep them; prefer receipt-backed).
	sort.SliceStable(hits, func(i, j int) bool {
		si := domain.RelevanceScore(hits[i].Entry.Statement, query)
		sj := domain.RelevanceScore(hits[j].Entry.Statement, query)
		if hits[i].Receipt != nil {
			si += 50
		}
		if hits[j].Receipt != nil {
			sj += 50
		}
		if si != sj {
			return si > sj
		}
		vi := hits[i].Entry.VerificationStatus == domain.VerificationVerified
		vj := hits[j].Entry.VerificationStatus == domain.VerificationVerified
		if vi != vj {
			return vi && !vj
		}
		return hits[i].Entry.CreatedAt.After(hits[j].Entry.CreatedAt)
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, remaining, nil
}

func (h *SearchKnowledgeHandler) getLore(ctx context.Context, id string) (domain.LoreEntry, error) {
	uow, err := h.begin(ctx)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	defer uow.Rollback(ctx)
	return uow.LoreEntries().Get(ctx, id)
}

func linkedLoreID(refs []string) string {
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			return ref
		}
	}
	return ""
}

func receiptFromFact(fact ports.GraphFact) *GraphReceipt {
	refs := fact.ProvenanceRefs
	if refs == nil {
		refs = []string{}
	}
	return &GraphReceipt{
		ID:             fact.ID,
		Statement:      fact.Statement,
		Score:          fact.Score,
		ProvenanceRefs: refs,
	}
}
