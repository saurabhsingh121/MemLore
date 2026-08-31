package commands

import (
	"context"
	"strings"
	"time"

	appingest "github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

const maxPRPathEvidence = 8

// IngestPullRequestsCommand triggers GitHub PR capture for a repository scope.
type IngestPullRequestsCommand struct {
	Scope   domain.Scope
	ActorID string
	PR      int
	MaxPRs  int
}

// IngestPullRequestsHandler ingests observational lore from merged GitHub PRs.
type IngestPullRequestsHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
	prs   ports.PullRequestReader
}

func NewIngestPullRequestsHandler(begin ports.UnitOfWorkFactory, clock ports.Clock, prs ports.PullRequestReader) *IngestPullRequestsHandler {
	return &IngestPullRequestsHandler{begin: begin, clock: clock, prs: prs}
}

func (h *IngestPullRequestsHandler) Handle(ctx context.Context, cmd IngestPullRequestsCommand) (domain.PRIngestRun, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	if actor == "" {
		return domain.PRIngestRun{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	if h.prs == nil {
		return domain.PRIngestRun{}, &domain.ValidationError{Message: "pull request reader is not configured"}
	}

	now := h.clock.Now()
	run, err := domain.NewPRIngestRun(domain.NewPRIngestRunInput{
		Scope:    cmd.Scope,
		ActorID:  actor,
		PRNumber: cmd.PR,
		Now:      now,
	})
	if err != nil {
		return domain.PRIngestRun{}, err
	}

	if err := h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.PRIngest().InsertRun(ctx, run)
	}); err != nil {
		return domain.PRIngestRun{}, err
	}

	owner, repo, err := domain.GitHubRepoFromScopeKey(cmd.Scope.Key)
	if err != nil {
		return h.finishPR(ctx, run, false, err.Error())
	}

	var after *time.Time
	if cmd.PR == 0 {
		_ = h.withUoW(ctx, func(uow ports.UnitOfWork) error {
			cursor, ok, err := uow.PRIngest().GetCursor(ctx, cmd.Scope)
			if err != nil {
				return err
			}
			if ok {
				t := cursor.LastMergedAt
				after = &t
			}
			return nil
		})
	}

	snapshots, err := h.prs.ListPullRequests(ctx, ports.PullRequestQuery{
		Owner:         owner,
		Repo:          repo,
		AfterMergedAt: after,
		Number:        cmd.PR,
		MaxPRs:        cmd.MaxPRs,
	})
	if err != nil {
		return h.finishPR(ctx, run, false, err.Error())
	}

	for _, pr := range snapshots {
		run.PRsSeen++
		if err := h.processPR(ctx, cmd.Scope, actor, &run, pr); err != nil {
			return h.finishPR(ctx, run, false, err.Error())
		}
	}
	return h.finishPR(ctx, run, true, "")
}

func (h *IngestPullRequestsHandler) processPR(ctx context.Context, scope domain.Scope, actor string, run *domain.PRIngestRun, pr domain.PullRequestSnapshot) error {
	now := h.clock.Now()
	return h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		existing, ok, err := uow.PRIngest().GetProcessedPR(ctx, scope, pr.Number)
		if err != nil {
			return err
		}
		if ok {
			_ = existing
			h.advancePRCursor(ctx, uow, scope, pr, now, run)
			return nil
		}
		extracted := appingest.ExtractPRCandidate(pr)
		if !extracted.OK {
			run.PRsSkipped++
			if err := uow.PRIngest().InsertProcessedPR(ctx, domain.ProcessedPR{
				Scope:       scope,
				PRNumber:    pr.Number,
				NodeID:      pr.NodeID,
				Skipped:     true,
				SkipReason:  extracted.SkipReason,
				ProcessedAt: now,
			}); err != nil {
				if conflictAlreadyProcessed(err) {
					h.advancePRCursor(ctx, uow, scope, pr, now, run)
					return nil
				}
				return err
			}
			h.advancePRCursor(ctx, uow, scope, pr, now, run)
			return nil
		}

		evidence, err := evidenceForPR(pr, extracted)
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
		if err := uow.PRIngest().InsertProcessedPR(ctx, domain.ProcessedPR{
			Scope:       scope,
			PRNumber:    pr.Number,
			NodeID:      pr.NodeID,
			LoreEntryID: entry.ID,
			Skipped:     false,
			ProcessedAt: now,
		}); err != nil {
			if conflictAlreadyProcessed(err) {
				h.advancePRCursor(ctx, uow, scope, pr, now, run)
				return nil
			}
			return err
		}
		run.CandidatesStored++
		h.advancePRCursor(ctx, uow, scope, pr, now, run)
		return nil
	})
}

func (h *IngestPullRequestsHandler) advancePRCursor(ctx context.Context, uow ports.UnitOfWork, scope domain.Scope, pr domain.PullRequestSnapshot, now time.Time, run *domain.PRIngestRun) {
	run.CursorPR = pr.Number
	if pr.MergedAt != nil {
		t := pr.MergedAt.UTC()
		run.CursorAt = &t
		_ = uow.PRIngest().UpsertCursor(ctx, domain.PRIngestCursor{
			Scope:        scope,
			LastPR:       pr.Number,
			LastMergedAt: t,
			UpdatedAt:    now,
		})
	}
}

func (h *IngestPullRequestsHandler) finishPR(ctx context.Context, run domain.PRIngestRun, ok bool, errMsg string) (domain.PRIngestRun, error) {
	now := h.clock.Now()
	if ok {
		run = run.MarkSucceeded(now)
	} else {
		run = run.MarkFailed(now, errMsg)
	}
	_ = h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.PRIngest().UpdateRun(ctx, run)
	})
	return run, nil
}

func (h *IngestPullRequestsHandler) withUoW(ctx context.Context, fn func(ports.UnitOfWork) error) error {
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

func evidenceForPR(pr domain.PullRequestSnapshot, extracted appingest.PRExtractResult) ([]domain.EvidenceReference, error) {
	owner := pr.Owner
	repo := pr.Repo
	prRef, err := domain.NewEvidenceReference(domain.EvidenceTypePR, domain.PREvidenceValue(owner, repo, pr.Number))
	if err != nil {
		return nil, err
	}
	out := []domain.EvidenceReference{prRef}
	for _, u := range extracted.UsedCommentURLs {
		ref, err := domain.NewEvidenceReference(domain.EvidenceTypeURL, u)
		if err != nil {
			continue
		}
		out = append(out, ref)
	}
	for _, u := range appingest.LinkedIssueURLs(pr) {
		ref, err := domain.NewEvidenceReference(domain.EvidenceTypeURL, u)
		if err != nil {
			continue
		}
		out = append(out, ref)
	}
	for i, p := range pr.Files {
		if i >= maxPRPathEvidence {
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
