package commands_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

type stubADR struct {
	files []ports.ADRFile
	err   error
	query ports.ADRListQuery
}

func (s *stubADR) ListADRFiles(_ context.Context, q ports.ADRListQuery) ([]ports.ADRFile, error) {
	s.query = q
	if s.err != nil {
		return nil, s.err
	}
	return s.files, nil
}

func acceptedADRFile(path, checksum, decision string) ports.ADRFile {
	body := "# Title\n\n## Status\n\nAccepted\n\n## Decision\n\n" + decision + "\n"
	return ports.ADRFile{RelativePath: path, Checksum: checksum, Content: body}
}

func adrFixture(t *testing.T) (domain.Scope, time.Time, *memory.UnitOfWork, *commands.IngestADRsHandler, *stubADR) {
	t.Helper()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/payments")
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	uow := memory.NewUnitOfWork()
	adrs := &stubADR{files: []ports.ADRFile{
		acceptedADRFile("docs/adr/0001-use-postgres.md", "sum-1", "Use PostgreSQL as the system of record."),
		{RelativePath: "docs/adr/0002-draft.md", Checksum: "sum-2", Content: "# Draft\n\n## Status\nDraft\n\n## Decision\nMaybe.\n"},
		{RelativePath: "docs/adr/README.md", Checksum: "sum-r", Content: "# ADRs\n"},
		{RelativePath: "docs/adr/template.md", Checksum: "sum-t", Content: "# Template\n\n## Status\nProposed\n\n## Decision\nFill in.\n"},
	}}
	handler := commands.NewIngestADRsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, adrs)
	return scope, now, uow, handler, adrs
}

func TestIngestADRStoresOneVerifiedArchitectureDecision(t *testing.T) {
	scope, _, uow, handler, _ := adrFixture(t)
	run, err := handler.Handle(context.Background(), commands.IngestADRsCommand{
		Scope: scope, Path: "/tmp/payments", ActorID: "alice",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if run.Status != domain.IngestRunSucceeded {
		t.Fatalf("status = %q err=%s", run.Status, run.ErrorSummary)
	}
	if run.FilesSeen != 4 || run.LoreStored != 1 {
		t.Fatalf("seen=%d stored=%d skipped=%d", run.FilesSeen, run.LoreStored, run.FilesSkipped)
	}
	entries, err := uow.LoreEntries().ListByScope(context.Background(), scope)
	if err != nil || len(entries) != 1 {
		t.Fatalf("lore = %d err=%v", len(entries), err)
	}
	e := entries[0]
	if e.Origin != domain.KnowledgeOriginArchitectureDecision {
		t.Fatalf("origin = %q", e.Origin)
	}
	if e.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", e.VerificationStatus)
	}
	if e.Origin == domain.KnowledgeOriginHumanAuthored || e.Origin == domain.KnowledgeOriginRepositoryObservation {
		t.Fatal("must not be human or observational")
	}
	hasADR := false
	for _, ev := range e.Evidence {
		if ev.Type == domain.EvidenceTypeADR && ev.Value == "0001-use-postgres" {
			hasADR = true
		}
	}
	if !hasADR {
		t.Fatalf("evidence = %+v", e.Evidence)
	}
	if len(uow.Outbox().(*memory.OutboxRepository).ListPending()) != 1 {
		t.Fatal("expected one outbox event")
	}
}

func TestIngestADRIsIdempotentOnSameChecksum(t *testing.T) {
	scope, _, uow, handler, _ := adrFixture(t)
	cmd := commands.IngestADRsCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	run, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if run.LoreStored != 0 {
		t.Fatalf("second run stored %d", run.LoreStored)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 {
		t.Fatalf("lore count = %d", len(entries))
	}
}

func TestIngestADRChecksumChangeSupersedesPriorIngestLore(t *testing.T) {
	scope, _, uow, handler, adrs := adrFixture(t)
	cmd := commands.IngestADRsCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	first, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	adrs.files[0] = acceptedADRFile("docs/adr/0001-use-postgres.md", "sum-1-b", "Use PostgreSQL with JSONB.")
	run, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if run.LoreStored != 1 || run.LoreSuperseded != 1 {
		t.Fatalf("run stored=%d superseded=%d", run.LoreStored, run.LoreSuperseded)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 2 {
		t.Fatalf("len = %d", len(entries))
	}
	current := 0
	for _, e := range entries {
		if domain.IsCurrent(e) {
			current++
			if !strings.Contains(e.Statement, "JSONB") {
				t.Fatalf("current statement = %q", e.Statement)
			}
		}
	}
	if current != 1 {
		t.Fatalf("current = %d", current)
	}
	old, err := uow.LoreEntries().Get(context.Background(), first[0].ID)
	if err != nil || !domain.IsSuperseded(old) {
		t.Fatalf("predecessor must remain superseded: %+v err=%v", old, err)
	}
}

func TestIngestADRRetryDoesNotDuplicate(t *testing.T) {
	scope, _, uow, handler, _ := adrFixture(t)
	cmd := commands.IngestADRsCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	if _, err := handler.Handle(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	current := 0
	for _, e := range entries {
		if domain.IsCurrent(e) {
			current++
		}
	}
	if current != 1 {
		t.Fatalf("current = %d total=%d", current, len(entries))
	}
}

func TestIngestADRExplicitSupersedesIngestCreatedOnly(t *testing.T) {
	scope, now, uow, _, _ := adrFixture(t)
	oldFile := acceptedADRFile("docs/adr/0003-old.md", "old", "Use MySQL.")
	handler := commands.NewIngestADRsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, &stubADR{files: []ports.ADRFile{oldFile}})
	if _, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/p", ActorID: "alice"}); err != nil {
		t.Fatal(err)
	}

	humanEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0003")
	human, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Human ADR note", Scope: scope, CreatedBy: "bob", Evidence: []domain.EvidenceReference{humanEv}, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), human); err != nil {
		t.Fatal(err)
	}

	successor := ports.ADRFile{
		RelativePath: "docs/adr/0007-sqlite.md",
		Checksum:     "new",
		Content:      "# Switch\n\n## Status\nAccepted\n\n## Supersedes\nADR-0003\n\n## Decision\nUse SQLite.\n",
	}
	handler = commands.NewIngestADRsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, &stubADR{files: []ports.ADRFile{successor}})
	if _, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/p", ActorID: "alice"}); err != nil {
		t.Fatal(err)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	var ingestOld, ingestNew, humanGot domain.LoreEntry
	for _, e := range entries {
		switch {
		case e.ID == human.ID:
			humanGot = e
		case e.Origin == domain.KnowledgeOriginArchitectureDecision && strings.Contains(e.Statement, "MySQL"):
			ingestOld = e
		case e.Origin == domain.KnowledgeOriginArchitectureDecision && strings.Contains(e.Statement, "SQLite"):
			ingestNew = e
		}
	}
	if ingestNew.ID == "" || !domain.IsCurrent(ingestNew) {
		t.Fatalf("successor current = %+v", ingestNew)
	}
	if ingestOld.ID == "" || !domain.IsSuperseded(ingestOld) {
		t.Fatalf("ingest predecessor should be superseded: %+v", ingestOld)
	}
	if !domain.IsCurrent(humanGot) {
		t.Fatalf("human lore must remain current: %+v", humanGot)
	}
}

func TestIngestADRDeprecatedIsNotCurrent(t *testing.T) {
	scope, now, uow, _, _ := adrFixture(t)
	handler := commands.NewIngestADRsHandler(memory.BeginFactory(uow), clock.FixedClock{Instant: now}, &stubADR{files: []ports.ADRFile{{
		RelativePath: "docs/adr/0009-old.md",
		Checksum:     "d",
		Content:      "# Old\n\n## Status\nDeprecated\n\n## Decision\nDo not use this.\n",
	}}})
	run, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/p", ActorID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if run.LoreStored != 1 {
		t.Fatalf("stored = %d", run.LoreStored)
	}
	entries, _ := uow.LoreEntries().ListByScope(context.Background(), scope)
	if len(entries) != 1 || domain.IsCurrent(entries[0]) {
		t.Fatalf("historical must exist but not be current: %+v", entries)
	}
	if entries[0].Origin != domain.KnowledgeOriginArchitectureDecision {
		t.Fatalf("origin = %q", entries[0].Origin)
	}
}

func TestIngestADRDoesNotUpgradeGitOrPR(t *testing.T) {
	scope, now, uow, handler, _ := adrFixture(t)
	gitEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeCommit, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	obs, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use the outbox because dual-writes race.", Scope: scope, CreatedBy: "alice", Evidence: []domain.EvidenceReference{gitEv}, Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.LoreEntries().Add(context.Background(), obs); err != nil {
		t.Fatal(err)
	}
	if _, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/p", ActorID: "alice"}); err != nil {
		t.Fatal(err)
	}
	got, err := uow.LoreEntries().Get(context.Background(), obs.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Origin != domain.KnowledgeOriginRepositoryObservation || got.VerificationStatus != domain.VerificationUnverified {
		t.Fatalf("observation upgraded: %+v", got)
	}
}

func TestIngestADRConflictWhenAlreadyRunning(t *testing.T) {
	scope, now, uow, handler, _ := adrFixture(t)
	running, err := domain.NewADRIngestRun(domain.NewADRIngestRunInput{
		Scope: scope, ActorID: "bob", LocalPath: "/tmp/payments", Now: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := uow.ADRIngest().InsertRun(context.Background(), running); err != nil {
		t.Fatal(err)
	}
	_, err = handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/payments", ActorID: "alice"})
	var ce *domain.ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v", err)
	}
}

func TestIngestADRRecordsFailedRunWhenPathMissing(t *testing.T) {
	scope, _, _, handler, adrs := adrFixture(t)
	adrs.err = &ports.PathNotDirectoryError{Path: "/tmp/missing"}
	run, err := handler.Handle(context.Background(), commands.IngestADRsCommand{
		Scope: scope, Path: "/tmp/missing", ActorID: "alice",
	})
	if err != nil {
		t.Fatalf("failed run should not bubble, got %v", err)
	}
	if run.Status != domain.IngestRunFailed || run.ErrorSummary == "" || run.LoreStored != 0 {
		t.Fatalf("run = %+v", run)
	}
}

func TestIngestADRPassesExtraDirs(t *testing.T) {
	scope, _, _, handler, adrs := adrFixture(t)
	_, err := handler.Handle(context.Background(), commands.IngestADRsCommand{
		Scope: scope, Path: "/tmp/p", ActorID: "alice", ExtraDirs: []string{"architecture/records"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(adrs.query.Dirs) != 1 || adrs.query.Dirs[0] != "architecture/records" {
		t.Fatalf("query = %+v", adrs.query)
	}
}

func TestIngestADRRequiresActorAndPath(t *testing.T) {
	scope, _, _, handler, _ := adrFixture(t)
	if _, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, Path: "/tmp/x"}); err == nil {
		t.Fatal("expected actor error")
	}
	if _, err := handler.Handle(context.Background(), commands.IngestADRsCommand{Scope: scope, ActorID: "alice"}); err == nil {
		t.Fatal("expected path error")
	}
}
