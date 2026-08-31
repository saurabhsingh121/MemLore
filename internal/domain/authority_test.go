package domain_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func TestEvaluateAuthorityBandMatrix(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-dual-plane")
	urlEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeURL, "https://example.test/adr")
	req := &scope

	cases := []struct {
		name string
		in   domain.FactorInputs
		band domain.TrustBand
	}{
		{
			name: "verified ADR human current is canonical",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
				CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
				EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandCanonical,
		},
		{
			name: "verified architecture_decision current is canonical",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginArchitectureDecision, VerificationStatus: domain.VerificationVerified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandCanonical,
		},
		{
			name: "verified human no evidence is high",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandHigh,
		},
		{
			name: "unverified human is medium",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandMedium,
		},
		{
			name: "unverified agent inference is low",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationUnverified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandLow,
		},
		{
			name: "unverified agent observation is low",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginAgentObservation, VerificationStatus: domain.VerificationUnverified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandLow,
		},
		{
			name: "verified agent inference is high not canonical",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationVerified,
				CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
				EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandHigh,
		},
		{
			name: "graph is low",
			in:   domain.FactorInputs{FromGraph: true, GraphScore: 0.99, Now: now, RequestedScope: req},
			band: domain.TrustBandLow,
		},
		{
			name: "invalidated is untrusted",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationInvalidated,
				CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
				EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandUntrusted,
		},
		{
			name: "superseded verified ADR is medium not canonical",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
				Superseded: true, CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
				EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandMedium,
		},
		{
			name: "unverified repo observation is low",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginRepositoryObservation, VerificationStatus: domain.VerificationUnverified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandLow,
		},
		{
			name: "unverified imported is medium",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginImportedSource, VerificationStatus: domain.VerificationUnverified,
				CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandMedium,
		},
		{
			name: "verified with url evidence only is high",
			in: domain.FactorInputs{
				Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
				CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{urlEv},
				EntryScope: scope, RequestedScope: req,
			},
			band: domain.TrustBandHigh,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.EvaluateAuthority(tc.in)
			if got.Band != tc.band {
				t.Fatalf("band = %s want %s score=%.4f factors=%+v", got.Band, tc.band, got.Score, got.Factors)
			}
			if len(got.Breakdown) == 0 {
				t.Fatal("expected factor breakdown")
			}
		})
	}
}

func TestEvaluateAuthorityOrdering(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	req := &scope

	canonical := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
		CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
		EntryScope: scope, RequestedScope: req,
	})
	high := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationVerified,
		CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
	})
	medium := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
		CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
	})
	graph := domain.EvaluateAuthority(domain.FactorInputs{FromGraph: true, GraphScore: 0.99, Now: now})
	agent := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationUnverified,
		CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: req,
	})
	invalidated := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationInvalidated,
		CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
		EntryScope: scope, RequestedScope: req,
	})
	verifiedAgent := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: domain.VerificationVerified,
		CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
		EntryScope: scope, RequestedScope: req,
	})

	order := []struct {
		name  string
		score float64
	}{
		{"canonical", canonical.Score},
		{"high", high.Score},
		{"medium", medium.Score},
		{"graph", graph.Score},
		{"agent", agent.Score},
		{"invalidated", invalidated.Score},
	}
	for i := 0; i < len(order)-1; i++ {
		if order[i].score <= order[i+1].score {
			t.Fatalf("%s (%.4f) should outrank %s (%.4f)", order[i].name, order[i].score, order[i+1].name, order[i+1].score)
		}
	}
	if canonical.Score <= verifiedAgent.Score {
		t.Fatalf("human ADR %.4f should outrank verified agent %.4f", canonical.Score, verifiedAgent.Score)
	}
	if invalidated.Score > 0.20 {
		t.Fatalf("invalidated cap = %.4f", invalidated.Score)
	}
	if graph.Score > 0.45 {
		t.Fatalf("graph cap = %.4f", graph.Score)
	}
}

func TestEvaluateAuthorityExplainScopeIsExact(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	got := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
		CreatedAt: now, Now: now, EntryScope: scope, RequestedScope: nil,
	})
	if got.Factors.ScopeMatch == nil || *got.Factors.ScopeMatch != 1.0 {
		t.Fatalf("explain scope_match = %v want 1.0", got.Factors.ScopeMatch)
	}
}

func TestEvaluateAuthorityScopeMismatch(t *testing.T) {
	now := time.Now().UTC()
	entry, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	other, _ := domain.NewScope(domain.ScopeKindRepository, "r2")
	got := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, VerificationStatus: domain.VerificationUnverified,
		CreatedAt: now, Now: now, EntryScope: entry, RequestedScope: &other,
	})
	if got.Factors.ScopeMatch == nil || *got.Factors.ScopeMatch != 0.5 {
		t.Fatalf("kind-only scope_match = %v want 0.5", got.Factors.ScopeMatch)
	}
}

func TestEvaluateAuthorityEvidenceStrength(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	urlEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeURL, "https://example.test")

	none := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, Now: now,
		EntryScope: scope, RequestedScope: &scope,
	})
	other := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, Now: now,
		Evidence: []domain.EvidenceReference{urlEv}, EntryScope: scope, RequestedScope: &scope,
	})
	strong := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, Now: now,
		Evidence: []domain.EvidenceReference{adr}, EntryScope: scope, RequestedScope: &scope,
	})
	if *none.Factors.EvidenceStrength != 0 || *other.Factors.EvidenceStrength != 0.6 || *strong.Factors.EvidenceStrength != 1.0 {
		t.Fatalf("strength none=%v other=%v strong=%v", *none.Factors.EvidenceStrength, *other.Factors.EvidenceStrength, *strong.Factors.EvidenceStrength)
	}
	if none.Factors.SourceType != string(domain.SourceTypeHumanStatement) {
		t.Fatalf("source_type none = %s", none.Factors.SourceType)
	}
	if strong.Factors.SourceType != string(domain.SourceTypeADR) {
		t.Fatalf("source_type adr = %s", strong.Factors.SourceType)
	}
}

func TestEvaluateAuthorityRecencyDecays(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	fresh := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now, Now: now,
		EntryScope: scope, RequestedScope: &scope,
	})
	old := domain.EvaluateAuthority(domain.FactorInputs{
		Origin: domain.KnowledgeOriginHumanAuthored, CreatedAt: now.Add(-400 * 24 * time.Hour), Now: now,
		EntryScope: scope, RequestedScope: &scope,
	})
	if *fresh.Factors.RecencyBoost != 0.10 {
		t.Fatalf("fresh recency = %v", *fresh.Factors.RecencyBoost)
	}
	if *old.Factors.RecencyBoost != 0 {
		t.Fatalf("old recency = %v", *old.Factors.RecencyBoost)
	}
	if fresh.Score <= old.Score {
		t.Fatalf("fresh %.4f should beat old %.4f", fresh.Score, old.Score)
	}
}

func TestAgentInferenceNeverCanonical(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	for _, verified := range []domain.VerificationStatus{domain.VerificationUnverified, domain.VerificationVerified} {
		got := domain.EvaluateAuthority(domain.FactorInputs{
			Origin: domain.KnowledgeOriginAgentInference, VerificationStatus: verified,
			CreatedAt: now, Now: now, Evidence: []domain.EvidenceReference{adr},
			EntryScope: scope, RequestedScope: &scope,
		})
		if got.Band == domain.TrustBandCanonical {
			t.Fatalf("agent inference verified=%s reached canonical", verified)
		}
	}
}
