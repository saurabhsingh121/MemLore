package ingest_test

import (
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/domain"
)

func TestExtractKeepsRationaleCommit(t *testing.T) {
	got := ingest.ExtractCandidate(domain.GitCommitSnapshot{
		SHA:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Subject: "feat: add payment outbox",
		Body:    "Use the transactional outbox because dual-writes race.",
		Message: "feat: add payment outbox\n\nUse the transactional outbox because dual-writes race.",
	})
	if !got.OK {
		t.Fatalf("expected candidate, got skip %q", got.SkipReason)
	}
	if !strings.Contains(got.Statement, "because dual-writes") {
		t.Fatalf("statement = %q", got.Statement)
	}
}

func TestExtractSkipsMergeAndChoreAndEmpty(t *testing.T) {
	cases := []struct {
		name string
		c    domain.GitCommitSnapshot
		want string
	}{
		{
			name: "merge parents",
			c:    domain.GitCommitSnapshot{ParentCount: 2, Subject: "Merge branch 'x'", Message: "Merge branch 'x'"},
			want: domain.IngestSkipMerge,
		},
		{
			name: "merge subject",
			c:    domain.GitCommitSnapshot{Subject: "Merge pull request #12", Message: "Merge pull request #12"},
			want: domain.IngestSkipMerge,
		},
		{
			name: "chore bump",
			c:    domain.GitCommitSnapshot{Subject: "chore: bump version", Message: "chore: bump version"},
			want: domain.IngestSkipNoisy,
		},
		{
			name: "version tag",
			c:    domain.GitCommitSnapshot{Subject: "v1.2.3", Message: "v1.2.3"},
			want: domain.IngestSkipNoisy,
		},
		{
			name: "empty",
			c:    domain.GitCommitSnapshot{},
			want: domain.IngestSkipEmpty,
		},
		{
			name: "no rationale",
			c:    domain.GitCommitSnapshot{Subject: "feat: add helper", Message: "feat: add helper"},
			want: domain.IngestSkipNoRationale,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ingest.ExtractCandidate(tc.c)
			if got.OK || got.SkipReason != tc.want {
				t.Fatalf("got %+v want skip %q", got, tc.want)
			}
		})
	}
}

func TestExtractSkipWinsOverCue(t *testing.T) {
	got := ingest.ExtractCandidate(domain.GitCommitSnapshot{
		Subject: "chore: cleanup",
		Body:    "because the tests were flaky",
		Message: "chore: cleanup\n\nbecause the tests were flaky",
	})
	if got.OK || got.SkipReason != domain.IngestSkipNoisy {
		t.Fatalf("got %+v", got)
	}
}

func TestExtractDoesNotInventFromPaths(t *testing.T) {
	got := ingest.ExtractCandidate(domain.GitCommitSnapshot{
		Subject: "wip",
		Message: "wip",
		Paths:   []string{"docs/adr/0001-architecture.md", "internal/payments/outbox.go"},
	})
	if got.OK {
		t.Fatal("paths must not invent a candidate")
	}
}

func TestExtractTooLongSkipped(t *testing.T) {
	msg := strings.Repeat("because ", domain.MaxStatementLength)
	got := ingest.ExtractCandidate(domain.GitCommitSnapshot{Subject: "feat: x", Message: msg, Body: msg})
	if got.OK || got.SkipReason != domain.IngestSkipTooLong {
		t.Fatalf("got %+v", got)
	}
}
