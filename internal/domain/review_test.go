package domain_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func reviewScope(t *testing.T) domain.Scope {
	t.Helper()
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	if err != nil {
		t.Fatalf("scope: %v", err)
	}
	return scope
}

func observationalPR(t *testing.T, statement string) domain.LoreEntry {
	t.Helper()
	ev, err := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: statement,
		Scope:     reviewScope(t),
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("observational: %v", err)
	}
	return entry
}

func TestExtractIdentityPrefersCommitOverPR(t *testing.T) {
	scope := reviewScope(t)
	commit, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	pr, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "  Use the outbox.  ",
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{pr, commit},
	})
	if err != nil {
		t.Fatal(err)
	}
	id, err := domain.ExtractIdentityOf(entry)
	if err != nil {
		t.Fatalf("ExtractIdentityOf: %v", err)
	}
	if id.EvidenceType != domain.EvidenceTypeCommit || id.EvidenceValue != commit.Value {
		t.Fatalf("identity evidence = %s %s", id.EvidenceType, id.EvidenceValue)
	}
	sum := sha256.Sum256([]byte("Use the outbox."))
	if id.StatementChecksum != hex.EncodeToString(sum[:]) {
		t.Fatalf("checksum = %s", id.StatementChecksum)
	}
	if id.Scope.Key != scope.Key {
		t.Fatalf("scope = %+v", id.Scope)
	}
}

func TestExtractIdentityUsesPRWhenNoCommit(t *testing.T) {
	entry := observationalPR(t, "Payment events use transactional outbox.")
	id, err := domain.ExtractIdentityOf(entry)
	if err != nil {
		t.Fatal(err)
	}
	if id.EvidenceType != domain.EvidenceTypePR || id.EvidenceValue != "acme/payments#1842" {
		t.Fatalf("identity = %+v", id)
	}
}

func TestIsReviewEligibleObservationalCurrent(t *testing.T) {
	entry := observationalPR(t, "Payment events use transactional outbox.")
	if !domain.IsReviewEligible(entry) {
		t.Fatal("expected eligible")
	}
}

func TestIsReviewEligibleRejectsADRAndHumanAndSuperseded(t *testing.T) {
	scope := reviewScope(t)
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{adrEv},
	})
	if err != nil {
		t.Fatal(err)
	}
	if domain.IsReviewEligible(adr) {
		t.Fatal("ADR must not be eligible")
	}

	human, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Human rule",
		Scope:     scope,
		CreatedBy: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if domain.IsReviewEligible(human) {
		t.Fatal("human remember must not be eligible")
	}

	obs := observationalPR(t, "obs")
	succ := "successor"
	obs.SupersededByID = &succ
	if domain.IsReviewEligible(obs) {
		t.Fatal("superseded must not be eligible")
	}
}

func TestSourceTypeFromPrimaryEvidence(t *testing.T) {
	pr := observationalPR(t, "x")
	if got := domain.ReviewSourceType(pr); got != "pr" {
		t.Fatalf("source = %s", got)
	}
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	git, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "x",
		Scope:     reviewScope(t),
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := domain.ReviewSourceType(git); got != "commit" {
		t.Fatalf("source = %s", got)
	}
}

func TestAcceptSuggestedLoreAsStated(t *testing.T) {
	pred := observationalPR(t, "Payment events use transactional outbox.")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	result, err := domain.AcceptSuggestedLore(pred, "", "alice", now)
	if err != nil {
		t.Fatalf("AcceptSuggestedLore: %v", err)
	}
	if result.Successor.Origin != domain.KnowledgeOriginHumanVerified {
		t.Fatalf("origin = %q", result.Successor.Origin)
	}
	if result.Successor.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", result.Successor.VerificationStatus)
	}
	if result.Successor.Statement != pred.Statement {
		t.Fatalf("statement = %q", result.Successor.Statement)
	}
	if len(result.Successor.Evidence) != 1 || result.Successor.Evidence[0].Value != "acme/payments#1842" {
		t.Fatalf("evidence dropped: %+v", result.Successor.Evidence)
	}
	if result.Predecessor.SupersededByID == nil || *result.Predecessor.SupersededByID != result.Successor.ID {
		t.Fatalf("predecessor not superseded: %+v", result.Predecessor)
	}
	if result.Predecessor.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatal("predecessor origin rewritten")
	}
	if result.Decision.Status != domain.ReviewStatusAccepted {
		t.Fatalf("decision = %s", result.Decision.Status)
	}
	if result.Decision.SuccessorLoreID == nil || *result.Decision.SuccessorLoreID != result.Successor.ID {
		t.Fatalf("decision successor = %v", result.Decision.SuccessorLoreID)
	}
}

func TestAcceptSuggestedLoreSameStatementAfterTrimIsAsStated(t *testing.T) {
	pred := observationalPR(t, "Payment events use transactional outbox.")
	result, err := domain.AcceptSuggestedLore(pred, "  Payment events use transactional outbox.  ", "alice", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if result.Successor.Origin != domain.KnowledgeOriginHumanVerified {
		t.Fatalf("origin = %q", result.Successor.Origin)
	}
}

func TestAcceptSuggestedLoreEditIsHumanAuthored(t *testing.T) {
	pred := observationalPR(t, "Payment events use transactional outbox.")
	result, err := domain.AcceptSuggestedLore(pred, "Payment events MUST use the transactional outbox.", "alice", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if result.Successor.Origin != domain.KnowledgeOriginHumanAuthored {
		t.Fatalf("origin = %q", result.Successor.Origin)
	}
	if result.Successor.Statement != "Payment events MUST use the transactional outbox." {
		t.Fatalf("statement = %q", result.Successor.Statement)
	}
	if pred.Statement == result.Successor.Statement {
		t.Fatal("predecessor statement should remain distinct")
	}
	if result.Predecessor.Statement != "Payment events use transactional outbox." {
		t.Fatalf("predecessor rewritten: %q", result.Predecessor.Statement)
	}
	if len(result.Successor.Evidence) != 1 {
		t.Fatal("edit dropped evidence")
	}
}

func TestAcceptSuggestedLoreRejectsIneligible(t *testing.T) {
	scope := reviewScope(t)
	human, _ := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Human",
		Scope:     scope,
		CreatedBy: "alice",
	})
	_, err := domain.AcceptSuggestedLore(human, "", "alice", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAcceptSuggestedLoreRejectsBlankActor(t *testing.T) {
	pred := observationalPR(t, "x")
	_, err := domain.AcceptSuggestedLore(pred, "", "  ", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRejectSuggestedLoreDoesNotMutateEntry(t *testing.T) {
	pred := observationalPR(t, "Payment events use transactional outbox.")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	decision, err := domain.RejectSuggestedLore(pred, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Status != domain.ReviewStatusRejected {
		t.Fatalf("status = %s", decision.Status)
	}
	if decision.SuccessorLoreID != nil {
		t.Fatal("reject must not set successor")
	}
	if decision.LoreEntryID != pred.ID {
		t.Fatalf("lore id = %s", decision.LoreEntryID)
	}
	if domain.IsSuperseded(pred) {
		t.Fatal("reject must not supersede")
	}
	if pred.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatal("origin rewritten")
	}
}

func TestRejectSuggestedLoreIdentityMatchesExtract(t *testing.T) {
	pred := observationalPR(t, "Payment events use transactional outbox.")
	decision, err := domain.RejectSuggestedLore(pred, "alice", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	id, err := domain.ExtractIdentityOf(pred)
	if err != nil {
		t.Fatal(err)
	}
	if decision.EvidenceType != id.EvidenceType || decision.EvidenceValue != id.EvidenceValue || decision.StatementChecksum != id.StatementChecksum {
		t.Fatalf("decision identity %+v vs %+v", decision, id)
	}
}

func TestAcceptSuggestedLoreRejectsOversizedEdit(t *testing.T) {
	pred := observationalPR(t, "short")
	_, err := domain.AcceptSuggestedLore(pred, strings.Repeat("x", domain.MaxStatementLength+1), "alice", time.Now().UTC())
	if err == nil {
		t.Fatal("expected validation error")
	}
}
