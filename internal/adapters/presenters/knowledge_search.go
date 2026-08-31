package presenters

import (
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
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

// GraphReceipt is a collapsed graph fact attached to a governance hit.
type GraphReceipt struct {
	ID             string   `json:"id"`
	Statement      string   `json:"statement"`
	Score          float64  `json:"score"`
	ProvenanceRefs []string `json:"provenance_refs"`
}

// KnowledgeGovernanceItem is a lore entry plus optional graph receipt.
type KnowledgeGovernanceItem struct {
	LoreEntry
	GraphReceipt *GraphReceipt `json:"graph_receipt,omitempty"`
}

// GraphFactList wraps graph search hits.
type GraphFactList struct {
	Items []GraphFact `json:"items"`
}

// KnowledgeGovernanceList wraps governance hits for knowledge search.
type KnowledgeGovernanceList struct {
	Items []KnowledgeGovernanceItem `json:"items"`
}

// KnowledgeSearchResult is the dual-plane search response shared by REST and MCP.
type KnowledgeSearchResult struct {
	Query      string                   `json:"query"`
	Scope      *Scope                   `json:"scope"`
	Governance KnowledgeGovernanceList  `json:"governance"`
	Graph      GraphFactList            `json:"graph"`
	Warnings   []string                 `json:"warnings"`
}

// ToKnowledgeSearchResult maps orchestration output to the API response.
func ToKnowledgeSearchResult(result queries.SearchKnowledgeResult) KnowledgeSearchResult {
	govItems := make([]KnowledgeGovernanceItem, 0, len(result.Governance))
	for _, hit := range result.Governance {
		item := KnowledgeGovernanceItem{LoreEntry: ToLoreEntry(hit.Entry)}
		if hit.Receipt != nil {
			refs := hit.Receipt.ProvenanceRefs
			if refs == nil {
				refs = []string{}
			}
			item.GraphReceipt = &GraphReceipt{
				ID:             hit.Receipt.ID,
				Statement:      hit.Receipt.Statement,
				Score:          hit.Receipt.Score,
				ProvenanceRefs: refs,
			}
		}
		govItems = append(govItems, item)
	}
	graphItems := make([]GraphFact, 0, len(result.Graph))
	for _, fact := range result.Graph {
		graphItems = append(graphItems, ToGraphFact(fact))
	}
	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	var scopeDTO *Scope
	if result.Scope != nil {
		scopeDTO = &Scope{Kind: result.Scope.Kind, Key: result.Scope.Key}
	}
	return KnowledgeSearchResult{
		Query:      result.Query,
		Scope:      scopeDTO,
		Governance: KnowledgeGovernanceList{Items: govItems},
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
