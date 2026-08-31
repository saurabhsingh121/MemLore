package context

import (
	"sort"
	"strings"
	"time"

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
	GraphScore         *float64 `json:"graph_score,omitempty"`
}

// RankedItem is a scored candidate before token budgeting.
type RankedItem struct {
	ID               string
	Statement        string
	Source           ItemSource
	AuthorityScore   float64
	AuthorityFactors AuthorityFactors
	Scope            domain.Scope
	Evidence         []domain.EvidenceReference
	ProvenanceRefs   []string
	EstimatedTokens  int
}

// Meta summarizes compilation output.
type Meta struct {
	TokenBudget      int
	EstimatedTokens  int
	ItemsIncluded    int
	ItemsTotalRanked int
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

func recencyBoost(createdAt time.Time, now time.Time) float64 {
	age := now.Sub(createdAt)
	if age < 0 {
		age = 0
	}
	days := age.Hours() / 24
	fraction := days / 365
	if fraction > 1 {
		fraction = 1
	}
	return 0.10 * (1 - fraction)
}

func governanceScore(entry domain.LoreEntry, now time.Time) (float64, AuthorityFactors) {
	boost := recencyBoost(entry.CreatedAt, now)
	factors := AuthorityFactors{
		VerificationStatus: string(entry.VerificationStatus),
		Origin:             string(entry.Origin),
		RecencyBoost:       &boost,
	}
	base := 0.55
	switch entry.VerificationStatus {
	case domain.VerificationVerified:
		base = 0.85
	case domain.VerificationInvalidated:
		// Defense: invalidated must never outrank unverified if it reaches ranking.
		base = 0.20
	}
	if !domain.IsSuperseded(entry) && entry.VerificationStatus != domain.VerificationInvalidated {
		factors.SupersessionStatus = "current"
	} else if domain.IsSuperseded(entry) {
		factors.SupersessionStatus = "superseded"
	}
	return base + boost, factors
}

func graphScore(fact ports.GraphFact) (float64, AuthorityFactors) {
	score := fact.Score * 0.80
	if score > 0.80 {
		score = 0.80
	}
	factors := AuthorityFactors{GraphScore: &fact.Score}
	return score, factors
}

// RankAndDedup merges search results into ranked items with graph dedup.
func RankAndDedup(governance []domain.LoreEntry, graph []ports.GraphFact, now time.Time) []RankedItem {
	govStatements := make(map[string]struct{}, len(governance))
	items := make([]RankedItem, 0, len(governance)+len(graph))

	for _, entry := range governance {
		score, factors := governanceScore(entry, now)
		items = append(items, RankedItem{
			ID:               entry.ID,
			Statement:        entry.Statement,
			Source:           ItemSourceGovernance,
			AuthorityScore:   score,
			AuthorityFactors: factors,
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
		score, factors := graphScore(fact)
		refs := fact.ProvenanceRefs
		if refs == nil {
			refs = []string{}
		}
		items = append(items, RankedItem{
			ID:               fact.ID,
			Statement:        fact.Statement,
			Source:           ItemSourceGraph,
			AuthorityScore:   score,
			AuthorityFactors: factors,
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
