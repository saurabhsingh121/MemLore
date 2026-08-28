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

func TestSupersedeLoreCreatesSuccessorAndAudits(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	predecessor := seedLore(t, uow, "Old rule")
	now := time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC)
	handler := commands.NewSupersedeLoreHandler(begin, clock.FixedClock{Instant: now})

	successor, err := handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   predecessor.ID,
		Statement: "New rule",
		ActorID:   "bob",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if successor.Statement != "New rule" {
		t.Fatalf("statement = %q", successor.Statement)
	}
	if successor.ID == predecessor.ID {
		t.Fatal("successor must have a new id")
	}

	gotPred, err := uow.LoreEntries().Get(context.Background(), predecessor.ID)
	if err != nil {
		t.Fatalf("get predecessor: %v", err)
	}
	if gotPred.SupersededByID == nil || *gotPred.SupersededByID != successor.ID {
		t.Fatal("predecessor must point at successor")
	}

	predAudits, _ := uow.Audits().ListByTarget(context.Background(), predecessor.ID)
	succAudits, _ := uow.Audits().ListByTarget(context.Background(), successor.ID)
	if len(predAudits) != 2 || predAudits[1].Action != domain.AuditActionSupersede {
		t.Fatalf("predecessor audits = %+v", predAudits)
	}
	if len(succAudits) != 1 || succAudits[0].Action != domain.AuditActionCreate {
		t.Fatalf("successor audits = %+v", succAudits)
	}

	outbox, ok := uow.Outbox().(*memory.OutboxRepository)
	if !ok {
		t.Fatal("expected memory outbox")
	}
	pending := outbox.ListPending()
	if len(pending) != 2 {
		t.Fatalf("outbox events = %d (create + successor ingest)", len(pending))
	}
}

func TestSupersedeLoreRejectsAlreadySuperseded(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	predecessor := seedLore(t, uow, "Old rule")
	now := time.Date(2026, 8, 28, 14, 0, 0, 0, time.UTC)
	handler := commands.NewSupersedeLoreHandler(begin, clock.FixedClock{Instant: now})

	_, err := handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   predecessor.ID,
		Statement: "New rule",
		ActorID:   "bob",
	})
	if err != nil {
		t.Fatalf("first supersede: %v", err)
	}
	_, err = handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   predecessor.ID,
		Statement: "Another rule",
		ActorID:   "bob",
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("second supersede: %v", err)
	}
	listed, _ := uow.LoreEntries().ListByScope(context.Background(), predecessor.Scope)
	if len(listed) != 2 {
		t.Fatalf("expected predecessor + one successor, got %d", len(listed))
	}
}

func TestSupersedeLoreRejectsInvalidatedMissingActorAndUnknownID(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	entry := seedLore(t, uow, "Rule")
	inv := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	if _, err := inv.Handle(context.Background(), commands.InvalidateLoreCommand{EntryID: entry.ID, ActorID: "alice"}); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	handler := commands.NewSupersedeLoreHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})

	_, err := handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   entry.ID,
		Statement: "Replacement",
		ActorID:   "alice",
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("invalidated supersede: %v", err)
	}

	fresh := seedLore(t, uow, "Other")
	_, err = handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   fresh.ID,
		Statement: "Replacement",
		ActorID:   "  ",
	})
	if !errors.As(err, &ve) {
		t.Fatalf("blank actor: %v", err)
	}

	_, err = handler.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   "00000000-0000-0000-0000-000000000000",
		Statement: "Replacement",
		ActorID:   "alice",
	})
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("unknown id: %v", err)
	}
}
