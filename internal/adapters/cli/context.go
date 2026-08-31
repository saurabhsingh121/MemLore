package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/memlore/memlore/internal/adapters/presenters"
)

// ContextArgs is parsed input for `memlore context`.
type ContextArgs struct {
	Task         string
	Repository   string
	Query        string
	Ticket       string
	Branch       string
	ChangedFiles []string
	WorkingFiles []string
	TokenBudget  int
	AgentID      string
}

// ParseContextArgs parses CLI flags for the context subcommand.
func ParseContextArgs(args []string) (ContextArgs, error) {
	fs := flag.NewFlagSet("context", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	task := fs.String("task", "", "task description")
	repo := fs.String("repository", "", "repository scope key")
	query := fs.String("query", "", "search query (defaults to task)")
	ticket := fs.String("ticket", "", "ticket identifier")
	branch := fs.String("branch", "", "git branch (echoed; not a filter)")
	budget := fs.Int("token-budget", 0, "token budget (default 4096)")
	agentID := fs.String("agent-id", "", "agent identity (provenance only)")
	var changed, working []string
	fs.Func("changed-file", "changed file path (repeatable)", func(s string) error {
		if t := strings.TrimSpace(s); t != "" {
			changed = append(changed, t)
		}
		return nil
	})
	fs.Func("working-file", "working file path (repeatable)", func(s string) error {
		if t := strings.TrimSpace(s); t != "" {
			working = append(working, t)
		}
		return nil
	})
	if err := fs.Parse(args); err != nil {
		return ContextArgs{}, err
	}
	taskVal := strings.TrimSpace(*task)
	if taskVal == "" {
		return ContextArgs{}, fmt.Errorf("validation_error: --task is required")
	}
	key := strings.TrimSpace(*repo)
	if key == "" {
		return ContextArgs{}, fmt.Errorf("validation_error: --repository is required")
	}
	return ContextArgs{
		Task:         taskVal,
		Repository:   key,
		Query:        strings.TrimSpace(*query),
		Ticket:       strings.TrimSpace(*ticket),
		Branch:       strings.TrimSpace(*branch),
		ChangedFiles: changed,
		WorkingFiles: working,
		TokenBudget:  *budget,
		AgentID:      strings.TrimSpace(*agentID),
	}, nil
}

// FormatContext renders a compact human-readable task briefing.
func FormatContext(packet presenters.ContextPacket) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n", packet.Task)
	fmt.Fprintf(&b, "Repository: %s\n", packet.Scope.Key)
	if packet.Ticket != "" {
		fmt.Fprintf(&b, "Ticket: %s\n", packet.Ticket)
	}
	if len(packet.Sections) == 0 {
		b.WriteString("\n(no classified current knowledge)\n")
	} else {
		for _, sec := range packet.Sections {
			title := packetTitleFor(sec.ID)
			fmt.Fprintf(&b, "\n%s\n%s\n", title, strings.Repeat("-", len(title)))
			for _, item := range sec.Items {
				fmt.Fprintf(&b, "%s\n", item.Statement)
				for _, ev := range item.Evidence {
					fmt.Fprintf(&b, "  evidence: %s %s\n", ev.Type, ev.Value)
				}
			}
		}
	}
	if len(packet.Conflicts) > 0 {
		b.WriteString("\nConflicts\n---------\n")
		for _, c := range packet.Conflicts {
			for _, stmt := range c.Statements {
				fmt.Fprintf(&b, "- %s\n", stmt)
			}
		}
	}
	if len(packet.Sources) > 0 {
		b.WriteString("\nEvidence / Sources\n------------------\n")
		for _, ev := range packet.Sources {
			fmt.Fprintf(&b, "- %s %s\n", ev.Type, ev.Value)
		}
	}
	return b.String()
}

func packetTitleFor(id string) string {
	switch id {
	case "architecture":
		return "Relevant Architecture"
	case "decisions":
		return "Applicable Decisions"
	case "conventions":
		return "Coding Conventions"
	case "task_context":
		return "Task Context"
	case "gotchas":
		return "Known Gotchas"
	default:
		return id
	}
}
