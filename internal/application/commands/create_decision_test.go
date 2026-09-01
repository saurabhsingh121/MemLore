package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestCreateDecisionDualWritesVerifiedLoreAndOutbox(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeURL, "https://wiki.example/outbox")
	handler := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: now})

	got, err := handler.Handle(context.Background(), commands.CreateDecisionCommand{
		Scope:    scope,
		Question: "How should payment events be published?",
		Choice:   "Transactional outbox",
		Owner:    "alice",
		Alternatives: []domain.DecisionAlternative{
			{Label: "Dual-write to the topic", Note: "Lost updates on crash"},
		},
		Evidence: []domain.EvidenceReference{ev},
		ActorID:  "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceKind != domain.DecisionSourceHuman || !got.Current {
		t.Fatalf("decision = %+v", got)
	}
	if got.Question == "" || got.Choice != "Transactional outbox" || got.Owner != "alice" {
		t.Fatalf("fields = %+v", got)
	}
	if len(got.Alternatives) != 1 || got.Alternatives[0].Label != "Dual-write to the topic" {
		t.Fatalf("alternatives = %+v", got.Alternatives)
	}
	lore, err := uow.LoreEntries().Get(context.Background(), got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if lore.Statement != got.Choice {
		t.Fatalf("lore statement = %q", lore.Statement)
	}
	if lore.Origin != domain.KnowledgeOriginHumanAuthored || lore.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("lore = %+v", lore)
	}
	if lore.Origin == domain.KnowledgeOriginRepositoryObservation {
		t.Fatal("must not be observational")
	}
	outbox := uow.Outbox().(*memory.OutboxRepository)
	if len(outbox.ListPending()) != 1 {
		t.Fatalf("outbox = %d", len(outbox.ListPending()))
	}
}

func TestCreateDecisionRejectsBlankActorAndNonRepository(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	handler := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	_, err := handler.Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: scope, Question: "Q", Choice: "C", Owner: "alice",
	})
	if err == nil {
		t.Fatal("expected blank actor error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want validation, got %v", err)
	}
	team, _ := domain.NewScope(domain.ScopeKindTeam, "payments")
	_, err = handler.Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: team, Question: "Q", Choice: "C", Owner: "alice", ActorID: "alice",
	})
	if err == nil {
		t.Fatal("expected non-repository error")
	}
}
