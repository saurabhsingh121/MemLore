//go:build integration

package graphclient_test

import (
	"context"
	"os"
	"testing"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/graphclient"
)

func graphServiceURL() string {
	if url := os.Getenv("MEMLORE_GRAPH_SERVICE_URL"); url != "" {
		return url
	}
	return "http://127.0.0.1:8090"
}

func TestClientHealthAndIngestEpisodeIntegration(t *testing.T) {
	client := graphclient.NewClient(graphServiceURL(), nil)
	ctx := context.Background()

	if err := client.Health(ctx); err != nil {
		t.Skipf("graph-service unavailable: %v", err)
	}

	episodeID, err := client.IngestEpisode(ctx, ports.EpisodeIngestRequest{
		Statement: "Go contract test: outbox for payment events.",
		Scope: ports.GraphScope{
			Kind: "repository",
			Key:  "github.com/acme/payments-go-contract",
		},
		ProvenanceRefs: []string{"go-contract-test"},
	})
	if err != nil {
		t.Fatalf("IngestEpisode: %v", err)
	}
	if episodeID == "" {
		t.Fatal("expected non-empty episode_id")
	}
}
