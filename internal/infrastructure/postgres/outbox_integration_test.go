//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
	"github.com/memlore/memlore/internal/infrastructure/postgres"
)

func TestOutboxRoundTripIntegration(t *testing.T) {
	pool := setupPool(t)
	defer pool.Close()

	ctx := context.Background()
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	uow, err := postgres.BeginUnitOfWork(ctx, pool)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer uow.Rollback(ctx)

	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/outbox-integration")
	if err != nil {
		t.Fatalf("scope: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Outbox integration test statement",
		Scope:     scope,
		CreatedBy: "integration",
		Evidence:  nil,
		Now:       now,
	})
	if err != nil {
		t.Fatalf("entry: %v", err)
	}
	event, err := domain.NewEpisodeIngestOutboxEvent(entry, now)
	if err != nil {
		t.Fatalf("outbox event: %v", err)
	}
	if err := uow.LoreEntries().Add(ctx, entry); err != nil {
		t.Fatalf("add lore: %v", err)
	}
	if err := uow.Outbox().Add(ctx, event); err != nil {
		t.Fatalf("add outbox: %v", err)
	}
	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	graph := &memory.KnowledgeGraph{}
	runner := postgres.NewOutboxProcessor(pool)
	worker := commands.NewProcessOutboxHandler(runner, graph, clock.FixedClock{Instant: now}, 10)
	processed, err := worker.ProcessOnce(ctx)
	if err != nil {
		t.Fatalf("process: %v", err)
	}
	if processed != 1 {
		t.Fatalf("processed = %d", processed)
	}
	if len(graph.IngestCalls) != 1 {
		t.Fatalf("ingest calls = %d", len(graph.IngestCalls))
	}
}
