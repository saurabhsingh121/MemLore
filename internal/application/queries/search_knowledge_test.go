package queries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestSearchKnowledgeRequiresQuery(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	graph := &memory.KnowledgeGraph{}
	handler := queries.NewSearchKnowledgeHandler(begin, graph, nil)

	_, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if _, ok := err.(*domain.ValidationError); !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestSearchKnowledgeParallelGovernanceAndGraph(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	_, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Use outbox for payments.",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create lore: %v", err)
	}

	graph := &memory.KnowledgeGraph{
		SearchFacts: []ports.GraphFact{
			{
				ID:        "fact-1",
				Statement: "Payment events use transactional outbox.",
				Score:     0.92,
				Scope:     &ports.GraphScope{Kind: "repository", Key: "github.com/acme/payments"},
			},
			{
				ID:        "fact-2",
				Statement: "Lower score fact.",
				Score:     0.40,
			},
		},
	}
	handler := queries.NewSearchKnowledgeHandler(begin, graph, nil)

	result, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{
		Query: "payment outbox",
		Scope: &scope,
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Governance) != 1 {
		t.Fatalf("governance = %+v", result.Governance)
	}
	if len(result.Graph) != 2 {
		t.Fatalf("graph = %+v", result.Graph)
	}
	if result.Graph[0].Score < result.Graph[1].Score {
		t.Fatalf("graph not sorted by score desc: %+v", result.Graph)
	}
	if len(graph.SearchCalls) != 1 {
		t.Fatalf("SearchCalls = %d", len(graph.SearchCalls))
	}
	call := graph.SearchCalls[0]
	if call.Query != "payment outbox" || call.Limit != 5 {
		t.Fatalf("search call = %+v", call)
	}
	if call.Scope == nil || call.Scope.Key != scope.Key {
		t.Fatalf("search scope = %+v", call.Scope)
	}
}

func TestSearchKnowledgeWithoutScopeSkipsGovernance(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	graph := &memory.KnowledgeGraph{
		SearchFacts: []ports.GraphFact{{ID: "f1", Statement: "fact", Score: 0.5}},
	}
	handler := queries.NewSearchKnowledgeHandler(begin, graph, nil)

	result, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{
		Query: "anything",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Governance) != 0 {
		t.Fatalf("governance = %+v", result.Governance)
	}
	if len(result.Graph) != 1 {
		t.Fatalf("graph = %+v", result.Graph)
	}
	if graph.SearchCalls[0].Scope != nil {
		t.Fatalf("expected nil scope, got %+v", graph.SearchCalls[0].Scope)
	}
}

func TestSearchKnowledgeGraphDownReturnsGovernanceWithWarning(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create lore: %v", err)
	}

	graph := &memory.KnowledgeGraph{SearchErr: errors.New("connection refused")}
	handler := queries.NewSearchKnowledgeHandler(begin, graph, nil)

	result, err := handler.Handle(context.Background(), queries.SearchKnowledgeQuery{
		Query: "rule",
		Scope: &scope,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(result.Governance) != 1 {
		t.Fatalf("governance = %+v", result.Governance)
	}
	if len(result.Graph) != 0 {
		t.Fatalf("graph = %+v", result.Graph)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "graph_service_unavailable" {
		t.Fatalf("warnings = %v", result.Warnings)
	}
}
