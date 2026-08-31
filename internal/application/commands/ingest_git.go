package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	appingest "github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

const maxPathEvidence = 8

// IngestGitCommand triggers git commit capture for a repository scope.
type IngestGitCommand struct {
	Scope      domain.Scope
	Path       string
	ActorID    string
	MaxCommits int
}

// IngestGitHandler ingests observational lore from a local git directory.
type IngestGitHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
	git   ports.GitReader
}

func NewIngestGitHandler(begin ports.UnitOfWorkFactory, clock ports.Clock, git ports.GitReader) *IngestGitHandler {
	return &IngestGitHandler{begin: begin, clock: clock, git: git}
}

func (h *IngestGitHandler) Handle(ctx context.Context, cmd IngestGitCommand) (domain.IngestRun, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	path := strings.TrimSpace(cmd.Path)
	if actor == "" {
		return domain.IngestRun{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	if path == "" {
		return domain.IngestRun{}, &domain.ValidationError{Message: "path must be non-empty"}
	}
	if h.git == nil {
		return domain.IngestRun{}, &domain.ValidationError{Message: "git reader is not configured"}
	}

	now := h.clock.Now()
	run, err := domain.NewIngestRun(domain.NewIngestRunInput{
		Scope:     cmd.Scope,
		ActorID:   actor,
		LocalPath: path,
		Now:       now,
	})
	if err != nil {
		return domain.IngestRun{}, err
	}

	if err := h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.Ingest().InsertRun(ctx, run)
	}); err != nil {
		return domain.IngestRun{}, err
	}

	var after *time.Time
	_ = h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		cursor, ok, err := uow.Ingest().GetCursor(ctx, cmd.Scope)
		if err != nil {
			return err
		}
		if ok {
			t := cursor.LastCommittedAt
			after = &t
		}
		return nil
	})

	commits, err := h.git.ListCommits(ctx, ports.GitLogQuery{
		Path:       path,
		After:      after,
		MaxCommits: cmd.MaxCommits,
	})
	if err != nil {
		return h.finish(ctx, run, false, err.Error())
	}

	for _, commit := range commits {
		run.CommitsSeen++
		if err := h.processCommit(ctx, cmd.Scope, actor, &run, commit); err != nil {
			return h.finish(ctx, run, false, err.Error())
		}
	}
	return h.finish(ctx, run, true, "")
}

func (h *IngestGitHandler) processCommit(ctx context.Context, scope domain.Scope, actor string, run *domain.IngestRun, commit domain.GitCommitSnapshot) error {
	now := h.clock.Now()
	return h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		existing, ok, err := uow.Ingest().GetProcessedSHA(ctx, scope, commit.SHA)
		if err != nil {
			return err
		}
		if ok {
			_ = existing
			h.advanceCursor(ctx, uow, scope, commit, now, run)
			return nil
		}
		extracted := appingest.ExtractCandidate(commit)
		if !extracted.OK {
			run.CommitsSkipped++
			if err := uow.Ingest().InsertProcessedSHA(ctx, domain.ProcessedSHA{
				Scope:       scope,
				SHA:         commit.SHA,
				Skipped:     true,
				SkipReason:  extracted.SkipReason,
				ProcessedAt: now,
			}); err != nil {
				if conflictAlreadyProcessed(err) {
					h.advanceCursor(ctx, uow, scope, commit, now, run)
					return nil
				}
				return err
			}
			h.advanceCursor(ctx, uow, scope, commit, now, run)
			return nil
		}

		evidence, err := evidenceForCommit(commit)
		if err != nil {
			return err
		}
		entry, err := domain.NewObservationalLoreEntry(domain.NewLoreEntryInput{
			Statement: extracted.Statement,
			Scope:     scope,
			CreatedBy: actor,
			Evidence:  evidence,
			Now:       now,
		})
		if err != nil {
			return err
		}
		audit, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
			TargetID:  entry.ID,
			Action:    domain.AuditActionCreate,
			ActorID:   actor,
			CreatedAt: now,
		})
		if err != nil {
			return err
		}
		outboxEvent, err := domain.NewEpisodeIngestOutboxEvent(entry, now)
		if err != nil {
			return err
		}
		if err := uow.LoreEntries().Add(ctx, entry); err != nil {
			return err
		}
		if err := uow.Audits().Add(ctx, audit); err != nil {
			return err
		}
		if err := uow.Outbox().Add(ctx, outboxEvent); err != nil {
			return err
		}
		if err := uow.Ingest().InsertProcessedSHA(ctx, domain.ProcessedSHA{
			Scope:       scope,
			SHA:         commit.SHA,
			LoreEntryID: entry.ID,
			Skipped:     false,
			ProcessedAt: now,
		}); err != nil {
			if conflictAlreadyProcessed(err) {
				h.advanceCursor(ctx, uow, scope, commit, now, run)
				return nil
			}
			return err
		}
		run.CandidatesStored++
		h.advanceCursor(ctx, uow, scope, commit, now, run)
		return nil
	})
}

func (h *IngestGitHandler) advanceCursor(ctx context.Context, uow ports.UnitOfWork, scope domain.Scope, commit domain.GitCommitSnapshot, now time.Time, run *domain.IngestRun) {
	run.CursorSHA = commit.SHA
	t := commit.CommittedAt.UTC()
	run.CursorAt = &t
	_ = uow.Ingest().UpsertCursor(ctx, domain.IngestCursor{
		Scope:           scope,
		LastSHA:         commit.SHA,
		LastCommittedAt: t,
		UpdatedAt:       now,
	})
}

func (h *IngestGitHandler) finish(ctx context.Context, run domain.IngestRun, ok bool, errMsg string) (domain.IngestRun, error) {
	now := h.clock.Now()
	if ok {
		run = run.MarkSucceeded(now)
	} else {
		run = run.MarkFailed(now, errMsg)
	}
	_ = h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.Ingest().UpdateRun(ctx, run)
	})
	if !ok {
		return run, nil
	}
	return run, nil
}

func (h *IngestGitHandler) withUoW(ctx context.Context, fn func(ports.UnitOfWork) error) error {
	uow, err := h.begin(ctx)
	if err != nil {
		return err
	}
	defer uow.Rollback(ctx)
	if err := fn(uow); err != nil {
		return err
	}
	return uow.Commit(ctx)
}

func evidenceForCommit(commit domain.GitCommitSnapshot) ([]domain.EvidenceReference, error) {
	commitRef, err := domain.NewEvidenceReference(domain.EvidenceTypeCommit, commit.SHA)
	if err != nil {
		return nil, err
	}
	out := []domain.EvidenceReference{commitRef}
	for i, p := range commit.Paths {
		if i >= maxPathEvidence {
			break
		}
		ref, err := domain.NewEvidenceReference(domain.EvidenceTypePath, p)
		if err != nil {
			continue
		}
		out = append(out, ref)
	}
	return out, nil
}

func conflictAlreadyProcessed(err error) bool {
	var ce *domain.ConflictError
	return errors.As(err, &ce)
}
