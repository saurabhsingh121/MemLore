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
	kind := fs.String("kind", "git", "ingest kind: git, pr, or adr")
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
	if k != "git" && k != "pr" && k != "adr" {
		return IngestStatusArgs{}, fmt.Errorf("validation_error: --kind must be git, pr, or adr")
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

// IngestADRArgs is parsed input for `memlore ingest adr`.
type IngestADRArgs struct {
	Repository string
	Path       string
	ADRDirs    []string
	Actor      string
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	v := strings.TrimSpace(value)
	if v != "" {
		*s = append(*s, v)
	}
	return nil
}

// ParseIngestADRArgs parses flags for ingest adr.
func ParseIngestADRArgs(args []string) (IngestADRArgs, error) {
	fs := flag.NewFlagSet("ingest adr", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	path := fs.String("path", "", "local working-copy directory")
	actor := fs.String("actor", "", "acting subject (or MEMLORE_ACTOR)")
	var extras stringList
	fs.Var(&extras, "adr-dir", "extra ADR directory relative to --path (repeatable)")
	if err := fs.Parse(args); err != nil {
		return IngestADRArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	dir := strings.TrimSpace(*path)
	if key == "" {
		return IngestADRArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	if dir == "" {
		return IngestADRArgs{}, fmt.Errorf("validation_error: --path is required")
	}
	return IngestADRArgs{Repository: key, Path: dir, ADRDirs: []string(extras), Actor: strings.TrimSpace(*actor)}, nil
}

// FormatADRIngestStatus renders a human-readable latest ADR-run summary.
func FormatADRIngestStatus(repository string, run *domain.ADRIngestRun) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n", repository)
	if run == nil {
		b.WriteString("Latest ADR run: (none)\n")
		return b.String()
	}
	fmt.Fprintf(&b, "Latest ADR run: %s\n", run.Status)
	fmt.Fprintf(&b, "  id: %s\n", run.ID)
	fmt.Fprintf(&b, "  files seen: %d\n", run.FilesSeen)
	fmt.Fprintf(&b, "  skipped: %d\n", run.FilesSkipped)
	fmt.Fprintf(&b, "  lore stored: %d\n", run.LoreStored)
	fmt.Fprintf(&b, "  superseded: %d\n", run.LoreSuperseded)
	if run.ErrorSummary != "" {
		fmt.Fprintf(&b, "  error: %s\n", run.ErrorSummary)
	}
	return b.String()
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
