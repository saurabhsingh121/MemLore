package domain_test

import (
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

// Characterization: tests/unit/domain/test_lore_enums.py

func TestScopeKindMatchesPythonValues(t *testing.T) {
	if domain.ScopeKindTeam != "team" {
		t.Fatalf("ScopeKindTeam = %q", domain.ScopeKindTeam)
	}
	if domain.ScopeKindRepository != "repository" {
		t.Fatalf("ScopeKindRepository = %q", domain.ScopeKindRepository)
	}
}

func TestEvidenceTypeMatchesPythonSet(t *testing.T) {
	want := map[domain.EvidenceType]bool{
		domain.EvidenceTypeURL:  true,
		domain.EvidenceTypePath: true,
		domain.EvidenceTypeADR:  true,
	}
	types := []domain.EvidenceType{
		domain.EvidenceTypeURL,
		domain.EvidenceTypePath,
		domain.EvidenceTypeADR,
	}
	if len(types) != len(want) {
		t.Fatalf("unexpected evidence type count: %d", len(types))
	}
	for _, et := range types {
		if !want[et] {
			t.Fatalf("unexpected evidence type %q", et)
		}
	}
}

func TestKnowledgeOriginReservesAgentValues(t *testing.T) {
	if domain.KnowledgeOriginHumanAuthored != "human_authored" {
		t.Fatal("human_authored mismatch")
	}
	if domain.KnowledgeOriginAgentInference != "agent_inference" {
		t.Fatal("agent_inference mismatch")
	}
}

func TestVerificationAndAuditEnumsMatchPython(t *testing.T) {
	if domain.VerificationUnverified != "unverified" {
		t.Fatal("unverified mismatch")
	}
	if domain.VerificationVerified != "verified" {
		t.Fatal("verified mismatch")
	}
	if domain.AuditActionCreate != "create" {
		t.Fatal("create mismatch")
	}
	if domain.AuditActionVerify != "verify" {
		t.Fatal("verify mismatch")
	}
}

func TestParseScopeKindRejectsUnknown(t *testing.T) {
	_, err := domain.ParseScopeKind("invalid")
	if err == nil {
		t.Fatal("expected error for invalid scope kind")
	}
}
