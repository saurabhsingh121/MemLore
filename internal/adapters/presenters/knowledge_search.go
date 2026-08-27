package presenters

import (
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// GraphFact is a knowledge-plane search hit in API responses.
type GraphFact struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Score          float64  `json:"score"`
	Scope          *Scope   `json:"scope,omitempty"`
	ProvenanceRefs []string `json:"provenance_refs"`
}

// GraphFactList wraps graph search hits.
type GraphFactList struct {
	Items []GraphFact `json:"items"`
}

// KnowledgeSearchResult is the dual-plane search response shared by REST and MCP.
type KnowledgeSearchResult struct {
	Query      string        `json:"query"`
	Scope      *Scope        `json:"scope"`
	Governance LoreEntryList `json:"governance"`
	Graph      GraphFactList `json:"graph"`
	Warnings   []string      `json:"warnings"`
}

// ToKnowledgeSearchResult maps orchestration output to the API response.
func ToKnowledgeSearchResult(query string, scope *domain.Scope, governance []domain.LoreEntry, graph []ports.GraphFact, warnings []string) KnowledgeSearchResult {
	govItems := make([]LoreEntry, 0, len(governance))
	for _, entry := range governance {
		govItems = append(govItems, ToLoreEntry(entry))
	}
	graphItems := make([]GraphFact, 0, len(graph))
	for _, fact := range graph {
		graphItems = append(graphItems, ToGraphFact(fact))
	}
	if warnings == nil {
		warnings = []string{}
	}
	var scopeDTO *Scope
	if scope != nil {
		scopeDTO = &Scope{Kind: scope.Kind, Key: scope.Key}
	}
	return KnowledgeSearchResult{
		Query:      query,
		Scope:      scopeDTO,
		Governance: LoreEntryList{Items: govItems},
		Graph:      GraphFactList{Items: graphItems},
		Warnings:   warnings,
	}
}

// ToGraphFact maps a port graph fact to its API representation.
func ToGraphFact(fact ports.GraphFact) GraphFact {
	var scope *Scope
	if fact.Scope != nil {
		scope = &Scope{Kind: domain.ScopeKind(fact.Scope.Kind), Key: fact.Scope.Key}
	}
	refs := fact.ProvenanceRefs
	if refs == nil {
		refs = []string{}
	}
	return GraphFact{
		ID:             fact.ID,
		Statement:      fact.Statement,
		Score:          fact.Score,
		Scope:          scope,
		ProvenanceRefs: refs,
	}
}
