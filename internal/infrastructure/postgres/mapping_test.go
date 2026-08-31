package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

func TestLoreEntryMappingRoundTripWithEvidence(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/app")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	evidence, err := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-dual-plane")
	if err != nil {
		t.Fatalf("NewEvidenceReference: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Outbox required",
		Scope:     scope,
		CreatedBy: "alice",
		ID:        uuid.NewString(),
		Now:       now,
		Evidence:  []domain.EvidenceReference{evidence},
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}

	insert, err := loreEntryToInsertParams(entry)
	if err != nil {
		t.Fatalf("loreEntryToInsertParams: %v", err)
	}
	row := sqlc.LoreEntry{
		ID:                 insert.ID,
		Statement:          insert.Statement,
		ScopeKind:          insert.ScopeKind,
		ScopeKey:           insert.ScopeKey,
		Origin:             insert.Origin,
		VerificationStatus: insert.VerificationStatus,
		Evidence:           insert.Evidence,
		CreatedBy:          insert.CreatedBy,
		CreatedAt:          insert.CreatedAt,
		VerifiedBy:         insert.VerifiedBy,
		VerifiedAt:         insert.VerifiedAt,
		UpdatedAt:          insert.UpdatedAt,
		SupersededByID:     insert.SupersededByID,
		InvalidatedBy:      insert.InvalidatedBy,
		InvalidatedAt:      insert.InvalidatedAt,
	}

	got, err := loreEntryFromRow(row)
	if err != nil {
		t.Fatalf("loreEntryFromRow: %v", err)
	}
	if got.Statement != entry.Statement {
		t.Fatalf("statement = %q", got.Statement)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Value != "0001-dual-plane" {
		t.Fatalf("evidence = %+v", got.Evidence)
	}
}

func TestAuditRecordMappingRoundTrip(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	record, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID:  "entry-1",
		Action:    domain.AuditActionCreate,
		ActorID:   "alice",
		ID:        uuid.NewString(),
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewAuditRecord: %v", err)
	}
	insert := auditRecordToInsertParams(record)
	row := sqlc.AuditRecord{
		ID:        insert.ID,
		TargetID:  insert.TargetID,
		Action:    insert.Action,
		ActorID:   insert.ActorID,
		CreatedAt: insert.CreatedAt,
	}
	got := auditRecordFromRow(row)
	if got.Action != domain.AuditActionCreate || got.ActorID != "alice" {
		t.Fatalf("audit = %+v", got)
	}
}

func TestReviewDecisionMappingRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	succ := "successor-id"
	decision := domain.ReviewDecision{
		ID:                uuid.NewString(),
		Scope:             scope,
		EvidenceType:      domain.EvidenceTypeCommit,
		EvidenceValue:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		StatementChecksum: domain.StatementChecksum("Use the outbox."),
		LoreEntryID:       uuid.NewString(),
		SuccessorLoreID:   &succ,
		Status:            domain.ReviewStatusAccepted,
		ActorID:           "alice",
		DecidedAt:         now,
	}
	insert := reviewDecisionToInsertParams(decision)
	row := sqlc.LoreReviewDecision{
		ID:                insert.ID,
		ScopeKind:         insert.ScopeKind,
		ScopeKey:          insert.ScopeKey,
		EvidenceType:      insert.EvidenceType,
		EvidenceValue:     insert.EvidenceValue,
		StatementChecksum: insert.StatementChecksum,
		LoreEntryID:       insert.LoreEntryID,
		SuccessorLoreID:   insert.SuccessorLoreID,
		Status:            insert.Status,
		ActorID:           insert.ActorID,
		DecidedAt:         insert.DecidedAt,
	}
	got, err := reviewDecisionFromRow(row)
	if err != nil {
		t.Fatalf("reviewDecisionFromRow: %v", err)
	}
	if got.Status != domain.ReviewStatusAccepted || got.ActorID != "alice" {
		t.Fatalf("decision = %+v", got)
	}
	if got.SuccessorLoreID == nil || *got.SuccessorLoreID != succ {
		t.Fatalf("successor = %v", got.SuccessorLoreID)
	}
	if got.EvidenceType != domain.EvidenceTypeCommit || got.StatementChecksum != decision.StatementChecksum {
		t.Fatalf("identity = %+v", got)
	}
}
