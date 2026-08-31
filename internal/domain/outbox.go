package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// OutboxEventType identifies the async action to perform.
type OutboxEventType string

const OutboxEventTypeEpisodeIngest OutboxEventType = "episode.ingest"

// OutboxStatus is the processing state of an outbox row.
type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusCompleted OutboxStatus = "completed"
)

// OutboxEvent is a governance-plane async work item.
type OutboxEvent struct {
	ID             string
	AggregateType  string
	AggregateID    string
	EventType      OutboxEventType
	Payload        []byte
	Status         OutboxStatus
	Attempts       int
	IdempotencyKey string
	CreatedAt      time.Time
	ProcessedAt    *time.Time
	LastError      string
}

// EpisodeIngestPayload is the JSON payload for episode.ingest events.
type EpisodeIngestPayload struct {
	Statement      string   `json:"statement"`
	ScopeKind      string   `json:"scope_kind"`
	ScopeKey       string   `json:"scope_key"`
	ProvenanceRefs []string `json:"provenance_refs"`
	EpisodeID      string   `json:"episode_id"`
}

// NewEpisodeIngestOutboxEvent builds a pending outbox row for lore graph sync.
func NewEpisodeIngestOutboxEvent(entry LoreEntry, now time.Time) (OutboxEvent, error) {
	if strings.TrimSpace(entry.Statement) == "" {
		return OutboxEvent{}, &ValidationError{Message: "statement must be non-empty"}
	}
	payload, err := json.Marshal(EpisodeIngestPayload{
		Statement:      entry.Statement,
		ScopeKind:      string(entry.Scope.Kind),
		ScopeKey:       entry.Scope.Key,
		ProvenanceRefs: []string{entry.ID},
		EpisodeID:      entry.ID,
	})
	if err != nil {
		return OutboxEvent{}, fmt.Errorf("marshal episode ingest payload: %w", err)
	}
	return OutboxEvent{
		ID:             uuid.NewString(),
		AggregateType:  "lore_entry",
		AggregateID:    entry.ID,
		EventType:      OutboxEventTypeEpisodeIngest,
		Payload:        payload,
		Status:         OutboxStatusPending,
		Attempts:       0,
		IdempotencyKey: fmt.Sprintf("episode.ingest:%s", entry.ID),
		CreatedAt:      now,
	}, nil
}

// ParseEpisodeIngestPayload decodes an episode.ingest payload.
func ParseEpisodeIngestPayload(payload []byte) (EpisodeIngestPayload, error) {
	var out EpisodeIngestPayload
	if err := json.Unmarshal(payload, &out); err != nil {
		return EpisodeIngestPayload{}, fmt.Errorf("decode episode ingest payload: %w", err)
	}
	return out, nil
}
