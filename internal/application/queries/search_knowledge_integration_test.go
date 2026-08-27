//go:build integration

package queries_test

import (
	"context"
	"os"
	"testing"

	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/graphclient"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func graphServiceURL() string {
	if url := os.Getenv("MEMLORE_GRAPH_SERVICE_URL"); url != "" {
		return url
	}
	return "http://127.0.0.1:8090"
}

func TestSearchKnowledgeIntegrationWithGraphService(t *testing.T) {
	client := graphclient.NewClient(graphServiceURL(), nil)
	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		t.Skipf("graph-service unavailable: %v", err)
	}

	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	handler := queries.NewSearchKnowledgeHandler(begin, client, nil)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments-integration")

	result, err := handler.Handle(ctx, queries.SearchKnowledgeQuery{
		Query: "payment outbox",
		Scope: &scope,
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if result.Graph == nil {
		t.Fatal("graph results should not be nil slice")
	}
}
