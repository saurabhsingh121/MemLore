package domain_test

import (
	"errors"
	"strings"
	"testing"

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
	if !errors.As(err, &ve) || ve.Message != "observational lore requires commit evidence" {
		t.Fatalf("unexpected error: %v", err)
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
