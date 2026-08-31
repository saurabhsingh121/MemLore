package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestListLoreByScopeOmitsStaleByDefault(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	invalidate := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: now})
	supersede := commands.NewSupersedeLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")

	current, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Current rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create current: %v", err)
	}
	toInvalidate, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Bad rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create invalidate target: %v", err)
	}
	if _, err := invalidate.Handle(context.Background(), commands.InvalidateLoreCommand{
		EntryID: toInvalidate.ID,
		ActorID: "alice",
	}); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	toSupersede, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Old rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create supersede target: %v", err)
	}
	successor, err := supersede.Handle(context.Background(), commands.SupersedeLoreCommand{
		EntryID:   toSupersede.ID,
		Statement: "Replacement rule",
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}

	list := queries.NewListLoreByScopeHandler(begin)
	got, err := list.Handle(context.Background(), queries.ListLoreByScopeQuery{Scope: scope})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	ids := map[string]bool{}
	for _, e := range got {
		ids[e.ID] = true
	}
	if !ids[current.ID] || !ids[successor.ID] {
		t.Fatalf("expected current entries, got %+v", got)
	}
	if ids[toInvalidate.ID] || ids[toSupersede.ID] {
		t.Fatalf("stale should be omitted, got %+v", got)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d want 2", len(got))
	}

	all, err := list.Handle(context.Background(), queries.ListLoreByScopeQuery{
		Scope:        scope,
		IncludeStale: true,
	})
	if err != nil {
		t.Fatalf("list include_stale: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("include_stale len = %d want 4", len(all))
	}
}

func TestSearchKnowledgeOmitsStaleByDefault(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	invalidate := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")

	_, _ = create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Keep me",
		Scope:     scope,
		ActorID:   "alice",
	})
	bad, _ := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Drop me",
		Scope:     scope,
		ActorID:   "alice",
	})
	_, _ = invalidate.Handle(context.Background(), commands.InvalidateLoreCommand{
		EntryID: bad.ID,
		ActorID: "alice",
	})

	handler := queries.NewSearchKnowledgeHandler(begin, &memory.KnowledgeGraph{}, nil)
	result, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{
		Query: "rules",
		Scope: &scope,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(result.Governance) != 1 {
		t.Fatalf("governance = %+v", result.Governance)
	}
	if result.Governance[0].ID == bad.ID {
		t.Fatal("invalidated should be omitted")
	}

	withStale, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{
		Query:        "rules",
		Scope:        &scope,
		IncludeStale: true,
	})
	if err != nil {
		t.Fatalf("search include_stale: %v", err)
	}
	if len(withStale.Governance) != 2 {
		t.Fatalf("include_stale governance = %+v", withStale.Governance)
	}
}
