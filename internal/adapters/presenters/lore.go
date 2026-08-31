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
	Type  domain.EvidenceType `json:"type"`
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
	InvalidatedBy      *string    `json:"invalidated_by"`
	InvalidatedAt      *time.Time `json:"invalidated_at"`
	SupersededByID     *string    `json:"superseded_by_id"`
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

// ExplainResult is lore entry fields plus chronological audits and authority evaluation.
type ExplainResult struct {
	LoreEntry
	Audits           []AuditRecord    `json:"audits"`
	TrustBand        string           `json:"trust_band"`
	AuthorityScore   float64          `json:"authority_score"`
	AuthorityFactors AuthorityFactors `json:"authority_factors"`
	FactorBreakdown  []string         `json:"factor_breakdown"`
}

// ToExplainResult maps an entry, audits, and evaluation to the API payload.
func ToExplainResult(entry domain.LoreEntry, audits []domain.AuditRecord, eval domain.Evaluation) ExplainResult {
	breakdown := eval.Breakdown
	if breakdown == nil {
		breakdown = []string{}
	}
	mapped := make([]AuditRecord, 0, len(audits))
	for _, record := range audits {
		mapped = append(mapped, ToAuditRecord(record))
	}
	return ExplainResult{
		LoreEntry:      ToLoreEntry(entry),
		Audits:         mapped,
		TrustBand:      string(eval.Band),
		AuthorityScore: eval.Score,
		AuthorityFactors: AuthorityFactors{
			VerificationStatus: eval.Factors.VerificationStatus,
			Origin:             eval.Factors.Origin,
			SupersessionStatus: eval.Factors.SupersessionStatus,
			RecencyBoost:       eval.Factors.RecencyBoost,
			EvidenceStrength:   eval.Factors.EvidenceStrength,
			SourceType:         eval.Factors.SourceType,
			ScopeMatch:         eval.Factors.ScopeMatch,
			GraphScore:         eval.Factors.GraphScore,
		},
		FactorBreakdown: breakdown,
	}
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
		InvalidatedBy:      entry.InvalidatedBy,
		InvalidatedAt:      entry.InvalidatedAt,
		SupersededByID:     entry.SupersededByID,
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
