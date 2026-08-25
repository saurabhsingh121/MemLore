package memory

import (
	"context"
	"errors"

	"github.com/memlore/memlore/internal/application/ports"
)

// KnowledgeGraph is a test double for ports.KnowledgeGraph.
type KnowledgeGraph struct {
	IngestCalls []ports.EpisodeIngestRequest
	HealthErr   error
	IngestErr   error
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

func (g *KnowledgeGraph) Search(context.Context, ports.SearchRequest) ([]ports.GraphFact, error) {
	return nil, nil
}

func (g *KnowledgeGraph) GetFact(context.Context, string) (ports.GraphFact, error) {
	return ports.GraphFact{}, errors.New("not implemented")
}

var _ ports.KnowledgeGraph = (*KnowledgeGraph)(nil)
