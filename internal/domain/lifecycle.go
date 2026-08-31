package domain

import (
	"strings"
	"time"
)

// ApplyInvalidation marks a current lore entry invalidated on first call only.
// Re-invalidating an already-invalidated entry is a no-op (no audit).
// Superseded entries cannot be invalidated.
func ApplyInvalidation(entry LoreEntry, actorID string, now time.Time) (LoreEntry, *AuditRecord, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return LoreEntry{}, nil, validationError("actor must be non-empty")
	}
	if IsSuperseded(entry) {
		return LoreEntry{}, nil, validationError("cannot invalidate a superseded lore entry")
	}
	if entry.VerificationStatus == VerificationInvalidated {
		return entry, nil, nil
	}

	now = EnsureUTC(now)
	entry.VerificationStatus = VerificationInvalidated
	entry.InvalidatedBy = &actor
	entry.InvalidatedAt = &now
	entry.UpdatedAt = now

	audit, err := NewAuditRecord(NewAuditRecordInput{
		TargetID:  entry.ID,
		Action:    AuditActionInvalidate,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return LoreEntry{}, nil, err
	}
	return entry, &audit, nil
}

// SupersessionResult is the predecessor, successor, and both audits.
type SupersessionResult struct {
	Predecessor    LoreEntry
	Successor      LoreEntry
	SupersedeAudit AuditRecord
	CreateAudit    AuditRecord
}

// ApplySupersession replaces a current lore entry with a successor in the same scope.
// Already-superseded or invalidated predecessors are rejected.
func ApplySupersession(
	predecessor LoreEntry,
	statement string,
	actorID string,
	evidence []EvidenceReference,
	now time.Time,
) (SupersessionResult, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return SupersessionResult{}, validationError("actor must be non-empty")
	}
	if IsSuperseded(predecessor) {
		return SupersessionResult{}, validationError("cannot supersede a superseded lore entry")
	}
	if predecessor.VerificationStatus == VerificationInvalidated {
		return SupersessionResult{}, validationError("cannot supersede an invalidated lore entry")
	}

	now = EnsureUTC(now)
	successor, err := NewLoreEntry(NewLoreEntryInput{
		Statement: statement,
		Scope:     predecessor.Scope,
		CreatedBy: actor,
		Evidence:  evidence,
		Now:       now,
	})
	if err != nil {
		return SupersessionResult{}, err
	}

	successorID := successor.ID
	predecessor.SupersededByID = &successorID
	predecessor.UpdatedAt = now

	supersedeAudit, err := NewAuditRecord(NewAuditRecordInput{
		TargetID:  predecessor.ID,
		Action:    AuditActionSupersede,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return SupersessionResult{}, err
	}
	createAudit, err := NewAuditRecord(NewAuditRecordInput{
		TargetID:  successor.ID,
		Action:    AuditActionCreate,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return SupersessionResult{}, err
	}

	return SupersessionResult{
		Predecessor:    predecessor,
		Successor:      successor,
		SupersedeAudit: supersedeAudit,
		CreateAudit:    createAudit,
	}, nil
}
