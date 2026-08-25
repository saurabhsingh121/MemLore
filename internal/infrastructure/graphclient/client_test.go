package graphclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/graphclient"
)

func TestClientHealthAndIngestAgainstStub(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
				"neo4j":  "ok",
			})
		case "/episodes":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"episode_id": "ep-stub-1"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer stub.Close()

	client := graphclient.NewClient(stub.URL, stub.Client())
	ctx := context.Background()

	if err := client.Health(ctx); err != nil {
		t.Fatalf("Health: %v", err)
	}

	id, err := client.IngestEpisode(ctx, ports.EpisodeIngestRequest{
		Statement: "stub statement",
		Scope:     ports.GraphScope{Kind: "team", Key: "t1"},
	})
	if err != nil {
		t.Fatalf("IngestEpisode: %v", err)
	}
	if id != "ep-stub-1" {
		t.Fatalf("episode_id = %s", id)
	}
}
