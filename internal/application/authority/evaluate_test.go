package authority_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/authority"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

func TestEvaluateGovernanceMapsEntry(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	entry := domain.LoreEntry{
		ID:                 "gov-1",
		Statement:          "Use outbox",
		Scope:              scope,
		Origin:             domain.KnowledgeOriginHumanAuthored,
		VerificationStatus: domain.VerificationVerified,
		Evidence:           []domain.EvidenceReference{adr},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	got := authority.EvaluateGovernance(entry, &scope, now)
	if got.Band != domain.TrustBandCanonical {
		t.Fatalf("band = %s", got.Band)
	}
	if got.Factors.SourceType != string(domain.SourceTypeADR) {
		t.Fatalf("source_type = %s", got.Factors.SourceType)
	}
	if got.Factors.ScopeMatch == nil || *got.Factors.ScopeMatch != 1.0 {
		t.Fatalf("scope_match = %v", got.Factors.ScopeMatch)
	}
}

func TestEvaluateGraphIsLow(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	got := authority.EvaluateGraph(ports.GraphFact{
		ID:    "f1",
		Score: 0.9,
		Scope: &ports.GraphScope{Kind: "repository", Key: "r1"},
	}, &scope, now)
	if got.Band != domain.TrustBandLow {
		t.Fatalf("band = %s", got.Band)
	}
	if got.Factors.GraphScore == nil || *got.Factors.GraphScore != 0.9 {
		t.Fatalf("graph_score = %v", got.Factors.GraphScore)
	}
	if got.Factors.SourceType != string(domain.SourceTypeGraph) {
		t.Fatalf("source_type = %s", got.Factors.SourceType)
	}
}
