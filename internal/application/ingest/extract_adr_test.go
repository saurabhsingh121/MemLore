package ingest_test

import (
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/domain"
)

const acceptedMADR = `---
status: accepted
---

# Use PostgreSQL

## Context

We need a system of record.

## Decision

Use PostgreSQL as the system of record.

## Consequences

Transactions are available.

## Alternatives

SQLite was considered.
`

const nygardAccepted = `# 1. Record architecture decisions

## Status

Accepted

## Context

We need to record decisions.

## Decision

We will keep a collection of records for "architecturally significant" decisions.

## Consequences

See docs/adr.
`

func TestExtractADRAcceptedMADR(t *testing.T) {
	got := ingest.ExtractADR("docs/adr/0001-use-postgres.md", acceptedMADR)
	if !got.OK {
		t.Fatalf("skip %s", got.SkipReason)
	}
	if got.Snapshot.StatusClass != domain.ADRStatusAccepted {
		t.Fatalf("class = %q", got.Snapshot.StatusClass)
	}
	if got.Snapshot.ADRID != "0001-use-postgres" {
		t.Fatalf("id = %q", got.Snapshot.ADRID)
	}
	if !strings.Contains(got.Statement, "Use PostgreSQL as the system of record") {
		t.Fatalf("statement = %q", got.Statement)
	}
	if !strings.Contains(got.Statement, "Context:") {
		t.Fatalf("expected context in statement: %q", got.Statement)
	}
}

func TestExtractADRAcceptedNygard(t *testing.T) {
	got := ingest.ExtractADR("docs/adr/0001-record-architecture-decisions.md", nygardAccepted)
	if !got.OK || got.Snapshot.StatusClass != domain.ADRStatusAccepted {
		t.Fatalf("got ok=%v class=%s reason=%s", got.OK, got.Snapshot.StatusClass, got.SkipReason)
	}
}

func TestExtractADRSkipsDraftReadmeTemplate(t *testing.T) {
	draft := ingest.ExtractADR("docs/adr/0002-draft.md", "# Draft\n\n## Status\n\nDraft\n\n## Decision\n\nMaybe later.\n")
	if draft.OK || draft.SkipReason != domain.IngestSkipDraft {
		t.Fatalf("draft = %+v", draft)
	}
	readme := ingest.ExtractADR("docs/adr/README.md", "# ADRs\n")
	if readme.OK || readme.SkipReason != domain.IngestSkipReadme {
		t.Fatalf("readme = %+v", readme)
	}
	tmpl := ingest.ExtractADR("docs/adr/template.md", "# Template\n\n## Status\n\nProposed\n\n## Decision\n\nFill in.\n")
	if tmpl.OK || tmpl.SkipReason != domain.IngestSkipTemplate {
		t.Fatalf("template = %+v", tmpl)
	}
	placeholder := ingest.ExtractADR("docs/adr/NNNN-title.md", "# Title\n")
	if placeholder.OK || placeholder.SkipReason != domain.IngestSkipTemplate {
		t.Fatalf("placeholder = %+v", placeholder)
	}
}

func TestExtractADRSkipsRejectedAndUnknownAndNoDecision(t *testing.T) {
	rej := ingest.ExtractADR("adr/0003-no.md", "# No\n\n## Status\nRejected\n\n## Decision\nDo not.\n")
	if rej.OK || rej.SkipReason != domain.IngestSkipRejected {
		t.Fatalf("rejected = %+v", rej)
	}
	unknown := ingest.ExtractADR("adr/0004.md", "# X\n\n## Status\nmaybe\n\n## Decision\nY.\n")
	if unknown.OK || unknown.SkipReason != domain.IngestSkipUnknownStatus {
		t.Fatalf("unknown = %+v", unknown)
	}
	empty := ingest.ExtractADR("adr/0005.md", "# Empty\n\nHello world.\n")
	if empty.OK || empty.SkipReason != domain.IngestSkipNoDecision {
		t.Fatalf("no decision = %+v", empty)
	}
}

func TestExtractADRParsesSupersedesAndHistorical(t *testing.T) {
	body := `# Switch to SQLite

## Status

Accepted

## Supersedes

ADR-0003

## Decision

Use SQLite instead.
`
	got := ingest.ExtractADR("docs/adr/0007-sqlite.md", body)
	if !got.OK {
		t.Fatalf("skip %s", got.SkipReason)
	}
	found := false
	for _, id := range got.Snapshot.Supersedes {
		if id == "0003" || strings.HasPrefix(id, "0003") {
			found = true
		}
	}
	if !found {
		t.Fatalf("supersedes = %+v", got.Snapshot.Supersedes)
	}

	hist := ingest.ExtractADR("docs/adr/0003-old.md", "# Old\n\n## Status\nSuperseded\n\n## Decision\nOld choice.\n")
	if !hist.OK || hist.Snapshot.StatusClass != domain.ADRStatusHistorical {
		t.Fatalf("historical = %+v", hist)
	}
}

func TestExtractADRDoesNotInventDecision(t *testing.T) {
	got := ingest.ExtractADR("docs/adr/0008.md", "# Notes\n\n## Status\nAccepted\n\nNo decision heading here.\n")
	if got.OK {
		t.Fatalf("invented statement %q", got.Statement)
	}
}
