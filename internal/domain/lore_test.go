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
