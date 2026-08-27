package commands_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestCreateLorePersistsOutboxEvent(t *testing.T) {
	uow := memory.NewUnitOfWork()
	outboxRepo := uow.Outbox().(*memory.OutboxRepository)
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	handler := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")

	entry, err := handler.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	pending := outboxRepo.ListPending()
	if len(pending) != 1 {
		t.Fatalf("pending outbox events = %d", len(pending))
	}
	event := pending[0]
	if event.AggregateID != entry.ID {
		t.Fatalf("aggregate_id = %s", event.AggregateID)
	}
	if event.EventType != domain.OutboxEventTypeEpisodeIngest {
		t.Fatalf("event_type = %s", event.EventType)
	}
}

func TestProcessOutboxPublishesEpisode(t *testing.T) {
	uow := memory.NewUnitOfWork()
	outboxRepo := uow.Outbox().(*memory.OutboxRepository)
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")

	_, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Use the outbox",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	graph := &memory.KnowledgeGraph{}
	runner := memory.NewOutboxRunner(uow.Outbox().(*memory.OutboxRepository))
	worker := commands.NewProcessOutboxHandler(runner, graph, clock.FixedClock{Instant: now}, 10)

	processed, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("ProcessOnce: %v", err)
	}
	if processed != 1 {
		t.Fatalf("processed = %d", processed)
	}
	if len(graph.IngestCalls) != 1 {
		t.Fatalf("ingest calls = %d", len(graph.IngestCalls))
	}
	call := graph.IngestCalls[0]
	if call.Statement != "Use the outbox" {
		t.Fatalf("statement = %q", call.Statement)
	}
	if call.EpisodeID == "" {
		t.Fatal("expected episode id for idempotent ingest")
	}

	pending := outboxRepo.ListPending()
	if len(pending) != 0 {
		t.Fatalf("still pending = %d", len(pending))
	}
}
