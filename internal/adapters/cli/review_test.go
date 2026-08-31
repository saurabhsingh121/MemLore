package cli_test

import (
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/adapters/cli"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
)

func TestParseReviewListArgsRequiresRepo(t *testing.T) {
	_, err := cli.ParseReviewListArgs(nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
	got, err := cli.ParseReviewListArgs([]string{"--repository", "github.com/acme/payments"})
	if err != nil || got.Repository != "github.com/acme/payments" {
		t.Fatalf("got %+v err=%v", got, err)
	}
}

func TestParseReviewAcceptArgs(t *testing.T) {
	_, err := cli.ParseReviewAcceptArgs(nil)
	if err == nil {
		t.Fatal("expected id error")
	}
	got, err := cli.ParseReviewAcceptArgs([]string{"abc-id", "--actor", "alice"})
	if err != nil || got.ID != "abc-id" || got.Actor != "alice" || got.HasEdit {
		t.Fatalf("got %+v err=%v", got, err)
	}
	edited, err := cli.ParseReviewAcceptArgs([]string{"abc-id", "--statement", "Edited rule", "--actor", "alice"})
	if err != nil || !edited.HasEdit || edited.Statement != "Edited rule" {
		t.Fatalf("edited %+v err=%v", edited, err)
	}
}

func TestParseReviewRejectArgs(t *testing.T) {
	_, err := cli.ParseReviewRejectArgs([]string{"--actor", "alice"})
	if err == nil {
		t.Fatal("expected id error")
	}
	got, err := cli.ParseReviewRejectArgs([]string{"abc-id", "--actor", "alice"})
	if err != nil || got.ID != "abc-id" || got.Actor != "alice" {
		t.Fatalf("got %+v err=%v", got, err)
	}
}

func TestFormatReviewListOmitsConfidenceAndReason(t *testing.T) {
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	ev := domain.EvidenceReference{Type: domain.EvidenceTypePR, Value: "acme/payments#1842"}
	out := cli.FormatReviewList("github.com/acme/payments", []queries.SuggestedLoreItem{{
		Entry: domain.LoreEntry{
			ID:        "id-1",
			Statement: "Payment events use transactional outbox.",
			Scope:     scope,
			Evidence:  []domain.EvidenceReference{ev},
			CreatedAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
		},
		SourceType: "pr",
		Status:     queries.ReviewStatusPending,
	}})
	if !strings.Contains(out, "Suggested Lore (github.com/acme/payments)") {
		t.Fatalf("header: %s", out)
	}
	if !strings.Contains(out, "source: pr") || !strings.Contains(out, "pr acme/payments#1842") {
		t.Fatalf("evidence/source: %s", out)
	}
	if strings.Contains(out, "confidence") || strings.Contains(out, "reason") {
		t.Fatalf("invented fields: %s", out)
	}
	empty := cli.FormatReviewList("r1", nil)
	if !strings.Contains(empty, "(none pending)") {
		t.Fatalf("empty: %s", empty)
	}
}
