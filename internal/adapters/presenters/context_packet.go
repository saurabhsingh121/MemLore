package presenters

import (
	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/queries"
)

// AuthorityFactors is the JSON authority factors shape.
type AuthorityFactors struct {
	VerificationStatus string   `json:"verification_status,omitempty"`
	Origin             string   `json:"origin,omitempty"`
	SupersessionStatus string   `json:"supersession_status,omitempty"`
	RecencyBoost       *float64 `json:"recency_boost,omitempty"`
	EvidenceStrength   *float64 `json:"evidence_strength,omitempty"`
	SourceType         string   `json:"source_type,omitempty"`
	ScopeMatch         *float64 `json:"scope_match,omitempty"`
	GraphScore         *float64 `json:"graph_score,omitempty"`
}

// ContextItem is a single compiled context entry.
type ContextItem struct {
	ID               string           `json:"id"`
	Statement        string           `json:"statement"`
	Source           string           `json:"source"`
	AuthorityScore   float64          `json:"authority_score"`
	TrustBand        string           `json:"trust_band"`
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

// ConflictGroup is disagreeing current lore in one scope.
type ConflictGroup struct {
	Scope      Scope    `json:"scope"`
	EntryIDs   []string `json:"entry_ids"`
	Statements []string `json:"statements"`
}

// ContextPacket is the compiled agent context response.
type ContextPacket struct {
	Task      string          `json:"task"`
	Query     string          `json:"query"`
	Scope     Scope           `json:"scope"`
	Items     []ContextItem   `json:"items"`
	Meta      ContextMeta     `json:"meta"`
	Warnings  []string        `json:"warnings"`
	Conflicts []ConflictGroup `json:"conflicts"`
}

// ToContextPacket maps compilation output to API response.
func ToContextPacket(result queries.CompileContextResult) ContextPacket {
	items := make([]ContextItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toContextItem(item))
	}
	warnings := result.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	return ContextPacket{
		Task:      result.Task,
		Query:     result.Query,
		Scope:     Scope{Kind: result.Scope.Kind, Key: result.Scope.Key},
		Items:     items,
		Meta: ContextMeta{
			TokenBudget:      result.Meta.TokenBudget,
			EstimatedTokens:  result.Meta.EstimatedTokens,
			ItemsIncluded:    result.Meta.ItemsIncluded,
			ItemsTotalRanked: result.Meta.ItemsTotalRanked,
		},
		Warnings:  warnings,
		Conflicts: toConflictGroups(result.Conflicts),
	}
}

func toContextItem(item appcontext.RankedItem) ContextItem {
	evidence := make([]Evidence, 0, len(item.Evidence))
	for _, ref := range item.Evidence {
		evidence = append(evidence, Evidence{Type: ref.Type, Value: ref.Value})
	}
	refs := item.ProvenanceRefs
	if refs == nil {
		refs = []string{}
	}
	return ContextItem{
		ID:             item.ID,
		Statement:      item.Statement,
		Source:         string(item.Source),
		AuthorityScore: item.AuthorityScore,
		TrustBand:      string(item.TrustBand),
		AuthorityFactors: AuthorityFactors{
			VerificationStatus: item.AuthorityFactors.VerificationStatus,
			Origin:             item.AuthorityFactors.Origin,
			SupersessionStatus: item.AuthorityFactors.SupersessionStatus,
			RecencyBoost:       item.AuthorityFactors.RecencyBoost,
			EvidenceStrength:   item.AuthorityFactors.EvidenceStrength,
			SourceType:         item.AuthorityFactors.SourceType,
			ScopeMatch:         item.AuthorityFactors.ScopeMatch,
			GraphScore:         item.AuthorityFactors.GraphScore,
		},
		Scope:          Scope{Kind: item.Scope.Kind, Key: item.Scope.Key},
		Evidence:       evidence,
		ProvenanceRefs: refs,
	}
}

func toConflictGroups(groups []appcontext.ConflictGroup) []ConflictGroup {
	conflicts := make([]ConflictGroup, 0, len(groups))
	for _, c := range groups {
		ids := c.EntryIDs
		if ids == nil {
			ids = []string{}
		}
		stmts := c.Statements
		if stmts == nil {
			stmts = []string{}
		}
		conflicts = append(conflicts, ConflictGroup{
			Scope:      Scope{Kind: c.Scope.Kind, Key: c.Scope.Key},
			EntryIDs:   ids,
			Statements: stmts,
		})
	}
	return conflicts
}
