package commands

import (
	"context"
	"fmt"

	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// ProcessOutboxHandler publishes pending outbox events to the knowledge graph.
type ProcessOutboxHandler struct {
	runner ports.OutboxRunner
	graph  ports.KnowledgeGraph
	clock  ports.Clock
	limit  int
}

func NewProcessOutboxHandler(
	runner ports.OutboxRunner,
	graph ports.KnowledgeGraph,
	clock ports.Clock,
	limit int,
) *ProcessOutboxHandler {
	if limit <= 0 {
		limit = 10
	}
	return &ProcessOutboxHandler{
		runner: runner,
		graph:  graph,
		clock:  clock,
		limit:  limit,
	}
}

func (h *ProcessOutboxHandler) ProcessOnce(ctx context.Context) (int, error) {
	now := h.clock.Now()
	return h.runner.ProcessBatch(ctx, h.limit, now, h.handleEvent)
}

func (h *ProcessOutboxHandler) handleEvent(ctx context.Context, event domain.OutboxEvent) error {
	switch event.EventType {
	case domain.OutboxEventTypeEpisodeIngest:
		return h.ingestEpisode(ctx, event)
	default:
		return fmt.Errorf("unsupported outbox event type: %s", event.EventType)
	}
}

func (h *ProcessOutboxHandler) ingestEpisode(ctx context.Context, event domain.OutboxEvent) error {
	payload, err := domain.ParseEpisodeIngestPayload(event.Payload)
	if err != nil {
		return err
	}
	kind, err := domain.ParseScopeKind(payload.ScopeKind)
	if err != nil {
		return err
	}
	_, err = h.graph.IngestEpisode(ctx, ports.EpisodeIngestRequest{
		EpisodeID:      payload.EpisodeID,
		Statement:      payload.Statement,
		Scope:          ports.GraphScope{Kind: string(kind), Key: payload.ScopeKey},
		ProvenanceRefs: payload.ProvenanceRefs,
	})
	if err != nil {
		return err
	}
	return nil
}
