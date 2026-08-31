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

type stubGit struct {
	commits []domain.GitCommitSnapshot
	err     error
	calls   int
}

func (s *stubGit) ListCommits(_ context.Context, _ ports.GitLogQuery) ([]domain.GitCommitSnapshot, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.commits, nil
}

func ingestFixture(t *testing.T) (domain.Scope, time.Time, *memory.UnitOfWork, *commands.IngestGitHandler, *stubGit) {
	t.Helper()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uow := memory.NewUnitOfWork()
	git := &stubGit{commits: []domain.GitCommitSnapshot{
		{
			SHA:         "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Author:      "dev",
			CommittedAt: now.Add(-2 * time.Hour),
			Subject:     "feat: add outbox",
			Body:        "Use the transactional outbox because dual-writes race.",
			Message:     "feat: add outbox\n\nUse the transactional outbox because dual-writes race.",
			Paths:       []string{"internal/payments/outbox.go"},
		},
		{
			SHA:         "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Author:      "dev",
			CommittedAt: now.Add(-time.Hour),
			Subject:     "chore: bump version",
			Message:     "chore: bump version",
		},
		{
			SHA:         "cccccccccccccccccccccccccccccccccccccccc",
			Author:      "dev",
			CommittedAt: now.Add(-30 * time.Minute),
			Subject:     "Merge pull request #9",
			Message:     "Merge pull request #9",
			ParentCount: 2,
		},
	}}
	handler := commands.NewIngestGitHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, git)
	return scope, now, uow, handler, git
}

func TestIngestGitStoresOneObservationalCandidate(t *testing.T) {
	scope, _, uow, handler, _ := ingestFixture(t)
	run, err := handler.Handle(context.Background(), commands.IngestGitCommand{
		Scope: scope, Path: "/tmp/payments", ActorID: "alice",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if run.Status != domain.IngestRunSucceeded {
		t.Fatalf("status = %q err=%s", run.Status, run.ErrorSummary)
	}
	if run.CommitsSeen != 3 || run.CommitsSkipped != 2 || run.CandidatesStored != 1 {
		t.Fatalf("counts seen=%d skipped=%d stored=%d", run.CommitsSeen, run.CommitsSkipped, run.CandidatesStored)
	}
	entries, err := uow.LoreEntries().ListByScope(context.Background(), scope)
	if err != nil || len(entries) != 1 {
		t.Fatalf("lore = %d err=%v", len(entries), err)
	}
	e := entries[0]
	if e.Origin != domain.KnowledgeOriginRepositoryObservation {
		t.Fatalf("origin = %q", e.Origin)
	}
	if e.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("status = %q", e.VerificationStatus)
	}
	if e.Origin == domain.KnowledgeOriginHumanAuthored {
		t.Fatal("must not be human_authored")
	}
	hasCommit := false
	for _, ev := range e.Evidence {
		if ev.Type == domain.EvidenceTypeCommit && ev.Value == "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
			hasCommit = true
		}
	}
	if !hasCommit {
		t.Fatalf("evidence = %+v", e.Evidence)
	}
	if len(uow.Outbox().(*memory.OutboxRepository).ListPending()) != 1 {
		t.Fatalf("expected one outbox event")
	}
}

func TestIngestGitIsIdempotentOnSameSHA(t *testing.T) {
	scope, _, uow, handler, _ := ingestFixture(t)
	cmd := commands.IngestGitCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	run, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if run.CandidatesStored != 0 {
		t.Fatalf("second run stored %d", run.CandidatesStored)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("lore count = %d", len(entries))
	}
}

func TestIngestGitDoesNotInventFromNoisySHAOnRetry(t *testing.T) {
	scope, _, uow, handler, git := ingestFixture(t)
	cmd := commands.IngestGitCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	// Even if extractor rules changed, processed noisy SHA stays skipped.
	git.commits[1].Body = "because we needed a bump"
	git.commits[1].Message = "chore: bump version\n\nbecause we needed a bump"
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("noisy SHA must not produce lore later, got %d", len(entries))
	}
}

func TestIngestGitConflictWhenAlreadyRunning(t *testing.T) {
	scope, now, uow, handler, _ := ingestFixture(t)
	running, err := domain.NewIngestRun(domain.NewIngestRunInput{
		Scope: scope, ActorID: "bob", LocalPath: "/tmp/payments", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.Ingest().InsertRun(context.Background(), running); err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(context.Background(), commands.IngestGitCommand{
		Scope: scope, Path: "/tmp/payments", ActorID: "alice",
	})
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v", err)
	}
}

func TestIngestGitRecordsFailedRunWhenGitErrors(t *testing.T) {
	scope, _, _, handler, git := ingestFixture(t)
	git.err = &ports.GitNotRepositoryError{Path: "/tmp/not-git"}
	run, err := handler.Handle(context.Background(), commands.IngestGitCommand{
		Scope: scope, Path: "/tmp/not-git", ActorID: "alice",
	})
	if err != nil {
		t.Fatalf("failed run should not bubble git error, got %v", err)
	}
	if run.Status != domain.IngestRunFailed || run.ErrorSummary == "" {
		t.Fatalf("run = %+v", run)
	}
}

func TestIngestGitNeverVerifies(t *testing.T) {
	scope, _, uow, handler, _ := ingestFixture(t)
	if _, err := handler.Handle(context.Background(), commands.IngestGitCommand{
		Scope: scope, Path: "/tmp/payments", ActorID: "alice",
	}); err != nil {
		t.Fatal(err)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if entries[0].VerificationStatus != domain.VerificationUnverified || entries[0].VerifiedBy != nil {
		t.Fatalf("must remain unverified: %+v", entries[0])
	}
}

func TestIngestGitRequiresActorAndPath(t *testing.T) {
	scope, _, _, handler, _ := ingestFixture(t)
	_, err := handler.Handle(context.Background(), commands.IngestGitCommand{Scope: scope, Path: "/tmp/x"})
	if err == nil {
		t.Fatal("expected actor error")
	}
	_, err = handler.Handle(context.Background(), commands.IngestGitCommand{Scope: scope, ActorID: "alice"})
	if err == nil {
		t.Fatal("expected path error")
	}
}
