package ports

import (
	"context"
	"time"
)

// GraphScope filters knowledge graph data by governance scope.
type GraphScope struct {
	Kind string
	Key  string
}

// EpisodeIngestRequest ingests lore into the knowledge graph.
type EpisodeIngestRequest struct {
	EpisodeID       string
	Statement       string
	Scope           GraphScope
	Metadata        map[string]any
	ProvenanceRefs  []string
	ReferenceTime   *time.Time
}

// SearchRequest queries the knowledge graph.
type SearchRequest struct {
	Query string
	Scope *GraphScope
	Limit int
}

// GraphFact is a MemLore-shaped knowledge graph result.
type GraphFact struct {
	ID              string
	Statement       string
	Score           float64
	Scope           *GraphScope
	ProvenanceRefs  []string
}

// KnowledgeGraph coordinates knowledge-plane retrieval and ingest.
// Implementations MUST NOT expose Graphiti types.
type KnowledgeGraph interface {
	Health(ctx context.Context) error
	IngestEpisode(ctx context.Context, req EpisodeIngestRequest) (episodeID string, err error)
	Search(ctx context.Context, req SearchRequest) ([]GraphFact, error)
	GetFact(ctx context.Context, id string) (GraphFact, error)
}
