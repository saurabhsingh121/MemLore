package ingest_test

import (
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/domain"
)

func mergedAt() *time.Time {
	t := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	return &t
}

func TestExtractPRKeepsRationaleMerged(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 1842, Owner: "acme", Repo: "payments",
		Title:  "Use transactional outbox",
		Body:   "because dual-writes race on refunds",
		Merged: true, MergedAt: mergedAt(),
		AuthorLogin: "dev",
	}
	got := ingest.ExtractPRCandidate(pr)
	if !got.OK {
		t.Fatalf("skip = %s", got.SkipReason)
	}
	if !strings.Contains(got.Statement, "because dual-writes") {
		t.Fatalf("statement = %q", got.Statement)
	}
}

func TestExtractPRSkipsUnmerged(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 9, Owner: "acme", Repo: "payments",
		Title:  "Use outbox because dual-writes race",
		Merged: false, AuthorLogin: "dev",
	}
	got := ingest.ExtractPRCandidate(pr)
	if got.OK || got.SkipReason != domain.IngestSkipUnmerged {
		t.Fatalf("got %+v", got)
	}
}

func TestExtractPRSkipsDependabot(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 10, Owner: "acme", Repo: "payments",
		Title:  "chore(deps): bump lodash because security",
		Merged: true, MergedAt: mergedAt(),
		AuthorLogin: "dependabot[bot]", AuthorIsBot: true,
	}
	got := ingest.ExtractPRCandidate(pr)
	if got.OK || got.SkipReason != domain.IngestSkipBot {
		t.Fatalf("got %+v", got)
	}
}

func TestExtractPRSkipsNoisyChoreWithoutRationale(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 11, Owner: "acme", Repo: "payments",
		Title:  "chore: bump version",
		Merged: true, MergedAt: mergedAt(),
		AuthorLogin: "dev",
	}
	got := ingest.ExtractPRCandidate(pr)
	if got.OK || got.SkipReason != domain.IngestSkipNoisy {
		t.Fatalf("got %+v", got)
	}
}

func TestExtractPRKeepsNoisyTitleWhenReviewHasRationale(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 12, Owner: "acme", Repo: "payments",
		Title:  "chore: bump kafka client",
		Merged: true, MergedAt: mergedAt(),
		AuthorLogin: "dev",
		ReviewComments: []domain.PullRequestComment{{
			HTMLURL:     "https://github.com/acme/payments/pull/12#discussion_r1",
			Body:        "Keep 3.2 because 3.3 drops SASL.",
			AuthorLogin: "reviewer",
		}},
	}
	got := ingest.ExtractPRCandidate(pr)
	if !got.OK {
		t.Fatalf("skip = %s", got.SkipReason)
	}
	if !strings.Contains(got.Statement, "because 3.3") {
		t.Fatalf("statement = %q", got.Statement)
	}
	if len(got.UsedCommentURLs) != 1 {
		t.Fatalf("used = %+v", got.UsedCommentURLs)
	}
}

func TestExtractPRDoesNotInventFromPaths(t *testing.T) {
	pr := domain.PullRequestSnapshot{
		Number: 13, Owner: "acme", Repo: "payments",
		Title:  "update files",
		Merged: true, MergedAt: mergedAt(),
		AuthorLogin: "dev",
		Files:       []string{"internal/payments/outbox.go"},
	}
	got := ingest.ExtractPRCandidate(pr)
	if got.OK || got.SkipReason != domain.IngestSkipNoRationale {
		t.Fatalf("got %+v", got)
	}
}

func TestLinkedIssueURLsFromBody(t *testing.T) {
	pr := domain.PullRequestSnapshot{Owner: "acme", Repo: "payments", Body: "Fixes #381"}
	urls := ingest.LinkedIssueURLs(pr)
	if len(urls) != 1 || urls[0] != "https://github.com/acme/payments/issues/381" {
		t.Fatalf("urls = %+v", urls)
	}
}
