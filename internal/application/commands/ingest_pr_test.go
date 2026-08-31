package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

type stubPRs struct {
	prs   []domain.PullRequestSnapshot
	err   error
	calls int
	query ports.PullRequestQuery
}

func (s *stubPRs) ListPullRequests(_ context.Context, q ports.PullRequestQuery) ([]domain.PullRequestSnapshot, error) {
	s.calls++
	s.query = q
	if s.err != nil {
		return nil, s.err
	}
	if q.Number > 0 {
		for _, pr := range s.prs {
			if pr.Number == q.Number {
				return []domain.PullRequestSnapshot{pr}, nil
			}
		}
		return nil, nil
	}
	out := make([]domain.PullRequestSnapshot, 0, len(s.prs))
	for _, pr := range s.prs {
		if !pr.Merged {
			continue
		}
		if q.AfterMergedAt != nil && pr.MergedAt != nil && !pr.MergedAt.After(*q.AfterMergedAt) {
			continue
		}
		out = append(out, pr)
		if q.MaxPRs > 0 && len(out) >= q.MaxPRs {
			break
		}
	}
	return out, nil
}

func prMergedAt(delta time.Duration) *time.Time {
	t := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC).Add(delta)
	return &t
}

func prIngestFixture(t *testing.T) (domain.Scope, time.Time, *memory.UnitOfWork, *commands.IngestPullRequestsHandler, *stubPRs) {
	t.Helper()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uow := memory.NewUnitOfWork()
	prs := &stubPRs{prs: []domain.PullRequestSnapshot{
		{
			Number: 1842, Owner: "acme", Repo: "payments",
			Title:  "Use transactional outbox",
			Body:   "because dual-writes race on refunds\n\nFixes #381",
			Merged: true, MergedAt: prMergedAt(0),
			AuthorLogin: "dev",
			HTMLURL:     "https://github.com/acme/payments/pull/1842",
			Files:       []string{"internal/payments/outbox.go"},
			ReviewComments: []domain.PullRequestComment{{
				HTMLURL:     "https://github.com/acme/payments/pull/1842#discussion_r9",
				Body:        "because dual-writes race on refunds",
				AuthorLogin: "reviewer",
			}},
		},
		{
			Number: 99, Owner: "acme", Repo: "payments",
			Title:  "chore: bump lodash",
			Merged: true, MergedAt: prMergedAt(time.Hour),
			AuthorLogin: "dependabot[bot]", AuthorIsBot: true,
		},
		{
			Number: 100, Owner: "acme", Repo: "payments",
			Title:  "Use outbox because dual-writes race",
			Merged: false, AuthorLogin: "dev",
		},
	}}
	handler := commands.NewIngestPullRequestsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, prs)
	return scope, now, uow, handler, prs
}

func TestIngestPRStoresOneObservationalCandidate(t *testing.T) {
	scope, _, uow, handler, _ := prIngestFixture(t)
	run, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{
		Scope: scope, ActorID: "alice",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if run.Status != domain.IngestRunSucceeded {
		t.Fatalf("status = %q err=%s", run.Status, run.ErrorSummary)
	}
	if run.PRsSeen != 2 || run.PRsSkipped != 1 || run.CandidatesStored != 1 {
		t.Fatalf("counts seen=%d skipped=%d stored=%d", run.PRsSeen, run.PRsSkipped, run.CandidatesStored)
	}
	entries, err := uow.LoreEntries().ListByScope(context.Background(), scope)
	if err != nil || len(entries) != 1 {
		t.Fatalf("entries = %+v err=%v", entries, err)
	}
	e := entries[0]
	if e.Origin != domain.KnowledgeOriginRepositoryObservation || e.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("entry = %+v", e)
	}
	if e.VerifiedBy != nil {
		t.Fatal("must not auto-verify")
	}
	hasPR := false
	hasIssue := false
	hasComment := false
	for _, ev := range e.Evidence {
		if ev.Type == domain.EvidenceTypePR && ev.Value == "acme/payments#1842" {
			hasPR = true
		}
		if ev.Type == domain.EvidenceTypeURL && ev.Value == "https://github.com/acme/payments/issues/381" {
			hasIssue = true
		}
		if ev.Type == domain.EvidenceTypeURL && ev.Value == "https://github.com/acme/payments/pull/1842#discussion_r9" {
			hasComment = true
		}
	}
	if !hasPR || !hasIssue || !hasComment {
		t.Fatalf("evidence = %+v", e.Evidence)
	}
	if len(uow.Outbox().(*memory.OutboxRepository).ListPending()) != 1 {
		t.Fatal("expected one outbox event")
	}
}

func TestIngestPRIsIdempotent(t *testing.T) {
	scope, _, uow, handler, _ := prIngestFixture(t)
	cmd := commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"}
	first, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	second, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if first.CandidatesStored != 1 || second.CandidatesStored != 0 {
		t.Fatalf("first=%d second=%d", first.CandidatesStored, second.CandidatesStored)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("len = %d", len(entries))
	}
}

func TestIngestPRRetryDoesNotDuplicateAfterPartialFailure(t *testing.T) {
	scope, now, uow, _, prs := prIngestFixture(t)
	handler := commands.NewIngestPullRequestsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, prs)
	first, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice", MaxPRs: 1})
	if err != nil {
		t.Fatal(err)
	}
	if first.CandidatesStored != 1 {
		t.Fatalf("first stored = %d", first.CandidatesStored)
	}
	retry, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("after retry len=%d retry stored=%d", len(entries), retry.CandidatesStored)
	}
}

func TestIngestPRCursorSkipsOlderMerged(t *testing.T) {
	scope, _, _, handler, prs := prIngestFixture(t)
	_, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	prs.prs = append(prs.prs, domain.PullRequestSnapshot{
		Number: 2000, Owner: "acme", Repo: "payments",
		Title:  "Prefer ports because adapters leak",
		Merged: true, MergedAt: prMergedAt(48 * time.Hour),
		AuthorLogin: "dev",
	})
	run, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if run.PRsSeen != 1 || run.CandidatesStored != 1 {
		t.Fatalf("incremental seen=%d stored=%d query=%+v", run.PRsSeen, run.CandidatesStored, prs.query)
	}
}

func TestIngestPRConcurrentRunningConflicts(t *testing.T) {
	scope, now, uow, _, prs := prIngestFixture(t)
	handler := commands.NewIngestPullRequestsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, prs)
	running, err := domain.NewPRIngestRun(domain.NewPRIngestRunInput{Scope: scope, ActorID: "bob", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.PRIngest().InsertRun(context.Background(), running); err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v", err)
	}
}

func TestIngestPRSingleNumber(t *testing.T) {
	scope, _, uow, handler, prs := prIngestFixture(t)
	run, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{
		Scope: scope, ActorID: "alice", PR: 1842,
	})
	if err != nil {
		t.Fatal(err)
	}
	if prs.query.Number != 1842 {
		t.Fatalf("query = %+v", prs.query)
	}
	if run.CandidatesStored != 1 {
		t.Fatalf("stored = %d", run.CandidatesStored)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("len = %d", len(entries))
	}
}

func TestIngestPRSkipsUnmergedWhenRequested(t *testing.T) {
	scope, _, uow, handler, _ := prIngestFixture(t)
	run, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{
		Scope: scope, ActorID: "alice", PR: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if run.CandidatesStored != 0 || run.PRsSkipped != 1 {
		t.Fatalf("run = %+v", run)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 0 {
		t.Fatalf("len = %d", len(entries))
	}
}

func TestIngestPRRejectsUnmappableScope(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "gitlab.com/acme/payments")
	uow := memory.NewUnitOfWork()
	handler := commands.NewIngestPullRequestsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: time.Now()}, &stubPRs{})
	_, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
