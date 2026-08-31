package queries_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func reviewScope(t *testing.T) domain.Scope {
	t.Helper()
	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	if err != nil {
		t.Fatal(err)
	}
	return scope
}

func seedObservation(t *testing.T, uow *memory.UnitOfWork, statement string, ev domain.EvidenceReference, now time.Time) domain.LoreEntry {
	t.Helper()
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: statement,
		Scope:     reviewScope(t),
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestListReviewQueueShowsGitAndPRExcludesADRAndRejected(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	scope := reviewScope(t)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	gitEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	prEv, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")

	git := seedObservation(t, uow, "Use the outbox because dual-writes race.", gitEv, now)
	pr := seedObservation(t, uow, "Payment events use transactional outbox.", prEv, now.Add(time.Second))

	adr, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL as the system of record.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{adrEv},
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), adr); err != nil {
		t.Fatal(err)
	}

	rejectedEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	rejected := seedObservation(t, uow, "Noisy extract", rejectedEv, now)
	decision, err := domain.RejectSuggestedLore(rejected, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.ReviewDecisions().Insert(context.Background(), decision); err != nil {
		t.Fatal(err)
	}

	human, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Human authored rule",
		Scope:     scope,
		CreatedBy: "alice",
		Now:       now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), human); err != nil {
		t.Fatal(err)
	}

	handler := queries.NewListReviewQueueHandler(begin)
	items, err := handler.Handle(context.Background(), queries.ListReviewQueueQuery{Scope: scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d %+v", len(items), items)
	}
	ids := map[string]string{}
	for _, item := range items {
		ids[item.Entry.ID] = item.SourceType
		if item.Status != queries.ReviewStatusPending {
			t.Fatalf("status = %s", item.Status)
		}
		raw, _ := json.Marshal(item)
		if string(raw) != "" && (containsJSONKey(raw, "confidence") || containsJSONKey(raw, "reason")) {
			t.Fatalf("invented metadata: %s", raw)
		}
	}
	if ids[git.ID] != "commit" || ids[pr.ID] != "pr" {
		t.Fatalf("ids = %+v", ids)
	}
}

func containsJSONKey(raw []byte, key string) bool {
	s := string(raw)
	return len(s) > 0 && (stringContains(s, `"`+key+`"`))
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestListReviewQueueRejectsNonRepository(t *testing.T) {
	uow := memory.NewUnitOfWork()
	handler := queries.NewListReviewQueueHandler(memory.BeginFactory(uow))
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "t1")
	_, err := handler.Handle(context.Background(), queries.ListReviewQueueQuery{Scope: scope})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetReviewItemPendingAndRejected(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	prEv, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pending := seedObservation(t, uow, "Payment events use transactional outbox.", prEv, now)

	get := queries.NewGetReviewItemHandler(begin)
	item, err := get.Handle(context.Background(), queries.GetReviewItemQuery{ID: pending.ID})
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != queries.ReviewStatusPending || item.SourceType != "pr" {
		t.Fatalf("item = %+v", item)
	}

	decision, err := domain.RejectSuggestedLore(pending, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.ReviewDecisions().Insert(context.Background(), decision); err != nil {
		t.Fatal(err)
	}
	got, err := get.Handle(context.Background(), queries.GetReviewItemQuery{ID: pending.ID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != string(domain.ReviewStatusRejected) {
		t.Fatalf("status = %s", got.Status)
	}

	_, err = get.Handle(context.Background(), queries.GetReviewItemQuery{ID: "missing"})
	if err == nil {
		t.Fatal("expected not found")
	}
}
