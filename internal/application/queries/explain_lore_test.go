package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestExplainLoreIncludesAuthorityEvaluation(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	adr, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001")
	entry, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Use outbox",
		Scope:     scope,
		ActorID:   "alice",
		Evidence:  []domain.EvidenceReference{adr},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := commands.NewVerifyLoreHandler(begin, clock.FixedClock{Instant: now}).Handle(context.Background(), commands.VerifyLoreCommand{
		EntryID: entry.ID,
		ActorID: "alice",
	}); err != nil {
		t.Fatalf("verify: %v", err)
	}

	result, err := queries.NewExplainLoreHandler(begin).Handle(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	if result.Evaluation.Band != domain.TrustBandCanonical {
		t.Fatalf("band = %s", result.Evaluation.Band)
	}
	if len(result.Evaluation.Breakdown) == 0 {
		t.Fatal("expected factor_breakdown")
	}
	if len(result.Audits) < 2 {
		t.Fatalf("audits = %+v", result.Audits)
	}
}

func TestExplainLoreUnknownID(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	_, err := queries.NewExplainLoreHandler(begin).Handle(context.Background(), "00000000-0000-0000-0000-000000000000")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("err = %v", err)
	}
}

func TestExplainLoreInvalidatedIsUntrusted(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	create := commands.NewCreateLoreHandler(begin, clock.FixedClock{Instant: now})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entry, err := create.Handle(context.Background(), commands.CreateLoreCommand{
		Statement: "Bad rule",
		Scope:     scope,
		ActorID:   "alice",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := commands.NewInvalidateLoreHandler(begin, clock.FixedClock{Instant: now}).Handle(context.Background(), commands.InvalidateLoreCommand{
		EntryID: entry.ID,
		ActorID: "alice",
	}); err != nil {
		t.Fatalf("invalidate: %v", err)
	}
	result, err := queries.NewExplainLoreHandler(begin).Handle(context.Background(), entry.ID)
	if err != nil {
		t.Fatalf("explain: %v", err)
	}
	if result.Evaluation.Band != domain.TrustBandUntrusted {
		t.Fatalf("band = %s", result.Evaluation.Band)
	}
}
