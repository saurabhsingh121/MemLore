package domain_test

import (
	"errors"
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

func TestAuditRecordRejectsBlankActor(t *testing.T) {
	_, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID: "entry-1",
		Action:   domain.AuditActionCreate,
		ActorID:  "  ",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "actor_id must be non-empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditRecordRejectsBlankTarget(t *testing.T) {
	_, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID: " ",
		Action:   domain.AuditActionCreate,
		ActorID:  "alice",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "target_id must be non-empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}
