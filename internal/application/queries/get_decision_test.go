package queries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestGetDecisionReturnsHumanRow(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	created, err := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: now}).Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: scope, Question: "How should payment events be published?", Choice: "Transactional outbox",
		Owner: "alice", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: created.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.SourceKind != domain.DecisionSourceHuman || got.Choice != "Transactional outbox" {
		t.Fatalf("got = %+v", got)
	}
}

func TestGetDecisionUnknownIDNotFound(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	_, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: "missing"})
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("want not found, got %v", err)
	}
}

func TestGetDecisionProjectsADRLore(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	entry, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		ID:        "adr-1",
		Now:       time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	got, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: "adr-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceKind != domain.DecisionSourceADR || got.ID != "adr-1" || !got.Current {
		t.Fatalf("got = %+v", got)
	}
}

func TestGetDecisionDoesNotProjectGitObservation(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	entry, _ := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "We chose kafka", Scope: scope, CreatedBy: "ingest", Evidence: []domain.EvidenceReference{ev},
	})
	_ = uow.LoreEntries().Add(context.Background(), entry)
	_, err := queries.NewGetDecisionHandler(begin).Handle(context.Background(), queries.GetDecisionQuery{ID: entry.ID})
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("want not found, got %v", err)
	}
}
