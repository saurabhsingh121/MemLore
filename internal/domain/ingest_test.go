package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func TestNewIngestRunStartsRunning(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	run, err := domain.NewIngestRun(domain.NewIngestRunInput{
		Scope:     scope,
		ActorID:   "alice",
		LocalPath: "/tmp/payments",
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewIngestRun: %v", err)
	}
	if run.Status != domain.IngestRunRunning {
		t.Fatalf("status = %q", run.Status)
	}
	if run.ID == "" || run.StartedAt != now {
		t.Fatalf("run = %+v", run)
	}
}

func TestNewIngestRunRejectsNonRepositoryScope(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "t1")
	_, err := domain.NewIngestRun(domain.NewIngestRunInput{
		Scope:     scope,
		ActorID:   "alice",
		LocalPath: "/tmp/x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestIngestRunMarkSucceededAndFailed(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	run, _ := domain.NewIngestRun(domain.NewIngestRunInput{
		Scope: scope, ActorID: "alice", LocalPath: "/tmp/x", Now: now,
	})
	done := now.Add(time.Minute)
	ok := run.MarkSucceeded(done)
	if ok.Status != domain.IngestRunSucceeded || ok.FinishedAt == nil {
		t.Fatalf("succeeded = %+v", ok)
	}
	fail := run.MarkFailed(done, "not a git directory")
	if fail.Status != domain.IngestRunFailed || fail.ErrorSummary != "not a git directory" {
		t.Fatalf("failed = %+v", fail)
	}
}
