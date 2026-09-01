package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestSupersedeHumanDecisionPreservesPredecessor(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	create := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: now})
	first, err := create.Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: scope, Question: "How should payment events be published?", Choice: "Transactional outbox",
		Owner: "alice", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	later := now.Add(time.Hour)
	succ, err := commands.NewSupersedeDecisionHandler(begin, clock.FixedClock{Instant: later}).Handle(context.Background(), commands.SupersedeDecisionCommand{
		PredecessorID: first.ID,
		Question:      "How should payment events be published?",
		Choice:        "Outbox + idempotent consumers",
		Owner:         "alice",
		ActorID:       "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if succ.ID == first.ID || !succ.Current || succ.Choice != "Outbox + idempotent consumers" {
		t.Fatalf("successor = %+v", succ)
	}
	pred, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: first.ID})
	if err != nil {
		t.Fatal(err)
	}
	if pred.Current || pred.SupersededByID == nil || *pred.SupersededByID != succ.ID {
		t.Fatalf("predecessor = %+v", pred)
	}
	listed, err := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != succ.ID {
		t.Fatalf("list = %+v", listed)
	}
	outbox := uow.Outbox().(*memory.OutboxRepository)
	if len(outbox.ListPending()) != 2 {
		t.Fatalf("outbox = %d", len(outbox.ListPending()))
	}
}

func TestSupersedeADRProjectionCreatesHumanSuccessor(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope, CreatedBy: "ingest", Evidence: []domain.EvidenceReference{ev}, ID: "adr-1", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = uow.LoreEntries().Add(context.Background(), adr)

	succ, err := commands.NewSupersedeDecisionHandler(begin, clock.FixedClock{Instant: now.Add(time.Hour)}).Handle(
		context.Background(),
		commands.SupersedeDecisionCommand{
			PredecessorID: "adr-1",
			Question:      "What is the system of record?",
			Choice:        "PostgreSQL with read replicas",
			Owner:         "alice",
			ActorID:       "alice",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if succ.SourceKind != domain.DecisionSourceHuman {
		t.Fatalf("source = %s", succ.SourceKind)
	}
	pred, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: "adr-1"})
	if err != nil {
		t.Fatal(err)
	}
	if pred.Current || pred.SourceKind != domain.DecisionSourceADR {
		t.Fatalf("adr pred = %+v", pred)
	}
	listed, _ := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: scope})
	if len(listed) != 1 || listed[0].ID != succ.ID {
		t.Fatalf("list = %+v", listed)
	}
	oldLore, _ := uow.LoreEntries().Get(context.Background(), "adr-1")
	if oldLore.SupersededByID == nil {
		t.Fatal("adr lore not superseded")
	}
}

func TestSupersedeRejectsAlreadySupersededAndBlankActor(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	first, err := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: now}).Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: scope, Question: "Q", Choice: "C", Owner: "alice", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	handler := commands.NewSupersedeDecisionHandler(begin, clock.FixedClock{Instant: now.Add(time.Hour)})
	_, err = handler.Handle(context.Background(), commands.SupersedeDecisionCommand{
		PredecessorID: first.ID, Question: "Q2", Choice: "C2", Owner: "alice",
	})
	if err == nil {
		t.Fatal("expected blank actor")
	}
	_, err = handler.Handle(context.Background(), commands.SupersedeDecisionCommand{
		PredecessorID: first.ID, Question: "Q2", Choice: "C2", Owner: "alice", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(context.Background(), commands.SupersedeDecisionCommand{
		PredecessorID: first.ID, Question: "Q3", Choice: "C3", Owner: "alice", ActorID: "alice",
	})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want validation, got %v", err)
	}
}
