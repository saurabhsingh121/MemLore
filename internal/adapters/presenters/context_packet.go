package presenters

import (
	"github.com/memlore/memlore/internal/application/queries"
)

// AuthorityFactors is the JSON authority factors shape.
type AuthorityFactors struct {
	VerificationStatus string   `json:"verification_status,omitempty"`
	Origin             string   `json:"origin,omitempty"`
	RecencyBoost       *float64 `json:"recency_boost,omitempty"`
	GraphScore         *float64 `json:"graph_score,omitempty"`
}

// ContextItem is a single compiled context entry.
type ContextItem struct {
	ID               string           `json:"id"`
	Statement        string           `json:"statement"`
	Source           string           `json:"source"`
	AuthorityScore   float64          `json:"authority_score"`
	AuthorityFactors AuthorityFactors `json:"authority_factors"`
	Scope            Scope            `json:"scope"`
	Evidence         []Evidence       `json:"evidence"`
	ProvenanceRefs   []string         `json:"provenance_refs"`
}

// ContextMeta summarizes token budgeting.
type ContextMeta struct {
	TokenBudget      int `json:"token_budget"`
	EstimatedTokens  int `json:"estimated_tokens"`
	ItemsIncluded    int `json:"items_included"`
	ItemsTotalRanked int `json:"items_total_ranked"`
}

// ContextPacket is the compiled agent context response.
type ContextPacket struct {
	Task     string        `json:"task"`
	Query    string        `json:"query"`
	Scope    Scope         `json:"scope"`
	Items    []ContextItem `json:"items"`
	Meta     ContextMeta   `json:"meta"`
	Warnings []string      `json:"warnings"`
}

// ToContextPacket maps compilation output to API response.
func ToContextPacket(result queries.CompileContextResult) ContextPacket {
	items := make([]ContextItem, 0, len(result.Items))
	for _, item := range result.Items {
		evidence := make([]Evidence, 0, len(item.Evidence))
		for _, ref := range item.Evidence {
			evidence = append(evidence, Evidence{Type: ref.Type, Value: ref.Value})
		}
		refs := item.ProvenanceRefs
		if refs == nil {
			refs = []string{}
		}
		items = append(items, ContextItem{
			ID:             item.ID,
			Statement:      item.Statement,
			Source:         string(item.Source),
			AuthorityScore: item.AuthorityScore,
			AuthorityFactors: AuthorityFactors{
				VerificationStatus: item.AuthorityFactors.VerificationStatus,
				Origin:             item.AuthorityFactors.Origin,
				RecencyBoost:       item.AuthorityFactors.RecencyBoost,
				GraphScore:         item.AuthorityFactors.GraphScore,
			},
			Scope: Scope{
				Kind: item.Scope.Kind,
				Key:  item.Scope.Key,
			},
			Evidence:       evidence,
			ProvenanceRefs: refs,
		})
	}
	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	return ContextPacket{
		Task:  result.Task,
		Query: result.Query,
		Scope: Scope{
			Kind: result.Scope.Kind,
			Key:  result.Scope.Key,
		},
		Items: items,
		Meta: ContextMeta{
			TokenBudget:      result.Meta.TokenBudget,
			EstimatedTokens:  result.Meta.EstimatedTokens,
			ItemsIncluded:    result.Meta.ItemsIncluded,
			ItemsTotalRanked: result.Meta.ItemsTotalRanked,
		},
		Warnings: warnings,
	}
}
