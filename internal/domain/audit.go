package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuditRecord is an append-only governance audit event.
type AuditRecord struct {
	ID        string
	TargetID  string
	Action    AuditAction
	ActorID   string
	CreatedAt time.Time
}

// NewAuditRecordInput is input for creating a validated audit record.
type NewAuditRecordInput struct {
	TargetID  string
	Action    AuditAction
	ActorID   string
	ID        string
	CreatedAt time.Time
}

// NewAuditRecord validates and returns an audit record.
// Characterization: src/memlore/domain/models/audit_record.py
func NewAuditRecord(in NewAuditRecordInput) (AuditRecord, error) {
	actor := strings.TrimSpace(in.ActorID)
	target := strings.TrimSpace(in.TargetID)
	if actor == "" {
		return AuditRecord{}, validationError("actor_id must be non-empty")
	}
	if target == "" {
		return AuditRecord{}, validationError("target_id must be non-empty")
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}

	createdAt := in.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	} else {
		createdAt = createdAt.UTC()
	}

	return AuditRecord{
		ID:        id,
		TargetID:  target,
		Action:    in.Action,
		ActorID:   actor,
		CreatedAt: createdAt,
	}, nil
}
