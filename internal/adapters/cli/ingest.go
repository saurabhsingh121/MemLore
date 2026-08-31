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
	if err := fs.Parse(args); err != nil {
		return IngestStatusArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return IngestStatusArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	return IngestStatusArgs{Repository: key}, nil
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
