package cli_test

import (
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/adapters/cli"
	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/domain"
)

func TestParseContextArgsRequiresTaskAndRepository(t *testing.T) {
	if _, err := cli.ParseContextArgs([]string{"--repository", "r1"}); err == nil {
		t.Fatal("expected error for missing task")
	}
	if _, err := cli.ParseContextArgs([]string{"--task", "do it"}); err == nil {
		t.Fatal("expected error for missing repository")
	}
}

func TestParseContextArgsOK(t *testing.T) {
	args, err := cli.ParseContextArgs([]string{
		"--task", "Implement outbox",
		"--repository", "github.com/acme/payments",
		"--ticket", "PAY-1842",
		"--changed-file", "src/payments/outbox.go",
		"--token-budget", "2048",
		"--agent-id", "cursor-agent",
	})
	if err != nil {
		t.Fatal(err)
	}
	if args.Task != "Implement outbox" || args.Repository != "github.com/acme/payments" {
		t.Fatalf("%+v", args)
	}
	if args.Ticket != "PAY-1842" || args.TokenBudget != 2048 || args.AgentID != "cursor-agent" {
		t.Fatalf("%+v", args)
	}
	if len(args.ChangedFiles) != 1 || args.ChangedFiles[0] != "src/payments/outbox.go" {
		t.Fatalf("files = %v", args.ChangedFiles)
	}
}

func TestFormatContextIncludesSectionsAndConflicts(t *testing.T) {
	out := cli.FormatContext(presenters.ContextPacket{
		Task:  "Implement outbox",
		Scope: presenters.Scope{Kind: domain.ScopeKindRepository, Key: "payments-api"},
		Sections: []presenters.ProfileSection{{
			ID: "architecture",
			Items: []presenters.ContextItem{{
				Statement: "Hexagonal architecture with ports.",
			}},
		}},
		Conflicts: []presenters.ConflictGroup{{
			Statements: []string{"Use blue-green", "Use rolling"},
		}},
		Sources: []presenters.Evidence{{Type: domain.EvidenceTypeADR, Value: "ADR-023"}},
	})
	if !strings.Contains(out, "Task: Implement outbox") {
		t.Fatalf("missing task: %s", out)
	}
	if !strings.Contains(out, "Relevant Architecture") {
		t.Fatalf("missing heading: %s", out)
	}
	if !strings.Contains(out, "Hexagonal architecture with ports.") {
		t.Fatalf("missing statement: %s", out)
	}
	if !strings.Contains(out, "Conflicts") || !strings.Contains(out, "Use blue-green") {
		t.Fatalf("missing conflicts: %s", out)
	}
	if !strings.Contains(out, "ADR-023") {
		t.Fatalf("missing sources: %s", out)
	}
}
