package context_test

import (
	"testing"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

func TestRankVerifiedGovernanceAboveGraph(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	governance := []domain.LoreEntry{{
		ID:                 "gov-1",
		Statement:          "Verified rule",
		Scope:              scope,
		Origin:             domain.KnowledgeOriginHumanAuthored,
		VerificationStatus: domain.VerificationVerified,
		CreatedAt:          now,
		UpdatedAt:          now,
	}}
	graph := []ports.GraphFact{{
		ID:        "fact-1",
		Statement: "Graph fact",
		Score:     0.99,
	}}

	items := appcontext.RankAndDedup(governance, graph, now)
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
	if items[0].Source != appcontext.ItemSourceGovernance {
		t.Fatalf("first source = %s", items[0].Source)
	}
	if items[0].AuthorityScore <= items[1].AuthorityScore {
		t.Fatalf("scores = %f %f", items[0].AuthorityScore, items[1].AuthorityScore)
	}
}

func TestDedupSkipsGraphMatchingGovernanceStatement(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	stmt := "Use outbox for payments."
	governance := []domain.LoreEntry{{
		ID:        "gov-1",
		Statement: stmt,
		Scope:     scope,
		Origin:    domain.KnowledgeOriginHumanAuthored,
		CreatedAt: now,
		UpdatedAt: now,
	}}
	graph := []ports.GraphFact{{
		ID:        "fact-1",
		Statement: "  USE OUTBOX FOR PAYMENTS. ",
		Score:     0.9,
	}}

	items := appcontext.RankAndDedup(governance, graph, now)
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
}

func TestApplyTokenBudgetCapsItems(t *testing.T) {
	items := []appcontext.RankedItem{
		{Statement: stringsRepeat('a', 400), EstimatedTokens: 120},
		{Statement: stringsRepeat('b', 400), EstimatedTokens: 120},
		{Statement: stringsRepeat('c', 400), EstimatedTokens: 120},
	}
	selected, used := appcontext.ApplyTokenBudget(items, 200)
	if len(selected) != 1 {
		t.Fatalf("selected = %d", len(selected))
	}
	if used != 120 {
		t.Fatalf("used = %d", used)
	}
}

func stringsRepeat(ch rune, n int) string {
	out := make([]rune, n)
	for i := range out {
		out[i] = ch
	}
	return string(out)
}
