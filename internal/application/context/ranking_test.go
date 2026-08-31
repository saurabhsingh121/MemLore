package context_test

import (
	"testing"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

func rankingScope(t *testing.T) domain.Scope {
	t.Helper()
	scope, err := domain.NewScope(domain.ScopeKindRepository, "r1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	return scope
}

func TestRankVerifiedGovernanceAboveGraph(t *testing.T) {
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	scope := rankingScope(t)
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

	items := appcontext.RankAndDedup(governance, graph, scope, now)
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
	if items[0].Source != appcontext.ItemSourceGovernance {
		t.Fatalf("first source = %s", items[0].Source)
	}
	if items[0].AuthorityScore <= items[1].AuthorityScore {
		t.Fatalf("scores = %f %f", items[0].AuthorityScore, items[1].AuthorityScore)
	}
	if items[0].TrustBand != domain.TrustBandHigh {
		t.Fatalf("verified no-evidence band = %s", items[0].TrustBand)
	}
	if items[1].TrustBand != domain.TrustBandLow {
		t.Fatalf("graph band = %s", items[1].TrustBand)
	}
}

func TestRankCanonicalADRAboveVerifiedAndUnverified(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope := rankingScope(t)
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	governance := []domain.LoreEntry{
		{
			ID: "unv", Statement: "Unverified human", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "ver", Statement: "Verified human", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "adr", Statement: "Verified ADR", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
			Evidence: []domain.EvidenceReference{adr}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "agent", Statement: "Agent guess", Scope: scope,
			Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationUnverified,
			CreatedAt: now, UpdatedAt: now,
		},
	}
	graph := []ports.GraphFact{{ID: "g1", Statement: "Graph only", Score: 0.99}}

	items := appcontext.RankAndDedup(governance, graph, scope, now)
	want := []string{"adr", "ver", "unv", "g1", "agent"}
	if len(items) != len(want) {
		t.Fatalf("items = %+v", items)
	}
	for i, id := range want {
		if items[i].ID != id {
			t.Fatalf("rank[%d] = %s want %s (full=%+v)", i, items[i].ID, id, ids(items))
		}
	}
	if items[0].TrustBand != domain.TrustBandCanonical {
		t.Fatalf("adr band = %s", items[0].TrustBand)
	}
	if items[0].AuthorityFactors.SourceType != string(domain.SourceTypeADR) {
		t.Fatalf("adr source_type = %s", items[0].AuthorityFactors.SourceType)
	}
	if items[0].AuthorityFactors.EvidenceStrength == nil || *items[0].AuthorityFactors.EvidenceStrength != 1.0 {
		t.Fatalf("adr evidence = %v", items[0].AuthorityFactors.EvidenceStrength)
	}
	if items[2].TrustBand != domain.TrustBandMedium {
		t.Fatalf("unverified band = %s", items[2].TrustBand)
	}
	if items[3].AuthorityFactors.GraphScore == nil {
		t.Fatal("graph item missing graph_score")
	}
	if items[4].TrustBand != domain.TrustBandLow {
		t.Fatalf("agent band = %s", items[4].TrustBand)
	}
}

func TestRankVerifiedAgentBelowHumanADR(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope := rankingScope(t)
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	items := appcontext.RankAndDedup([]domain.LoreEntry{
		{
			ID: "agent", Statement: "Verified agent", Scope: scope,
			Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationVerified,
			Evidence: []domain.EvidenceReference{adr}, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "human", Statement: "Verified human ADR", Scope: scope,
			Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
			Evidence: []domain.EvidenceReference{adr}, CreatedAt: now, UpdatedAt: now,
		},
	}, nil, scope, now)
	if items[0].ID != "human" {
		t.Fatalf("expected human first, got %+v", ids(items))
	}
	if items[0].TrustBand != domain.TrustBandCanonical || items[1].TrustBand != domain.TrustBandHigh {
		t.Fatalf("bands = %s %s", items[0].TrustBand, items[1].TrustBand)
	}
}

func TestDedupSkipsGraphMatchingGovernanceStatement(t *testing.T) {
	now := time.Now().UTC()
	scope := rankingScope(t)
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

	items := appcontext.RankAndDedup(governance, graph, scope, now)
	if len(items) != 1 {
		t.Fatalf("items = %+v", items)
	}
}

func TestInvalidatedDoesNotOutrankUnverified(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope := rankingScope(t)
	governance := []domain.LoreEntry{
		{
			ID:                 "inv",
			Statement:          "Invalidated rule",
			Scope:              scope,
			Origin:             domain.KnowledgeOriginHumanAuthored,
			VerificationStatus: domain.VerificationInvalidated,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		{
			ID:                 "unv",
			Statement:          "Unverified rule",
			Scope:              scope,
			Origin:             domain.KnowledgeOriginHumanAuthored,
			VerificationStatus: domain.VerificationUnverified,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}
	items := appcontext.RankAndDedup(governance, nil, scope, now)
	if len(items) != 2 {
		t.Fatalf("items = %d", len(items))
	}
	if items[0].ID != "unv" {
		t.Fatalf("expected unverified first, got %+v", items)
	}
	if items[0].AuthorityScore <= items[1].AuthorityScore {
		t.Fatalf("scores = %f %f", items[0].AuthorityScore, items[1].AuthorityScore)
	}
	if items[1].TrustBand != domain.TrustBandUntrusted {
		t.Fatalf("invalidated band = %s", items[1].TrustBand)
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

func ids(items []appcontext.RankedItem) []string {
	out := make([]string, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

func stringsRepeat(ch rune, n int) string {
	out := make([]rune, n)
	for i := range out {
		out[i] = ch
	}
	return string(out)
}
