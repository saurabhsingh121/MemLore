package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// Characterization: tests/unit/domain/test_lore_entry.py

func TestLoreEntryDefaultsToUnverifiedHumanAuthored(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindRepository, "r1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
	if entry.VerifiedBy != nil {
		t.Fatal("expected verified_by nil")
	}
	if entry.InvalidatedBy != nil || entry.InvalidatedAt != nil {
		t.Fatal("expected invalidate fields nil")
	}
	if entry.SupersededByID != nil {
		t.Fatal("expected superseded_by_id nil")
	}
	if domain.IsSuperseded(entry) {
		t.Fatal("new entry must not be superseded")
	}
	if entry.ID == "" {
		t.Fatal("expected non-empty id")
	}
}

func TestLoreEntryRejectsOversizedStatement(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindTeam, "t1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	_, err = domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: strings.Repeat("x", domain.MaxStatementLength+1),
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestObservationalLoreEntryIsUnverifiedRepositoryObservation(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindRepository, "r1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	ev, err := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox because dual-writes race.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err != nil {
		t.Fatalf("NewObservationalLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
	if entry.VerifiedBy != nil || entry.VerifiedAt != nil {
		t.Fatal("expected unverified fields nil")
	}
}

func TestObservationalLoreEntryRequiresCommitEvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "observational lore requires commit or pr evidence" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestObservationalLoreEntryAcceptsPREvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, err := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox because dual-writes race.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err != nil {
		t.Fatalf("NewObservationalLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
}

func TestObservationalLoreEntryRejectsHumanOrigin(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	_, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginHumanAuthored,
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestArchitectureDecisionLoreEntryIsVerifiedWithADREvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, err := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	entry, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewArchitectureDecisionLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginArchitectureDecision {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
	if entry.VerifiedBy == nil || *entry.VerifiedBy != "alice" {
		t.Fatalf("verified_by = %v", entry.VerifiedBy)
	}
	if entry.VerifiedAt == nil || !entry.VerifiedAt.Equal(now) {
		t.Fatalf("verified_at = %v", entry.VerifiedAt)
	}
}

func TestArchitectureDecisionLoreEntryRequiresADREvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "architecture decision lore requires adr evidence" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArchitectureDecisionLoreEntryRejectsObservationalOrigin(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	_, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginRepositoryObservation,
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestObservationalLoreEntryRejectsADREvidenceAlone(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	_, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestHumanVerifiedLoreEntryIsVerifiedWithEvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, err := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	entry, err := domain.NewHumanVerifiedLoreEntry(domain.NewLoreEntryInput{
		Statement: "Payment events use transactional outbox.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewHumanVerifiedLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginHumanVerified {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
	if entry.VerifiedBy == nil || *entry.VerifiedBy != "alice" {
		t.Fatalf("verified_by = %v", entry.VerifiedBy)
	}
	if entry.VerifiedAt == nil || !entry.VerifiedAt.Equal(now) {
		t.Fatalf("verified_at = %v", entry.VerifiedAt)
	}
	if len(entry.Evidence) != 1 || entry.Evidence[0].Type != domain.EvidenceTypePR {
		t.Fatalf("evidence = %+v", entry.Evidence)
	}
}

func TestHumanVerifiedLoreEntryRequiresEvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewHumanVerifiedLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "human verified lore requires evidence" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHumanVerifiedLoreEntryRejectsObservationalOrigin(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	_, err := domain.NewHumanVerifiedLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginRepositoryObservation,
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestVerifiedHumanAuthoredLoreEntryIsVerified(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	entry, err := domain.NewVerifiedHumanAuthoredLoreEntry(domain.NewLoreEntryInput{
		Statement: "Payment events MUST use the transactional outbox.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewVerifiedHumanAuthoredLoreEntry: %v", err)
	}
	if entry.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatalf("origin = %q", entry.Origin)
	}
	if entry.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", entry.VerificationStatus)
	}
	if entry.VerifiedBy == nil || *entry.VerifiedBy != "alice" {
		t.Fatalf("verified_by = %v", entry.VerifiedBy)
	}
}

func TestVerifiedHumanAuthoredLoreEntryRequiresEvidence(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewVerifiedHumanAuthoredLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoreEntryRejectsHumanVerifiedOriginOnCreate(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginHumanVerified,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "create origin must be human_authored" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoreEntryRejectsArchitectureDecisionOriginOnCreate(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginArchitectureDecision,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "create origin must be human_authored" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoreEntryRejectsNonHumanOriginOnCreate(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindTeam, "t1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	_, err = domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		Origin:    domain.KnowledgeOriginAgentInference,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "create origin must be human_authored" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsCurrentExcludesSupersededAndInvalidated(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	current, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "current",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}
	if !domain.IsCurrent(current) {
		t.Fatal("expected current")
	}

	succ := "successor-id"
	superseded := current
	superseded.SupersededByID = &succ
	if domain.IsCurrent(superseded) {
		t.Fatal("superseded must not be current")
	}

	invalidated := current
	invalidated.VerificationStatus = domain.VerificationInvalidated
	if domain.IsCurrent(invalidated) {
		t.Fatal("invalidated must not be current")
	}
}
