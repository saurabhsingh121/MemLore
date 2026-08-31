package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

type stubGit struct {
	commits []domain.GitCommitSnapshot
}

func (s *stubGit) ListCommits(_ context.Context, _ ports.GitLogQuery) ([]domain.GitCommitSnapshot, error) {
	return s.commits, nil
}

func TestListIngestRunsAndCandidates(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	git := &stubGit{commits: []domain.GitCommitSnapshot{{
		SHA:         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CommittedAt: now,
		Subject:     "feat: outbox",
		Body:        "because dual-writes race",
		Message:     "feat: outbox\n\nbecause dual-writes race",
	}}}
	ingester := commands.NewIngestGitHandler(begin, clock.FixedClock{Instant: now}, git)
	run, err := ingester.Handle(context.Background(), commands.IngestGitCommand{
		Scope: scope, Path: "/tmp/p", ActorID: "alice",
	})
	if err != nil {
		t.Fatal(err)
	}

	runs, err := queries.NewListIngestRunsHandler(begin).Handle(context.Background(), queries.ListIngestRunsQuery{Scope: scope})
	if err != nil || len(runs) != 1 || runs[0].ID != run.ID {
		t.Fatalf("runs = %+v err=%v", runs, err)
	}
	got, err := queries.NewGetIngestRunHandler(begin).Handle(context.Background(), queries.GetIngestRunQuery{ID: run.ID})
	if err != nil || got.Status != domain.IngestRunSucceeded {
		t.Fatalf("get = %+v err=%v", got, err)
	}
	cands, err := queries.NewListIngestCandidatesHandler(begin).Handle(context.Background(), queries.ListIngestCandidatesQuery{Scope: scope})
	if err != nil || len(cands) != 1 {
		t.Fatalf("candidates = %+v err=%v", cands, err)
	}
	if cands[0].Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatalf("origin = %q", cands[0].Origin)
	}
}

func TestListIngestCandidatesFiltersEvidenceType(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	handler := commands.NewIngestPullRequestsHandler(begin, clock.FixedClock{Instant: now}, &stubPRReader{prs: []domain.PullRequestSnapshot{{
		Number: 1842, Owner: "acme", Repo: "payments",
		Title: "Use outbox", Body: "because dual-writes race",
		Merged: true, MergedAt: &now, AuthorLogin: "dev",
	}}})
	_, err := handler.Handle(context.Background(), commands.IngestPullRequestsCommand{Scope: scope, ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	prCands, err := queries.NewListIngestCandidatesHandler(begin).Handle(context.Background(), queries.ListIngestCandidatesQuery{
		Scope: scope, EvidenceType: domain.EvidenceTypePR,
	})
	if err != nil || len(prCands) != 1 {
		t.Fatalf("pr candidates = %+v err=%v", prCands, err)
	}
	commitCands, err := queries.NewListIngestCandidatesHandler(begin).Handle(context.Background(), queries.ListIngestCandidatesQuery{
		Scope: scope, EvidenceType: domain.EvidenceTypeCommit,
	})
	if err != nil || len(commitCands) != 0 {
		t.Fatalf("commit filter = %+v err=%v", commitCands, err)
	}
}

type stubPRReader struct {
	prs []domain.PullRequestSnapshot
}

func (s *stubPRReader) ListPullRequests(_ context.Context, _ ports.PullRequestQuery) ([]domain.PullRequestSnapshot, error) {
	return s.prs, nil
}

func TestListIngestRunsRejectsNonRepository(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "t1")
	uow := memory.NewUnitOfWork()
	_, err := queries.NewListIngestRunsHandler(memory.BeginFactory(uow)).Handle(context.Background(), queries.ListIngestRunsQuery{Scope: scope})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
