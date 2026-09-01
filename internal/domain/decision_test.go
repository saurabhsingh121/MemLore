package domain_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func TestNewHumanDecisionRequiresQuestionChoiceOwner(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	base := domain.NewDecisionInput{
		Scope:     scope,
		Question:  "How should payment events be published?",
		Choice:    "Transactional outbox",
		Owner:     "alice",
		CreatedBy: "alice",
		Now:       now,
	}
	if _, err := domain.NewHumanDecision(base); err != nil {
		t.Fatalf("valid: %v", err)
	}
	for _, in := range []domain.NewDecisionInput{
		{Scope: scope, Choice: "C", Owner: "alice", CreatedBy: "alice"},
		{Scope: scope, Question: "Q", Owner: "alice", CreatedBy: "alice"},
		{Scope: scope, Question: "Q", Choice: "C", CreatedBy: "alice"},
	} {
		if _, err := domain.NewHumanDecision(in); err == nil {
			t.Fatalf("expected validation error for %+v", in)
		}
	}
}

func TestNewHumanDecisionStoresAlternativesAndDefaultsDate(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	got, err := domain.NewHumanDecision(domain.NewDecisionInput{
		ID:       "dec-1",
		Scope:    scope,
		Question: "How should payment events be published?",
		Choice:   "Transactional outbox",
		Owner:    "alice",
		Alternatives: []domain.DecisionAlternative{
			{Label: " Dual-write ", Note: " Lost updates "},
		},
		AffectedComponents: []string{" payments-api "},
		CreatedBy:          "alice",
		Now:                now,
	})
	if err != nil {
		t.Fatalf("NewHumanDecision: %v", err)
	}
	if got.SourceKind != domain.DecisionSourceHuman || !got.Current {
		t.Fatalf("source/current = %s %v", got.SourceKind, got.Current)
	}
	if !got.DecidedAt.Equal(now) {
		t.Fatalf("decided_at default = %v", got.DecidedAt)
	}
	if len(got.Alternatives) != 1 || got.Alternatives[0].Label != "Dual-write" || got.Alternatives[0].Note != "Lost updates" {
		t.Fatalf("alternatives = %+v", got.Alternatives)
	}
	if len(got.AffectedComponents) != 1 || got.AffectedComponents[0] != "payments-api" {
		t.Fatalf("components = %+v", got.AffectedComponents)
	}
}

func TestNewHumanDecisionRejectsBlankAlternativeLabel(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := domain.NewHumanDecision(domain.NewDecisionInput{
		Scope:        scope,
		Question:     "Q",
		Choice:       "C",
		Owner:        "alice",
		CreatedBy:    "alice",
		Alternatives: []domain.DecisionAlternative{{Label: "  "}},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestNewHumanDecisionRejectsNonRepositoryScope(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "payments")
	_, err := domain.NewHumanDecision(domain.NewDecisionInput{
		Scope:     scope,
		Question:  "Q",
		Choice:    "C",
		Owner:     "alice",
		CreatedBy: "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestProjectADRDecisionFromArchitectureLore(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	entry, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		ID:        "adr-1",
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := domain.ProjectADRDecision(entry)
	if err != nil {
		t.Fatalf("ProjectADRDecision: %v", err)
	}
	if got.ID != "adr-1" || got.SourceKind != domain.DecisionSourceADR {
		t.Fatalf("id/source = %s %s", got.ID, got.SourceKind)
	}
	if got.Choice != entry.Statement || got.Owner != "ingest" || !got.Current {
		t.Fatalf("projection = %+v", got)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Type != domain.EvidenceTypeADR {
		t.Fatalf("evidence = %+v", got.Evidence)
	}
}

func TestProjectADRDecisionRejectsObservationalLore(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "We chose kafka in this commit",
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err != nil {
		t.Fatal(err)
	}
	if domain.CanProjectADRDecision(entry) {
		t.Fatal("observational lore must not project")
	}
	if _, err := domain.ProjectADRDecision(entry); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestWithSupersededByClearsCurrent(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	d, _ := domain.NewHumanDecision(domain.NewDecisionInput{
		ID: "a", Scope: scope, Question: "Q", Choice: "C", Owner: "alice", CreatedBy: "alice",
	})
	got := d.WithSupersededBy("b")
	if got.Current || got.SupersededByID == nil || *got.SupersededByID != "b" {
		t.Fatalf("superseded = %+v", got)
	}
	if d.Current == false {
		t.Fatal("original must remain current")
	}
}
