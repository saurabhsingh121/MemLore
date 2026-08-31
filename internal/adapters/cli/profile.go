package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/memlore/memlore/internal/adapters/presenters"
)

// ProfileArgs is parsed input for `memlore profile`.
type ProfileArgs struct {
	Repository  string
	TokenBudget int
}

// ParseProfileArgs parses CLI flags for the profile subcommand.
func ParseProfileArgs(args []string) (ProfileArgs, error) {
	fs := flag.NewFlagSet("profile", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repository", "", "repository scope key")
	budget := fs.Int("token-budget", 0, "token budget (default 4096)")
	if err := fs.Parse(args); err != nil {
		return ProfileArgs{}, err
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return ProfileArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	return ProfileArgs{Repository: key, TokenBudget: *budget}, nil
}

// FormatProfile renders a compact human-readable briefing.
func FormatProfile(profile presenters.RepositoryProfile) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Repository: %s\n", profile.Repository.Key)
	if len(profile.Sections) == 0 {
		b.WriteString("\n(no classified current knowledge)\n")
		return b.String()
	}
	for _, sec := range profile.Sections {
		fmt.Fprintf(&b, "\n%s\n%s\n", titleFor(sec.ID), strings.Repeat("-", len(titleFor(sec.ID))))
		for _, item := range sec.Items {
			fmt.Fprintf(&b, "%s\n", item.Statement)
			for _, ev := range item.Evidence {
				fmt.Fprintf(&b, "  evidence: %s %s\n", ev.Type, ev.Value)
			}
		}
	}
	if len(profile.Conflicts) > 0 {
		b.WriteString("\nConflicts\n---------\n")
		for _, c := range profile.Conflicts {
			for _, stmt := range c.Statements {
				fmt.Fprintf(&b, "- %s\n", stmt)
			}
		}
	}
	return b.String()
}

func titleFor(id string) string {
	switch id {
	case "decisions":
		return "Important decisions"
	case "conventions":
		return "Coding conventions"
	case "gotchas":
		return "Known gotchas"
	case "migrations":
		return "Current migration"
	case "ownership":
		return "Repository ownership"
	case "operational_risks":
		return "Operational risks"
	case "hotspots":
		return "Architectural hotspots"
	case "related_services":
		return "Related services and dependencies"
	case "architecture":
		return "Architecture"
	case "technologies":
		return "Major technologies"
	case "recent_changes":
		return "Recent important changes"
	default:
		return id
	}
}
