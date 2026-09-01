package cli_test

import (
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/adapters/cli"
	"github.com/memlore/memlore/internal/domain"
)

func TestParseDecisionCreateArgs(t *testing.T) {
	_, err := cli.ParseDecisionCreateArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	got, err := cli.ParseDecisionCreateArgs([]string{
		"--repository", "github.com/acme/payments",
		"--question", "How should payment events be published?",
		"--choice", "Transactional outbox",
		"--owner", "alice",
		"--alternative", "Dual-write",
		"--actor", "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != "github.com/acme/payments" || got.Choice != "Transactional outbox" || len(got.Alternatives) != 1 {
		t.Fatalf("got %+v", got)
	}
}

func TestParseDecisionListAndGetArgs(t *testing.T) {
	_, err := cli.ParseDecisionListArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	list, err := cli.ParseDecisionListArgs([]string{"--repository", "github.com/acme/payments"})
	if err != nil || list.Repository != "github.com/acme/payments" {
		t.Fatalf("list %+v err=%v", list, err)
	}
	_, err = cli.ParseDecisionGetArgs(nil)
	if err == nil {
		t.Fatal("expected id error")
	}
	get, err := cli.ParseDecisionGetArgs([]string{"abc-id"})
	if err != nil || get.ID != "abc-id" {
		t.Fatalf("get %+v err=%v", get, err)
	}
}

func TestFormatDecisionList(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	d, _ := domain.NewHumanDecision(domain.NewDecisionInput{
		ID: "d1", Scope: scope, Question: "How should payment events be published?",
		Choice: "Transactional outbox", Owner: "alice", CreatedBy: "alice",
	})
	out := cli.FormatDecisionList("github.com/acme/payments", []domain.Decision{d})
	if !strings.Contains(out, "Transactional outbox") || !strings.Contains(out, "source: human") {
		t.Fatalf("out = %s", out)
	}
}
