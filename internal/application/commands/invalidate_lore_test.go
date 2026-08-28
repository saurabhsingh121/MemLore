package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func seedLore(t *testing.T, uow *memory.UnitOfWork, statement string) domain.LoreEntry {
	t.Helper()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	handler := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entry, err := handler.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: statement,
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return entry
}

func TestInvalidateLoreSetsStatusAndAudit(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	entry := seedLore(t, uow, "Rule")
	now := time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC)
	handler := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: now})

	updated, err := handler.Handle(context.Background(), commands.InvalidateLoreCommand{
		EntryID: entry.ID,
		ActorID: "bob",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if updated.VerificationStatus != domain.VerificationInvalidated {
		t.Fatalf("status = %q", updated.VerificationStatus)
	}
	audits, _ := uow.Audits().ListByTarget(context.Background(), entry.ID)
	invalidateCount := 0
	for _, a := range audits {
		if a.Action == domain.AuditActionInvalidate {
			invalidateCount++
		}
	}
	if invalidateCount != 1 {
		t.Fatalf("invalidate audits = %d (%+v)", invalidateCount, audits)
	}
}

func TestInvalidateLoreIsIdempotent(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	entry := seedLore(t, uow, "Rule")
	now := time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC)
	handler := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: now})

	_, err := handler.Handle(context.Background(), commands.InvalidateLoreCommand{EntryID: entry.ID, ActorID: "alice"})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err = handler.Handle(context.Background(), commands.InvalidateLoreCommand{EntryID: entry.ID, ActorID: "bob"})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	audits, _ := uow.Audits().ListByTarget(context.Background(), entry.ID)
	invalidateCount := 0
	for _, a := range audits {
		if a.Action == domain.AuditActionInvalidate {
			invalidateCount++
		}
	}
	if invalidateCount != 1 {
		t.Fatalf("invalidate audits = %d", invalidateCount)
	}
}

func TestInvalidateLoreRejectsMissingActorAndUnknownID(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	entry := seedLore(t, uow, "Rule")
	handler := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})

	_, err := handler.Handle(context.Background(), commands.InvalidateLoreCommand{EntryID: entry.ID, ActorID: "  "})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("blank actor: %v", err)
	}

	_, err = handler.Handle(context.Background(), commands.InvalidateLoreCommand{
		EntryID: "00000000-0000-0000-0000-000000000000",
		ActorID: "alice",
	})
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("unknown id: %v", err)
	}
}
