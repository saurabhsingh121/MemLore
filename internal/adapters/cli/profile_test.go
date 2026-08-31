package cli_test

import (
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/adapters/cli"
	"github.com/memlore/memlore/internal/adapters/presenters"
	"github.com/memlore/memlore/internal/domain"
)

func TestParseProfileArgsRequiresRepository(t *testing.T) {
	_, err := cli.ParseProfileArgs([]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseProfileArgsOK(t *testing.T) {
	args, err := cli.ParseProfileArgs([]string{"--repository", "github.com/acme/payments", "--token-budget", "2048"})
	if err != nil {
		t.Fatal(err)
	}
	if args.Repository != "github.com/acme/payments" || args.TokenBudget != 2048 {
		t.Fatalf("%+v", args)
	}
}

func TestFormatProfileIncludesSections(t *testing.T) {
	out := cli.FormatProfile(presenters.RepositoryProfile{
		Repository: presenters.Scope{Kind: domain.ScopeKindRepository, Key: "payments-api"},
		Sections: []presenters.ProfileSection{{
			ID: "decisions",
			Items: []presenters.ContextItem{{
				Statement: "Use Kafka instead of RabbitMQ.",
				Evidence:  []presenters.Evidence{{Type: domain.EvidenceTypeADR, Value: "ADR-017"}},
			}},
		}},
	})
	if !strings.Contains(out, "Repository: payments-api") {
		t.Fatalf("missing repo: %s", out)
	}
	if !strings.Contains(out, "Important decisions") {
		t.Fatalf("missing heading: %s", out)
	}
	if !strings.Contains(out, "Use Kafka instead of RabbitMQ.") {
		t.Fatalf("missing statement: %s", out)
	}
}
