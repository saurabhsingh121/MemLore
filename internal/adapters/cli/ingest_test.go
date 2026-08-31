package cli_test

import (
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/adapters/cli"
	"github.com/memlore/memlore/internal/domain"
)

func TestParseIngestGitArgsRequiresRepoAndPath(t *testing.T) {
	_, err := cli.ParseIngestGitArgs([]string{"--repository", "r1"})
	if err == nil {
		t.Fatal("expected path error")
	}
	got, err := cli.ParseIngestGitArgs([]string{"--repository", "github.com/acme/payments", "--path", "/tmp/p", "--actor", "alice", "--max-commits", "5"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != "github.com/acme/payments" || got.Path != "/tmp/p" || got.Actor != "alice" || got.MaxCommits != 5 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIngestStatusArgsRequiresRepo(t *testing.T) {
	_, err := cli.ParseIngestStatusArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
}

func TestFormatIngestStatus(t *testing.T) {
	if !strings.Contains(cli.FormatIngestStatus("r1", nil), "(none)") {
		t.Fatal("expected none")
	}
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	run := &domain.IngestRun{
		ID: "run-1", Status: domain.IngestRunSucceeded,
		CommitsSeen: 12, CommitsSkipped: 10, CandidatesStored: 2,
		CursorSHA: "abc", CursorAt: &at,
	}
	out := cli.FormatIngestStatus("github.com/acme/payments", run)
	for _, want := range []string{"Repository: github.com/acme/payments", "succeeded", "commits seen: 12", "candidates stored: 2", "abc"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}
