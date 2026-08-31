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

func TestGitHubRepoFromScopeKey(t *testing.T) {
	owner, repo, err := domain.GitHubRepoFromScopeKey("github.com/acme/payments")
	if err != nil || owner != "acme" || repo != "payments" {
		t.Fatalf("got %s/%s err=%v", owner, repo, err)
	}
	_, _, err = domain.GitHubRepoFromScopeKey("gitlab.com/acme/payments")
	if err == nil {
		t.Fatal("expected validation error")
	}
	_, _, err = domain.GitHubRepoFromScopeKey("github.com/acme")
	if err == nil {
		t.Fatal("expected validation error for missing repo")
	}
}

func TestPREvidenceValue(t *testing.T) {
	if got := domain.PREvidenceValue("acme", "payments", 1842); got != "acme/payments#1842" {
		t.Fatalf("got %q", got)
	}
}

func TestNewPRIngestRunStartsRunning(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	run, err := domain.NewPRIngestRun(domain.NewPRIngestRunInput{
		Scope: scope, ActorID: "alice", Now: now,
	})
	if err != nil {
		t.Fatalf("NewPRIngestRun: %v", err)
	}
	if run.Status != domain.IngestRunRunning || run.ActorID != "alice" {
		t.Fatalf("run = %+v", run)
	}
	ok := run.MarkSucceeded(now.Add(time.Minute))
	if ok.Status != domain.IngestRunSucceeded {
		t.Fatalf("succeeded = %+v", ok)
	}
}

func TestNewPRIngestRunRejectsUnmappableKey(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "acme/payments")
	_, err := domain.NewPRIngestRun(domain.NewPRIngestRunInput{Scope: scope, ActorID: "alice"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewADRIngestRunStartsRunning(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	run, err := domain.NewADRIngestRun(domain.NewADRIngestRunInput{
		Scope: scope, ActorID: "alice", LocalPath: "/tmp/payments", ExtraDirs: []string{" architecture/records "}, Now: now,
	})
	if err != nil {
		t.Fatalf("NewADRIngestRun: %v", err)
	}
	if run.Status != domain.IngestRunRunning || run.LocalPath != "/tmp/payments" {
		t.Fatalf("run = %+v", run)
	}
	if len(run.ExtraDirs) != 1 || run.ExtraDirs[0] != "architecture/records" {
		t.Fatalf("extra dirs = %+v", run.ExtraDirs)
	}
	ok := run.MarkSucceeded(now.Add(time.Minute))
	if ok.Status != domain.IngestRunSucceeded {
		t.Fatalf("succeeded = %+v", ok)
	}
	fail := run.MarkFailed(now.Add(time.Minute), "path is not a directory")
	if fail.Status != domain.IngestRunFailed || fail.ErrorSummary != "path is not a directory" {
		t.Fatalf("failed = %+v", fail)
	}
}

func TestNewADRIngestRunRejectsNonRepositoryAndBlankPath(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "t1")
	_, err := domain.NewADRIngestRun(domain.NewADRIngestRunInput{Scope: scope, ActorID: "alice", LocalPath: "/tmp/x"})
	if err == nil {
		t.Fatal("expected non-repository error")
	}
	repo, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	_, err = domain.NewADRIngestRun(domain.NewADRIngestRunInput{Scope: repo, ActorID: "alice"})
	if err == nil {
		t.Fatal("expected blank path error")
	}
}

func TestClassifyADRStatus(t *testing.T) {
	if domain.ClassifyADRStatus("Accepted") != domain.ADRStatusAccepted {
		t.Fatal("accepted")
	}
	if domain.ClassifyADRStatus("adopted") != domain.ADRStatusAccepted {
		t.Fatal("adopted")
	}
	if domain.ClassifyADRStatus("Draft") != domain.ADRStatusSkip {
		t.Fatal("draft")
	}
	if domain.ClassifyADRStatus("proposed") != domain.ADRStatusSkip {
		t.Fatal("proposed")
	}
	if domain.ClassifyADRStatus("Rejected") != domain.ADRStatusSkip {
		t.Fatal("rejected")
	}
	if domain.ClassifyADRStatus("Deprecated") != domain.ADRStatusHistorical {
		t.Fatal("deprecated")
	}
	if domain.ClassifyADRStatus("Superseded by ADR-0007") != domain.ADRStatusHistorical {
		t.Fatal("superseded")
	}
	if domain.ClassifyADRStatus("maybe") != domain.ADRStatusUnknown {
		t.Fatal("unknown")
	}
	if domain.ClassifyADRStatus("") != domain.ADRStatusUnknown {
		t.Fatal("empty")
	}
}

func TestADRIdentityFromPathAndDefaultDirs(t *testing.T) {
	if got := domain.ADRIdentityFromPath("docs/adr/0001-use-postgres.md"); got != "0001-use-postgres" {
		t.Fatalf("id = %q", got)
	}
	if len(domain.DefaultADRDirs) != 3 {
		t.Fatalf("default dirs = %+v", domain.DefaultADRDirs)
	}
}
