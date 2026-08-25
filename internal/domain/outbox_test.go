package domain_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func TestNewEpisodeIngestOutboxEvent(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	if err != nil {
		t.Fatalf("scope: %v", err)
	}
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  nil,
		Now:       time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("entry: %v", err)
	}

	event, err := domain.NewEpisodeIngestOutboxEvent(entry, entry.CreatedAt)
	if err != nil {
		t.Fatalf("outbox: %v", err)
	}
	if event.Status != domain.OutboxStatusPending {
		t.Fatalf("status = %s", event.Status)
	}
	if event.IdempotencyKey != "episode.ingest:"+entry.ID {
		t.Fatalf("idempotency = %s", event.IdempotencyKey)
	}

	payload, err := domain.ParseEpisodeIngestPayload(event.Payload)
	if err != nil {
		t.Fatalf("payload: %v", err)
	}
	if payload.EpisodeID != entry.ID {
		t.Fatalf("episode_id = %s", payload.EpisodeID)
	}
}
