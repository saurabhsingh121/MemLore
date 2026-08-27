package memory

import (
	"context"
	"errors"

	"github.com/memlore/memlore/internal/application/ports"
)

// KnowledgeGraph is a test double for ports.KnowledgeGraph.
type KnowledgeGraph struct {
	IngestCalls []ports.EpisodeIngestRequest
	SearchCalls []ports.SearchRequest
	SearchFacts []ports.GraphFact
	HealthErr   error
	IngestErr   error
	SearchErr   error
}

func (g *KnowledgeGraph) Health(context.Context) error {
	return g.HealthErr
}

func (g *KnowledgeGraph) IngestEpisode(_ context.Context, req ports.EpisodeIngestRequest) (string, error) {
	if g.IngestErr != nil {
		return "", g.IngestErr
	}
	g.IngestCalls = append(g.IngestCalls, req)
	if req.EpisodeID != "" {
		return req.EpisodeID, nil
	}
	return "episode-stub", nil
}

func (g *KnowledgeGraph) Search(_ context.Context, req ports.SearchRequest) ([]ports.GraphFact, error) {
	if g.SearchErr != nil {
		return nil, g.SearchErr
	}
	g.SearchCalls = append(g.SearchCalls, req)
	if g.SearchFacts != nil {
		return g.SearchFacts, nil
	}
	return []ports.GraphFact{}, nil
}

func (g *KnowledgeGraph) GetFact(context.Context, string) (ports.GraphFact, error) {
	return ports.GraphFact{}, errors.New("not implemented")
}

var _ ports.KnowledgeGraph = (*KnowledgeGraph)(nil)
