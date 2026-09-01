package commands_test

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

func seedPending(t *testing.T, uow *memory.UnitOfWork, statement string, ev domain.EvidenceReference) domain.LoreEntry {
	t.Helper()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: statement,
		Scope:     scope,
		CreatedBy: "ingest",
		Evidence:  []domain.EvidenceReference{ev},
		Now:       time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	return entry
}

func TestAcceptReviewAsStatedCreatesHumanVerifiedSuccessor(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pred := seedPending(t, uow, "Payment events use transactional outbox.", ev)
	handler := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: now})

	succ, err := handler.Handle(context.Background(), commands.AcceptReviewCommand{
		EntryID: pred.ID,
		ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if succ.Origin != domain.KnowledgeOriginHumanVerified || succ.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("successor = %+v", succ)
	}
	if len(succ.Evidence) != 1 || succ.Evidence[0].Value != "acme/payments#1842" {
		t.Fatalf("evidence = %+v", succ.Evidence)
	}
	gotPred, err := uow.LoreEntries().Get(context.Background(), pred.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotPred.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatal("predecessor origin rewritten")
	}
	if gotPred.SupersededByID == nil || *gotPred.SupersededByID != succ.ID {
		t.Fatal("predecessor not superseded")
	}
	outbox, _ := uow.Outbox().(*memory.OutboxRepository)
	if len(outbox.ListPending()) != 1 {
		t.Fatalf("outbox = %d", len(outbox.ListPending()))
	}

	again, err := handler.Handle(context.Background(), commands.AcceptReviewCommand{
		EntryID: pred.ID,
		ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != succ.ID {
		t.Fatalf("idempotent successor = %s want %s", again.ID, succ.ID)
	}

	listed, err := queries.NewListDecisionsHandler(begin).Handle(context.Background(), queries.ListDecisionsQuery{Scope: pred.Scope})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range listed {
		if d.ID == succ.ID {
			t.Fatal("F035 Accept must not create a Decision")
		}
	}
}

func TestAcceptReviewEditIsHumanAuthored(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	pred := seedPending(t, uow, "Payment events use transactional outbox.", ev)
	edited := "Payment events MUST use the transactional outbox."
	handler := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	succ, err := handler.Handle(context.Background(), commands.AcceptReviewCommand{
		EntryID:   pred.ID,
		ActorID:   "alice",
		Statement: &edited,
	})
	if err != nil {
		t.Fatal(err)
	}
	if succ.Origin != domain.KnowledgeOriginHumanAuthored || succ.Statement != edited {
		t.Fatalf("successor = %+v", succ)
	}
	gotPred, _ := uow.LoreEntries().Get(context.Background(), pred.ID)
	if gotPred.Statement != "Payment events use transactional outbox." {
		t.Fatalf("predecessor rewritten: %q", gotPred.Statement)
	}
}

func TestAcceptReviewSameStatementAfterTrimIsAsStated(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pred := seedPending(t, uow, "Payment events use transactional outbox.", ev)
	same := "  Payment events use transactional outbox.  "
	handler := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	succ, err := handler.Handle(context.Background(), commands.AcceptReviewCommand{
		EntryID:   pred.ID,
		ActorID:   "alice",
		Statement: &same,
	})
	if err != nil {
		t.Fatal(err)
	}
	if succ.Origin != domain.KnowledgeOriginHumanVerified {
		t.Fatalf("origin = %s", succ.Origin)
	}
}

func TestAcceptReviewRejectsADRHumanBlankActorEmptyStatement(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL.",
		Scope:     scope,
		CreatedBy: "alice",
		Evidence:  []domain.EvidenceReference{adrEv},
	})
	_ = uow.LoreEntries().Add(context.Background(), adr)
	handler := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})

	_, err := handler.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: adr.ID, ActorID: "alice"})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("adr: %v", err)
	}

	human, _ := domain.NewLoreEntry(domain.NewLoreEntryInput{Statement: "Human", Scope: scope, CreatedBy: "alice"})
	_ = uow.LoreEntries().Add(context.Background(), human)
	_, err = handler.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: human.ID, ActorID: "alice"})
	if !errors.As(err, &ve) {
		t.Fatalf("human: %v", err)
	}

	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1")
	pred := seedPending(t, uow, "x", ev)
	_, err = handler.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: pred.ID, ActorID: "  "})
	if !errors.As(err, &ve) {
		t.Fatalf("blank actor: %v", err)
	}
	empty := "   "
	_, err = handler.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: pred.ID, ActorID: "alice", Statement: &empty})
	if !errors.As(err, &ve) {
		t.Fatalf("empty statement: %v", err)
	}

	_, err = handler.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: "missing", ActorID: "alice"})
	var nf *domain.NotFoundError
	if !errors.As(err, &nf) {
		t.Fatalf("missing: %v", err)
	}
}

func TestRejectReviewHidesFromPendingAndKeepsObservation(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	pred := seedPending(t, uow, "Noisy extract", ev)
	reject := commands.NewRejectReviewHandler(begin, clock.FixedClock{Instant: now})
	got, err := reject.Handle(context.Background(), commands.RejectReviewCommand{EntryID: pred.ID, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.ReviewStatusRejected || got.ID != pred.ID {
		t.Fatalf("got %+v", got)
	}
	stored, err := uow.LoreEntries().Get(context.Background(), pred.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Origin != domain.KnowledgeOriginRepositoryObservation || domain.IsSuperseded(stored) {
		t.Fatalf("lore mutated: %+v", stored)
	}

	again, err := reject.Handle(context.Background(), commands.RejectReviewCommand{EntryID: pred.ID, ActorID: "bob"})
	if err != nil {
		t.Fatal(err)
	}
	if again.Actor != "alice" {
		t.Fatalf("idempotent actor = %s", again.Actor)
	}

	list := queries.NewListReviewQueueHandler(begin)
	items, err := list.Handle(context.Background(), queries.ListReviewQueueQuery{Scope: pred.Scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("pending after reject = %+v", items)
	}

	accept := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: now})
	_, err = accept.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: pred.ID, ActorID: "alice"})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("accept after reject: %v", err)
	}
}

func TestRejectAfterAcceptFails(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pred := seedPending(t, uow, "Payment events use transactional outbox.", ev)
	accept := commands.NewAcceptReviewHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	if _, err := accept.Handle(context.Background(), commands.AcceptReviewCommand{EntryID: pred.ID, ActorID: "alice"}); err != nil {
		t.Fatal(err)
	}
	reject := commands.NewRejectReviewHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	_, err := reject.Handle(context.Background(), commands.RejectReviewCommand{EntryID: pred.ID, ActorID: "alice"})
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("reject after accept: %v", err)
	}
}

func TestVerifyDoesNotCountAsAccept(t *testing.T) {
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	ev, _ := domain.NewEvidenceReference(domain.EvidenceTypePR, "acme/payments#1842")
	pred := seedPending(t, uow, "Payment events use transactional outbox.", ev)
	verify := commands.NewVerifyLoreHandler(begin, clock.FixedClock{Instant: time.Now().UTC()})
	updated, err := verify.Handle(context.Background(), commands.VerifyLoreCommand{EntryID: pred.ID, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatalf("verify changed origin: %s", updated.Origin)
	}
	items, err := queries.NewListReviewQueueHandler(begin).Handle(context.Background(), queries.ListReviewQueueQuery{Scope: pred.Scope})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("verify should still be pending, got %d", len(items))
	}
}
