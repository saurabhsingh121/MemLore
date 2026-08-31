package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func testEntry(t *testing.T, id string) domain.LoreEntry {
	t.Helper()
	scope, err := domain.NewScope(domain.ScopeKindRepository, "r1")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox",
		Scope:     scope,
		CreatedBy: "alice",
		ID:        id,
		Now:       time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}
	return entry
}

func TestApplyInvalidationSetsStatusActorTimeAndAudit(t *testing.T) {
	entry := testEntry(t, "entry-1")
	now := time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC)

	updated, audit, err := domain.ApplyInvalidation(entry, "alice", now)
	if err != nil {
		t.Fatalf("ApplyInvalidation: %v", err)
	}
	if audit == nil {
		t.Fatal("expected invalidate audit")
	}
	if updated.VerificationStatus != domain.VerificationInvalidated {
		t.Fatalf("status = %q", updated.VerificationStatus)
	}
	if updated.InvalidatedBy == nil || *updated.InvalidatedBy != "alice" {
		t.Fatal("expected invalidated_by alice")
	}
	if updated.InvalidatedAt == nil || !updated.InvalidatedAt.Equal(now) {
		t.Fatal("expected invalidated_at")
	}
	if updated.Statement != entry.Statement || updated.Origin != entry.Origin {
		t.Fatal("statement and origin must be preserved")
	}
	if audit.Action != domain.AuditActionInvalidate {
		t.Fatalf("audit action = %q", audit.Action)
	}
	if audit.TargetID != entry.ID {
		t.Fatalf("audit target = %q", audit.TargetID)
	}
}

func TestApplyInvalidationIsIdempotent(t *testing.T) {
	entry := testEntry(t, "entry-1")
	now := time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC)
	invalidated, first, err := domain.ApplyInvalidation(entry, "alice", now)
	if err != nil || first == nil {
		t.Fatalf("first invalidate: audit=%v err=%v", first, err)
	}

	second, secondAudit, err := domain.ApplyInvalidation(invalidated, "bob", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("second invalidate: %v", err)
	}
	if secondAudit != nil {
		t.Fatal("expected no audit on idempotent invalidate")
	}
	if second.InvalidatedBy == nil || *second.InvalidatedBy != "alice" {
		t.Fatal("invalidated_by must remain alice")
	}
}

func TestApplyInvalidationRejectsBlankActor(t *testing.T) {
	entry := testEntry(t, "entry-1")
	_, _, err := domain.ApplyInvalidation(entry, "  ", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyInvalidationRejectsSuperseded(t *testing.T) {
	entry := testEntry(t, "entry-1")
	successor := "successor-1"
	entry.SupersededByID = &successor
	_, _, err := domain.ApplyInvalidation(entry, "alice", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplySupersessionCreatesSuccessorAndLink(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	now := time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC)
	evidence, err := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0003-lifecycle")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}

	result, err := domain.ApplySupersession(predecessor, "Use transactional outbox", "bob", []domain.EvidenceReference{evidence}, now)
	if err != nil {
		t.Fatalf("ApplySupersession: %v", err)
	}
	if result.Successor.Statement != "Use transactional outbox" {
		t.Fatalf("successor statement = %q", result.Successor.Statement)
	}
	if result.Successor.Scope != predecessor.Scope {
		t.Fatalf("successor scope = %+v", result.Successor.Scope)
	}
	if result.Successor.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatalf("origin = %q", result.Successor.Origin)
	}
	if result.Successor.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("successor status = %q", result.Successor.VerificationStatus)
	}
	if result.Successor.CreatedBy != "bob" {
		t.Fatalf("created_by = %q", result.Successor.CreatedBy)
	}
	if len(result.Successor.Evidence) != 1 || result.Successor.Evidence[0].Value != "0003-lifecycle" {
		t.Fatalf("evidence = %+v", result.Successor.Evidence)
	}
	if result.Predecessor.SupersededByID == nil || *result.Predecessor.SupersededByID != result.Successor.ID {
		t.Fatal("predecessor must point at successor")
	}
	if result.Predecessor.Statement != predecessor.Statement {
		t.Fatal("predecessor statement must be preserved")
	}
	if result.SupersedeAudit.Action != domain.AuditActionSupersede || result.SupersedeAudit.TargetID != predecessor.ID {
		t.Fatalf("supersede audit = %+v", result.SupersedeAudit)
	}
	if result.CreateAudit.Action != domain.AuditActionCreate || result.CreateAudit.TargetID != result.Successor.ID {
		t.Fatalf("create audit = %+v", result.CreateAudit)
	}
}

func TestApplySupersessionOmitsEvidenceWhenNotProvided(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	predEvidence, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "old-adr")
	predecessor.Evidence = []domain.EvidenceReference{predEvidence}

	result, err := domain.ApplySupersession(predecessor, "Replacement", "alice", nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("ApplySupersession: %v", err)
	}
	if len(result.Successor.Evidence) != 0 {
		t.Fatalf("successor evidence should be empty, got %+v", result.Successor.Evidence)
	}
	if len(result.Predecessor.Evidence) != 1 {
		t.Fatal("predecessor evidence must be preserved")
	}
}

func TestApplySupersessionRejectsAlreadySuperseded(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	id := "already"
	predecessor.SupersededByID = &id
	_, err := domain.ApplySupersession(predecessor, "Replacement", "alice", nil, time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestApplySupersessionRejectsInvalidated(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	predecessor.VerificationStatus = domain.VerificationInvalidated
	_, err := domain.ApplySupersession(predecessor, "Replacement", "alice", nil, time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestApplySupersessionRejectsBlankActorAndStatement(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	now := time.Now().UTC()
	if _, err := domain.ApplySupersession(predecessor, "Replacement", "  ", nil, now); err == nil {
		t.Fatal("expected blank actor error")
	}
	if _, err := domain.ApplySupersession(predecessor, "  ", "alice", nil, now); err == nil {
		t.Fatal("expected empty statement error")
	}
}

func TestApplySupersessionWithSuccessorUsesPrebuiltArchitectureDecision(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	predEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0003-old")
	predecessor, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Old decision.", Scope: scope, CreatedBy: "alice", Evidence: []domain.EvidenceReference{predEv}, ID: "pred", Now: now,
	})
	if err != nil {
		t.Fatalf("pred: %v", err)
	}
	succEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0007-new")
	successor, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "New decision.", Scope: scope, CreatedBy: "alice", Evidence: []domain.EvidenceReference{succEv}, ID: "succ", Now: now,
	})
	if err != nil {
		t.Fatalf("succ: %v", err)
	}
	result, err := domain.ApplySupersessionWithSuccessor(predecessor, successor, "alice", now)
	if err != nil {
		t.Fatalf("ApplySupersessionWithSuccessor: %v", err)
	}
	if result.Successor.Origin != domain.KnowledgeOriginArchitectureDecision {
		t.Fatalf("origin = %q", result.Successor.Origin)
	}
	if result.Successor.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", result.Successor.VerificationStatus)
	}
	if result.Predecessor.SupersededByID == nil || *result.Predecessor.SupersededByID != "succ" {
		t.Fatal("predecessor must point at successor")
	}
	if result.SupersedeAudit.Action != domain.AuditActionSupersede {
		t.Fatalf("audit = %+v", result.SupersedeAudit)
	}
}

func TestApplySupersessionStillCreatesHumanAuthoredSuccessor(t *testing.T) {
	predecessor := testEntry(t, "entry-1")
	result, err := domain.ApplySupersession(predecessor, "Human replacement", "bob", nil, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if result.Successor.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatalf("human supersede origin = %q", result.Successor.Origin)
	}
}
