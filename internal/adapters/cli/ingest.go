package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// IngestGitArgs is parsed input for `memlore ingest git`.
type IngestGitArgs struct {
	Repository string
	Path       string
	MaxCommits int
	Actor      string
}

// IngestStatusArgs is parsed input for `memlore ingest status`.
type IngestStatusArgs struct {
	Repository string
	Kind       string
}

// IngestPRArgs is parsed input for `memlore ingest pr`.
type IngestPRArgs struct {
	Repository string
	PR         int
	MaxPRs     int
	Actor      string
}

// ParseIngestGitArgs parses flags for ingest git.
func ParseIngestGitArgs(args []string) (IngestGitArgs, error) {
	fs := flag.NewFlagSet("ingest git", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	path := fs.String("path", "", "local git directory")
	maxCommits := fs.Int("max-commits", 0, "max commits to process (0 = no cap)")
	actor := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	if err := fs.Parse(args); err != nil {
		return IngestGitArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	dir := strings.TrimSpace(*path)
	if key == "" {
		return IngestGitArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	if dir == "" {
		return IngestGitArgs{}, fmt.Errorf("validation_error: --path is required")
	}
	return IngestGitArgs{Repository: key, Path: dir, MaxCommits: *maxCommits, Actor: strings.TrimSpace(*actor)}, nil
}

// ParseIngestStatusArgs parses flags for ingest status.
func ParseIngestStatusArgs(args []string) (IngestStatusArgs, error) {
	fs := flag.NewFlagSet("ingest status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	kind := fs.String("kind", "git", "ingest kind: git or pr")
	if err := fs.Parse(args); err != nil {
		return IngestStatusArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return IngestStatusArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	k := strings.ToLower(strings.TrimSpace(*kind))
	if k == "" {
		k = "git"
	}
	if k != "git" && k != "pr" {
		return IngestStatusArgs{}, fmt.Errorf("validation_error: --kind must be git or pr")
	}
	return IngestStatusArgs{Repository: key, Kind: k}, nil
}

// ParseIngestPRArgs parses flags for ingest pr.
func ParseIngestPRArgs(args []string) (IngestPRArgs, error) {
	fs := flag.NewFlagSet("ingest pr", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	pr := fs.Int("pr", 0, "single pull request number")
	maxPRs := fs.Int("max-prs", 0, "max PRs to process (0 = no cap)")
	actor := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	if err := fs.Parse(args); err != nil {
		return IngestPRArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return IngestPRArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	if *pr < 0 {
		return IngestPRArgs{}, fmt.Errorf("validation_error: --pr must be non-negative")
	}
	return IngestPRArgs{Repository: key, PR: *pr, MaxPRs: *maxPRs, Actor: strings.TrimSpace(*actor)}, nil
}

// FormatIngestStatus renders a human-readable latest-run summary.
func FormatIngestStatus(repository string, run *domain.IngestRun) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n", repository)
	if run == nil {
		b.WriteString("Latest run: (none)\n")
		return b.String()
	}
	fmt.Fprintf(&b, "Latest run: %s\n", run.Status)
	fmt.Fprintf(&b, "  id: %s\n", run.ID)
	fmt.Fprintf(&b, "  commits seen: %d\n", run.CommitsSeen)
	fmt.Fprintf(&b, "  skipped: %d\n", run.CommitsSkipped)
	fmt.Fprintf(&b, "  candidates stored: %d\n", run.CandidatesStored)
	if run.CursorSHA != "" {
		cursor := run.CursorSHA
		if run.CursorAt != nil {
			cursor += " @ " + run.CursorAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(&b, "  cursor: %s\n", cursor)
	}
	if run.ErrorSummary != "" {
		fmt.Fprintf(&b, "  error: %s\n", run.ErrorSummary)
	}
	return b.String()
}

// FormatPRIngestStatus renders a human-readable latest PR-run summary.
func FormatPRIngestStatus(repository string, run *domain.PRIngestRun) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n", repository)
	if run == nil {
		b.WriteString("Latest PR run: (none)\n")
		return b.String()
	}
	fmt.Fprintf(&b, "Latest PR run: %s\n", run.Status)
	fmt.Fprintf(&b, "  id: %s\n", run.ID)
	fmt.Fprintf(&b, "  prs seen: %d\n", run.PRsSeen)
	fmt.Fprintf(&b, "  skipped: %d\n", run.PRsSkipped)
	fmt.Fprintf(&b, "  candidates stored: %d\n", run.CandidatesStored)
	if run.CursorPR > 0 {
		cursor := fmt.Sprintf("#%d", run.CursorPR)
		if run.CursorAt != nil {
			cursor += " @ " + run.CursorAt.UTC().Format(time.RFC3339)
		}
		fmt.Fprintf(&b, "  cursor: %s\n", cursor)
	}
	if run.ErrorSummary != "" {
		fmt.Fprintf(&b, "  error: %s\n", run.ErrorSummary)
	}
	return b.String()
}
