package domain

import (
	"strings"
	"time"
)

// ApplyVerification marks a lore entry verified on first call only.
// Returns the entry, an optional verify audit, and an error.
// Characterization: src/memlore/domain/services/verification.py
func ApplyVerification(entry LoreEntry, actorID string, now time.Time) (LoreEntry, *AuditRecord, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return LoreEntry{}, nil, validationError("actor must be non-empty")
	}

	if entry.VerificationStatus == VerificationInvalidated {
		return LoreEntry{}, nil, validationError("cannot verify an invalidated lore entry")
	}
	if IsSuperseded(entry) {
		return LoreEntry{}, nil, validationError("cannot verify a superseded lore entry")
	}

	if entry.VerificationStatus == VerificationVerified {
		return entry, nil, nil
	}

	now = EnsureUTC(now)
	entry.VerificationStatus = VerificationVerified
	entry.VerifiedBy = &actor
	entry.VerifiedAt = &now
	entry.UpdatedAt = now

	audit, err := NewAuditRecord(NewAuditRecordInput{
		TargetID:  entry.ID,
		Action:    AuditActionVerify,
		ActorID:   actor,
		CreatedAt: now,
	})
	if err != nil {
		return LoreEntry{}, nil, err
	}

	return entry, &audit, nil
}

// EnsureUTC normalizes a timestamp to UTC.
func EnsureUTC(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	if value.Location() == time.UTC {
		return value
	}
	return value.UTC()
}
