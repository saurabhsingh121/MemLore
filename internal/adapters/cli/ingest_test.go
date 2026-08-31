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
	got, err := cli.ParseIngestStatusArgs([]string{"--repository", "github.com/acme/payments"})
	if err != nil || got.Kind != "git" {
		t.Fatalf("default kind: %+v err=%v", got, err)
	}
	pr, err := cli.ParseIngestStatusArgs([]string{"--repository", "r1", "--kind", "pr"})
	if err != nil || pr.Kind != "pr" {
		t.Fatalf("kind pr: %+v err=%v", pr, err)
	}
	adr, err := cli.ParseIngestStatusArgs([]string{"--repository", "r1", "--kind", "adr"})
	if err != nil || adr.Kind != "adr" {
		t.Fatalf("kind adr: %+v err=%v", adr, err)
	}
}

func TestParseIngestPRArgs(t *testing.T) {
	_, err := cli.ParseIngestPRArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	got, err := cli.ParseIngestPRArgs([]string{"--repository", "github.com/acme/payments", "--pr", "1842", "--max-prs", "5", "--actor", "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != "github.com/acme/payments" || got.PR != 1842 || got.MaxPRs != 5 || got.Actor != "alice" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseIngestADRArgs(t *testing.T) {
	_, err := cli.ParseIngestADRArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	got, err := cli.ParseIngestADRArgs([]string{
		"--repository", "github.com/acme/payments", "--path", "/tmp/p",
		"--adr-dir", "architecture/records", "--adr-dir", "decisions", "--actor", "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != "github.com/acme/payments" || got.Path != "/tmp/p" || got.Actor != "alice" {
		t.Fatalf("got %+v", got)
	}
	if len(got.ADRDirs) != 2 || got.ADRDirs[0] != "architecture/records" {
		t.Fatalf("dirs = %+v", got.ADRDirs)
	}
}

func TestFormatADRIngestStatus(t *testing.T) {
	if !strings.Contains(cli.FormatADRIngestStatus("r1", nil), "(none)") {
		t.Fatal("expected none")
	}
	run := &domain.ADRIngestRun{
		ID: "run-1", Status: domain.IngestRunSucceeded,
		FilesSeen: 4, FilesSkipped: 3, LoreStored: 1, LoreSuperseded: 0,
	}
	out := cli.FormatADRIngestStatus("github.com/acme/payments", run)
	for _, want := range []string{"Repository: github.com/acme/payments", "Latest ADR run: succeeded", "files seen: 4", "lore stored: 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestFormatPRIngestStatus(t *testing.T) {
	if !strings.Contains(cli.FormatPRIngestStatus("r1", nil), "(none)") {
		t.Fatal("expected none")
	}
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	run := &domain.PRIngestRun{
		ID: "run-1", Status: domain.IngestRunSucceeded,
		PRsSeen: 12, PRsSkipped: 10, CandidatesStored: 2,
		CursorPR: 1842, CursorAt: &at,
	}
	out := cli.FormatPRIngestStatus("github.com/acme/payments", run)
	for _, want := range []string{"Repository: github.com/acme/payments", "Latest PR run: succeeded", "prs seen: 12", "candidates stored: 2", "#1842"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
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
