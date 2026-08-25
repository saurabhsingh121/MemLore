package presenters

import (
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// Scope is the JSON scope shape shared by REST and MCP.
type Scope struct {
	Kind domain.ScopeKind `json:"kind"`
	Key  string           `json:"key"`
}

// Evidence is a single evidence reference in API responses.
type Evidence struct {
	Type domain.EvidenceType `json:"type"`
	Value string              `json:"value"`
}

// LoreEntry is the JSON lore entry shape shared by REST and MCP.
type LoreEntry struct {
	ID                 string     `json:"id"`
	Statement          string     `json:"statement"`
	Scope              Scope      `json:"scope"`
	Origin             string     `json:"origin"`
	VerificationStatus string     `json:"verification_status"`
	Evidence           []Evidence `json:"evidence"`
	CreatedBy          string     `json:"created_by"`
	CreatedAt          time.Time  `json:"created_at"`
	VerifiedBy         *string    `json:"verified_by"`
	VerifiedAt         *time.Time `json:"verified_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// LoreEntryList wraps lore entries for list/search responses.
type LoreEntryList struct {
	Items []LoreEntry `json:"items"`
}

// AuditRecord is the JSON audit record shape shared by REST and MCP.
type AuditRecord struct {
	ID        string    `json:"id"`
	TargetID  string    `json:"target_id"`
	Action    string    `json:"action"`
	ActorID   string    `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditList wraps audit records for REST list responses.
type AuditList struct {
	Items []AuditRecord `json:"items"`
}

// ExplainResult is lore entry fields plus chronological audits.
type ExplainResult struct {
	LoreEntry
	Audits []AuditRecord `json:"audits"`
}

// ToLoreEntry maps a domain lore entry to its API representation.
func ToLoreEntry(entry domain.LoreEntry) LoreEntry {
	evidence := make([]Evidence, len(entry.Evidence))
	for i, item := range entry.Evidence {
		evidence[i] = Evidence{Type: item.Type, Value: item.Value}
	}
	return LoreEntry{
		ID:        entry.ID,
		Statement: entry.Statement,
		Scope: Scope{
			Kind: entry.Scope.Kind,
			Key:  entry.Scope.Key,
		},
		Origin:             string(entry.Origin),
		VerificationStatus: string(entry.VerificationStatus),
		Evidence:           evidence,
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          entry.CreatedAt.UTC(),
		VerifiedBy:         entry.VerifiedBy,
		VerifiedAt:         entry.VerifiedAt,
		UpdatedAt:          entry.UpdatedAt.UTC(),
	}
}

// ToAuditRecord maps a domain audit record to its API representation.
func ToAuditRecord(record domain.AuditRecord) AuditRecord {
	return AuditRecord{
		ID:        record.ID,
		TargetID:  record.TargetID,
		Action:    string(record.Action),
		ActorID:   record.ActorID,
		CreatedAt: record.CreatedAt.UTC(),
	}
}
