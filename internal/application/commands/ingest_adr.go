package commands

import (
	"context"
	"errors"
	"sort"
	"strings"

	appingest "github.com/memlore/memlore/internal/application/ingest"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// IngestADRsCommand triggers ADR capture for a repository scope.
type IngestADRsCommand struct {
	Scope     domain.Scope
	Path      string
	ActorID   string
	ExtraDirs []string
}

// IngestADRsHandler ingests trusted-source lore from local ADR files.
type IngestADRsHandler struct {
	begin ports.UnitOfWorkFactory
	clock ports.Clock
	adrs  ports.ADRReader
}

func NewIngestADRsHandler(begin ports.UnitOfWorkFactory, clock ports.Clock, adrs ports.ADRReader) *IngestADRsHandler {
	return &IngestADRsHandler{begin: begin, clock: clock, adrs: adrs}
}

func (h *IngestADRsHandler) Handle(ctx context.Context, cmd IngestADRsCommand) (domain.ADRIngestRun, error) {
	actor := strings.TrimSpace(cmd.ActorID)
	path := strings.TrimSpace(cmd.Path)
	if actor == "" {
		return domain.ADRIngestRun{}, &domain.ValidationError{Message: "actor must be non-empty"}
	}
	if path == "" {
		return domain.ADRIngestRun{}, &domain.ValidationError{Message: "path must be non-empty"}
	}
	if h.adrs == nil {
		return domain.ADRIngestRun{}, &domain.ValidationError{Message: "adr reader is not configured"}
	}

	now := h.clock.Now()
	run, err := domain.NewADRIngestRun(domain.NewADRIngestRunInput{
		Scope:     cmd.Scope,
		ActorID:   actor,
		LocalPath: path,
		ExtraDirs: cmd.ExtraDirs,
		Now:       now,
	})
	if err != nil {
		return domain.ADRIngestRun{}, err
	}

	if err := h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.ADRIngest().InsertRun(ctx, run)
	}); err != nil {
		return domain.ADRIngestRun{}, err
	}

	files, err := h.adrs.ListADRFiles(ctx, ports.ADRListQuery{Path: path, Dirs: cmd.ExtraDirs})
	if err != nil {
		return h.finish(ctx, run, false, err.Error())
	}
	sort.Slice(files, func(i, j int) bool { return files[i].RelativePath < files[j].RelativePath })

	for _, file := range files {
		run.FilesSeen++
		if err := h.processFile(ctx, cmd.Scope, actor, &run, file); err != nil {
			return h.finish(ctx, run, false, err.Error())
		}
	}
	return h.finish(ctx, run, true, "")
}

func (h *IngestADRsHandler) processFile(ctx context.Context, scope domain.Scope, actor string, run *domain.ADRIngestRun, file ports.ADRFile) error {
	now := h.clock.Now()
	return h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		existing, ok, err := uow.ADRIngest().GetProcessedADR(ctx, scope, file.RelativePath, file.Checksum)
		if err != nil {
			return err
		}
		if ok {
			_ = existing
			h.advanceCursor(ctx, uow, scope, file)
			return nil
		}

		extracted := appingest.ExtractADR(file.RelativePath, file.Content)
		if !extracted.OK {
			run.FilesSkipped++
			if err := uow.ADRIngest().InsertProcessedADR(ctx, domain.ProcessedADR{
				Scope:        scope,
				RelativePath: file.RelativePath,
				Checksum:     file.Checksum,
				ADRID:        domain.ADRIdentityFromPath(file.RelativePath),
				Skipped:      true,
				SkipReason:   extracted.SkipReason,
				ProcessedAt:  now,
			}); err != nil {
				if conflictAlreadyProcessed(err) {
					h.advanceCursor(ctx, uow, scope, file)
					return nil
				}
				return err
			}
			h.advanceCursor(ctx, uow, scope, file)
			return nil
		}

		snap := extracted.Snapshot
		snap.Checksum = file.Checksum
		evidence, err := evidenceForADR(snap)
		if err != nil {
			return err
		}
		entry, err := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
			Statement: extracted.Statement,
			Scope:     scope,
			CreatedBy: actor,
			Evidence:  evidence,
			Now:       now,
		})
		if err != nil {
			return err
		}

		preds, err := h.predecessorsToSupersede(ctx, uow, scope, snap, entry.ID)
		if err != nil {
			return err
		}

		createAudit, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
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
		if err := uow.Audits().Add(ctx, createAudit); err != nil {
			return err
		}
		if err := uow.Outbox().Add(ctx, outboxEvent); err != nil {
			return err
		}

		for i := range preds {
			result, err := domain.ApplySupersessionWithSuccessor(preds[i], entry, actor, now)
			if err != nil {
				return err
			}
			if err := uow.LoreEntries().Save(ctx, result.Predecessor); err != nil {
				return err
			}
			if err := uow.Audits().Add(ctx, result.SupersedeAudit); err != nil {
				return err
			}
			run.LoreSuperseded++
		}

		if snap.StatusClass == domain.ADRStatusHistorical {
			updated, invAudit, err := domain.ApplyInvalidation(entry, actor, now)
			if err != nil {
				return err
			}
			if err := uow.LoreEntries().Save(ctx, updated); err != nil {
				return err
			}
			if invAudit != nil {
				if err := uow.Audits().Add(ctx, *invAudit); err != nil {
					return err
				}
			}
		}

		if err := uow.ADRIngest().InsertProcessedADR(ctx, domain.ProcessedADR{
			Scope:        scope,
			RelativePath: file.RelativePath,
			Checksum:     file.Checksum,
			ADRID:        snap.ADRID,
			LoreEntryID:  entry.ID,
			Skipped:      false,
			ProcessedAt:  now,
		}); err != nil {
			if conflictAlreadyProcessed(err) {
				h.advanceCursor(ctx, uow, scope, file)
				return nil
			}
			return err
		}
		run.LoreStored++
		h.advanceCursor(ctx, uow, scope, file)
		return nil
	})
}

func (h *IngestADRsHandler) predecessorsToSupersede(ctx context.Context, uow ports.UnitOfWork, scope domain.Scope, snap domain.ADRSnapshot, successorID string) ([]domain.LoreEntry, error) {
	seen := map[string]bool{}
	out := make([]domain.LoreEntry, 0)

	add := func(id string) error {
		if id == "" || seen[id] || id == successorID {
			return nil
		}
		entry, err := uow.LoreEntries().Get(ctx, id)
		if err != nil {
			var nf *domain.NotFoundError
			if errors.As(err, &nf) {
				return nil
			}
			return err
		}
		if entry.Origin != domain.KnowledgeOriginArchitectureDecision {
			return nil
		}
		if !domain.IsCurrent(entry) {
			return nil
		}
		seen[id] = true
		out = append(out, entry)
		return nil
	}

	prev, ok, err := uow.ADRIngest().LatestStoredByPath(ctx, scope, snap.RelativePath)
	if err != nil {
		return nil, err
	}
	if ok {
		if err := add(prev.LoreEntryID); err != nil {
			return nil, err
		}
	}
	for _, adrID := range snap.Supersedes {
		if adrID == "" || adrID == snap.ADRID {
			continue
		}
		row, ok, err := uow.ADRIngest().LatestStoredByADRID(ctx, scope, adrID)
		if err != nil {
			return nil, err
		}
		if ok {
			if err := add(row.LoreEntryID); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (h *IngestADRsHandler) advanceCursor(ctx context.Context, uow ports.UnitOfWork, scope domain.Scope, file ports.ADRFile) {
	now := h.clock.Now()
	_ = uow.ADRIngest().UpsertCursor(ctx, domain.ADRIngestCursor{
		Scope:        scope,
		LastPath:     file.RelativePath,
		LastChecksum: file.Checksum,
		UpdatedAt:    now,
	})
}

func evidenceForADR(snap domain.ADRSnapshot) ([]domain.EvidenceReference, error) {
	adrRef, err := domain.NewEvidenceReference(domain.EvidenceTypeADR, snap.ADRID)
	if err != nil {
		return nil, err
	}
	out := []domain.EvidenceReference{adrRef}
	if snap.RelativePath != "" {
		if ref, err := domain.NewEvidenceReference(domain.EvidenceTypePath, snap.RelativePath); err == nil {
			out = append(out, ref)
		}
	}
	return out, nil
}

func (h *IngestADRsHandler) finish(ctx context.Context, run domain.ADRIngestRun, ok bool, errMsg string) (domain.ADRIngestRun, error) {
	now := h.clock.Now()
	if ok {
		run = run.MarkSucceeded(now)
	} else {
		run = run.MarkFailed(now, errMsg)
	}
	_ = h.withUoW(ctx, func(uow ports.UnitOfWork) error {
		return uow.ADRIngest().UpdateRun(ctx, run)
	})
	return run, nil
}

func (h *IngestADRsHandler) withUoW(ctx context.Context, fn func(ports.UnitOfWork) error) error {
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
