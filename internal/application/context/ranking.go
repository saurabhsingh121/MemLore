package context

import (
	"sort"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/application/authority"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

const (
	DefaultTokenBudget    = 4096
	ItemTokenOverhead     = 80
	CharsPerToken         = 4
	DefaultRetrievalLimit = 20
)

// ItemSource identifies which plane produced a context item.
type ItemSource string

const (
	ItemSourceGovernance ItemSource = "governance"
	ItemSourceGraph      ItemSource = "graph"
)

// AuthorityFactors explain how authority_score was derived.
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

// RankedItem is a scored candidate before token budgeting.
type RankedItem struct {
	ID               string
	Statement        string
	Source           ItemSource
	AuthorityScore   float64
	TrustBand        domain.TrustBand
	AuthorityFactors AuthorityFactors
	Scope            domain.Scope
	Evidence         []domain.EvidenceReference
	ProvenanceRefs   []string
	EstimatedTokens  int
}

// Meta summarizes compilation output.
type Meta struct {
	TokenBudget       int
	EstimatedTokens   int
	ItemsIncluded     int
	ItemsTotalRanked  int
	UnclassifiedCount int
}

// NormalizeStatement lowercases and trims for dedup comparison.
func NormalizeStatement(statement string) string {
	return strings.ToLower(strings.TrimSpace(statement))
}

// EstimateTokens approximates token count for budgeting.
func EstimateTokens(statement string) int {
	length := len(statement) + ItemTokenOverhead
	tokens := length / CharsPerToken
	if tokens < 1 {
		return 1
	}
	return tokens
}

func factorsFrom(eval domain.Evaluation) AuthorityFactors {
	return AuthorityFactors{
		VerificationStatus: eval.Factors.VerificationStatus,
		Origin:             eval.Factors.Origin,
		SupersessionStatus: eval.Factors.SupersessionStatus,
		RecencyBoost:       eval.Factors.RecencyBoost,
		EvidenceStrength:   eval.Factors.EvidenceStrength,
		SourceType:         eval.Factors.SourceType,
		ScopeMatch:         eval.Factors.ScopeMatch,
		GraphScore:         eval.Factors.GraphScore,
	}
}

// RankAndDedup merges search results into ranked items with graph dedup.
func RankAndDedup(governance []domain.LoreEntry, graph []ports.GraphFact, requested domain.Scope, now time.Time) []RankedItem {
	govStatements := make(map[string]struct{}, len(governance))
	items := make([]RankedItem, 0, len(governance)+len(graph))
	req := &requested

	for _, entry := range governance {
		eval := authority.EvaluateGovernance(entry, req, now)
		items = append(items, RankedItem{
			ID:               entry.ID,
			Statement:        entry.Statement,
			Source:           ItemSourceGovernance,
			AuthorityScore:   eval.Score,
			TrustBand:        eval.Band,
			AuthorityFactors: factorsFrom(eval),
			Scope:            entry.Scope,
			Evidence:         entry.Evidence,
			ProvenanceRefs:   []string{},
			EstimatedTokens:  EstimateTokens(entry.Statement),
		})
		govStatements[NormalizeStatement(entry.Statement)] = struct{}{}
	}

	for _, fact := range graph {
		if _, dup := govStatements[NormalizeStatement(fact.Statement)]; dup {
			continue
		}
		scope := domain.Scope{}
		if fact.Scope != nil {
			kind, err := domain.ParseScopeKind(fact.Scope.Kind)
			if err == nil {
				scope, _ = domain.NewScope(kind, fact.Scope.Key)
			}
		}
		eval := authority.EvaluateGraph(fact, req, now)
		refs := fact.ProvenanceRefs
		if refs == nil {
			refs = []string{}
		}
		items = append(items, RankedItem{
			ID:               fact.ID,
			Statement:        fact.Statement,
			Source:           ItemSourceGraph,
			AuthorityScore:   eval.Score,
			TrustBand:        eval.Band,
			AuthorityFactors: factorsFrom(eval),
			Scope:            scope,
			Evidence:         nil,
			ProvenanceRefs:   refs,
			EstimatedTokens:  EstimateTokens(fact.Statement),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].AuthorityScore == items[j].AuthorityScore {
			return items[i].Statement < items[j].Statement
		}
		return items[i].AuthorityScore > items[j].AuthorityScore
	})
	return items
}

// ApplyTokenBudget returns items that fit within the token budget.
func ApplyTokenBudget(items []RankedItem, budget int) ([]RankedItem, int) {
	if budget <= 0 {
		budget = DefaultTokenBudget
	}
	selected := make([]RankedItem, 0, len(items))
	used := 0
	for _, item := range items {
		if used+item.EstimatedTokens > budget {
			continue
		}
		selected = append(selected, item)
		used += item.EstimatedTokens
	}
	return selected, used
}
