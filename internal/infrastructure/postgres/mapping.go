package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

type evidenceJSON struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func loreEntryToInsertParams(entry domain.LoreEntry) (sqlc.InsertLoreEntryParams, error) {
	evidence, err := marshalEvidence(entry.Evidence)
	if err != nil {
		return sqlc.InsertLoreEntryParams{}, err
	}
	return sqlc.InsertLoreEntryParams{
		ID:                 entry.ID,
		Statement:          entry.Statement,
		ScopeKind:          string(entry.Scope.Kind),
		ScopeKey:           entry.Scope.Key,
		Origin:             string(entry.Origin),
		VerificationStatus: string(entry.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          timestamptzFromTime(entry.CreatedAt),
		VerifiedBy:         textFromPtr(entry.VerifiedBy),
		VerifiedAt:         timestamptzFromPtr(entry.VerifiedAt),
		UpdatedAt:          timestamptzFromTime(entry.UpdatedAt),
	}, nil
}

func loreEntryToUpdateParams(entry domain.LoreEntry) (sqlc.UpdateLoreEntryParams, error) {
	evidence, err := marshalEvidence(entry.Evidence)
	if err != nil {
		return sqlc.UpdateLoreEntryParams{}, err
	}
	return sqlc.UpdateLoreEntryParams{
		ID:                 entry.ID,
		Statement:          entry.Statement,
		ScopeKind:          string(entry.Scope.Kind),
		ScopeKey:           entry.Scope.Key,
		Origin:             string(entry.Origin),
		VerificationStatus: string(entry.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          timestamptzFromTime(entry.CreatedAt),
		VerifiedBy:         textFromPtr(entry.VerifiedBy),
		VerifiedAt:         timestamptzFromPtr(entry.VerifiedAt),
		UpdatedAt:          timestamptzFromTime(entry.UpdatedAt),
	}, nil
}

func loreEntryFromRow(row sqlc.LoreEntry) (domain.LoreEntry, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	scope := domain.Scope{Kind: kind, Key: row.ScopeKey}
	evidence, err := unmarshalEvidence(row.Evidence)
	if err != nil {
		return domain.LoreEntry{}, err
	}
	return domain.LoreEntry{
		ID:                 row.ID,
		Statement:          row.Statement,
		Scope:              scope,
		Origin:             domain.KnowledgeOrigin(row.Origin),
		VerificationStatus: domain.VerificationStatus(row.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          row.CreatedBy,
		CreatedAt:          timeFromTimestamptz(row.CreatedAt),
		VerifiedBy:         ptrFromText(row.VerifiedBy),
		VerifiedAt:         ptrFromTimestamptz(row.VerifiedAt),
		UpdatedAt:          timeFromTimestamptz(row.UpdatedAt),
	}, nil
}

func auditRecordToInsertParams(record domain.AuditRecord) sqlc.InsertAuditRecordParams {
	return sqlc.InsertAuditRecordParams{
		ID:        record.ID,
		TargetID:  record.TargetID,
		Action:    string(record.Action),
		ActorID:   record.ActorID,
		CreatedAt: timestamptzFromTime(record.CreatedAt),
	}
}

func auditRecordFromRow(row sqlc.AuditRecord) domain.AuditRecord {
	return domain.AuditRecord{
		ID:        row.ID,
		TargetID:  row.TargetID,
		Action:    domain.AuditAction(row.Action),
		ActorID:   row.ActorID,
		CreatedAt: timeFromTimestamptz(row.CreatedAt),
	}
}

func marshalEvidence(items []domain.EvidenceReference) ([]byte, error) {
	payload := make([]evidenceJSON, len(items))
	for i, item := range items {
		payload[i] = evidenceJSON{Type: string(item.Type), Value: item.Value}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence: %w", err)
	}
	return raw, nil
}

func unmarshalEvidence(raw []byte) ([]domain.EvidenceReference, error) {
	if len(raw) == 0 {
		return []domain.EvidenceReference{}, nil
	}
	var payload []evidenceJSON
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal evidence: %w", err)
	}
	out := make([]domain.EvidenceReference, len(payload))
	for i, item := range payload {
		evidenceType, err := domain.ParseEvidenceType(item.Type)
		if err != nil {
			return nil, err
		}
		out[i] = domain.EvidenceReference{Type: evidenceType, Value: item.Value}
	}
	return out, nil
}

func timestamptzFromTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timestamptzFromPtr(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timeFromTimestamptz(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time.UTC()
}

func ptrFromTimestamptz(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func textFromPtr(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func ptrFromText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}
