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

func TestListDecisionsUnionsHumanAndADRExcludesObservations(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")

	_, err := commands.NewCreateDecisionHandler(begin, clock.FixedClock{Instant: now}).Handle(context.Background(), commands.CreateDecisionCommand{
		Scope: scope, Question: "How should payment events be published?", Choice: "Transactional outbox",
		Owner: "alice", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{adrEv},
		ID:        "adr-1",
		Now:       now.Add(-time.Hour),
	})
	_ = uow.LoreEntries().Add(context.Background(), adr)

	gitEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	git, _ := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox because dual-writes race.",
		Scope:     scope, CreatedBy: "ingest", Evidence: []domain.EvidenceReference{gitEv},
	})
	_ = uow.LoreEntries().Add(context.Background(), git)

	prEv, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pr, _ := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Payment events use transactional outbox.",
		Scope:     scope, CreatedBy: "ingest", Evidence: []domain.EvidenceReference{prEv},
	})
	_ = uow.LoreEntries().Add(context.Background(), pr)

	items, err := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %+v", items)
	}
	kinds := map[domain.DecisionSourceKind]int{}
	for _, item := range items {
		kinds[item.SourceKind]++
		if !item.Current {
			t.Fatalf("non-current in list: %+v", item)
		}
	}
	if kinds[domain.DecisionSourceHuman] != 1 || kinds[domain.DecisionSourceADR] != 1 {
		t.Fatalf("kinds = %+v", kinds)
	}
}

func TestListDecisionsExcludesSupersededADR(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL.", Scope: scope, CreatedBy: "ingest",
		Evidence: []domain.EvidenceReference{ev}, ID: "adr-old",
	})
	succ := "newer"
	adr.SupersededByID = &succ
	_ = uow.LoreEntries().Add(context.Background(), adr)

	items, err := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %+v", items)
	}
}
