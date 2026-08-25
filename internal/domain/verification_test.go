package domain_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// Characterization: src/memlore/domain/services/verification.py
// and tests/unit/application/test_verify_lore.py

func TestApplyVerificationSetsVerifiedAndAudit(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindRepository, "r1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		ID:        "entry-1",
		Now:       time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}

	now := time.Date(2026, 8, 25, 13, 0, 0, 0, time.UTC)
	updated, audit, err := domain.ApplyVerification(entry, "alice", now)
	if err != nil {
		t.Fatalf("ApplyVerification: %v", err)
	}
	if audit == nil {
		t.Fatal("expected verify audit")
	}
	if updated.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", updated.VerificationStatus)
	}
	if updated.VerifiedBy == nil || *updated.VerifiedBy != "alice" {
		t.Fatal("expected verified_by alice")
	}
	if updated.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatal("origin must remain human_authored")
	}
	if audit.Action != domain.AuditActionVerify {
		t.Fatalf("audit action = %q", audit.Action)
	}
}

func TestApplyVerificationIsIdempotent(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entry, _ := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
		ID:        "entry-1",
	})
	now := time.Date(2026, 8, 25, 13, 0, 0, 0, time.UTC)
	verified, firstAudit, err := domain.ApplyVerification(entry, "alice", now)
	if err != nil || firstAudit == nil {
		t.Fatalf("first verify: audit=%v err=%v", firstAudit, err)
	}

	second, secondAudit, err := domain.ApplyVerification(verified, "bob", now)
	if err != nil {
		t.Fatalf("second verify: %v", err)
	}
	if secondAudit != nil {
		t.Fatal("expected no audit on idempotent verify")
	}
	if second.VerifiedBy == nil || *second.VerifiedBy != "alice" {
		t.Fatal("verified_by must remain alice after idempotent verify")
	}
}

func TestApplyVerificationRejectsBlankActor(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entry, _ := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	_, _, err := domain.ApplyVerification(entry, "  ", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
}
