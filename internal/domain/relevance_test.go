package domain_test

import (
	"testing"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

func TestStatementMatchesQuery(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, statement, query string
		want                   bool
	}{
		{"substring", "Use transactional outbox for payments.", "outbox", true},
		{"case", "Use Outbox", "OUTBOX", true},
		{"all tokens", "Payment events use transactional outbox.", "payment outbox", true},
		{"missing token", "Deploy with blue green.", "payment outbox", false},
		{"any token from task", "Use blue-green deploys.", "choose deploy strategy", true},
		{"empty query", "anything", "  ", false},
		{"short tokens ignored needs real", "Use a transactional outbox.", "a outbox", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := domain.StatementMatchesQuery(tc.statement, tc.query)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestSortLoreByRelevance(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	entries := []domain.LoreEntry{
		{ID: "off", Statement: "Unrelated deploy notes", CreatedAt: now},
		{ID: "on", Statement: "Transactional outbox required", CreatedAt: now.Add(-time.Hour)},
		{ID: "on-verified", Statement: "Use outbox pattern", VerificationStatus: domain.VerificationVerified, CreatedAt: now.Add(-2 * time.Hour)},
	}
	domain.SortLoreByRelevance(entries, "outbox")
	if entries[0].ID != "on-verified" {
		t.Fatalf("order[0]=%s want on-verified (verified tie-break among matches)", entries[0].ID)
	}
	if !domain.StatementMatchesQuery(entries[0].Statement, "outbox") {
		t.Fatal("top should match")
	}
}
