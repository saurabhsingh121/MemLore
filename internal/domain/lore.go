package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LoreEntry is governance-plane engineering knowledge with provenance.
type LoreEntry struct {
	ID                 string
	Statement          string
	Scope              Scope
	Origin             KnowledgeOrigin
	VerificationStatus VerificationStatus
	Evidence           []EvidenceReference
	CreatedBy          string
	CreatedAt          time.Time
	VerifiedBy         *string
	VerifiedAt         *time.Time
	InvalidatedBy      *string
	InvalidatedAt      *time.Time
	SupersededByID     *string
	UpdatedAt          time.Time
}

// NewLoreEntryInput is input for creating a validated lore entry.
type NewLoreEntryInput struct {
	Statement string
	Scope     Scope
	CreatedBy string
	Origin    KnowledgeOrigin
	Evidence  []EvidenceReference
	ID        string
	Now       time.Time
}

// NewLoreEntry creates a human-authored lore entry with validation.
// Characterization: src/memlore/domain/models/lore_entry.py
func NewLoreEntry(in NewLoreEntryInput) (LoreEntry, error) {
	statement := strings.TrimSpace(in.Statement)
	createdBy := strings.TrimSpace(in.CreatedBy)

	if statement == "" {
		return LoreEntry{}, validationError("statement must be non-empty")
	}
	if len(statement) > MaxStatementLength {
		return LoreEntry{}, validationError(
			fmt.Sprintf("statement must be at most %d characters", MaxStatementLength),
		)
	}
	if createdBy == "" {
		return LoreEntry{}, validationError("created_by must be non-empty")
	}
	origin := in.Origin
	if origin == "" {
		origin = KnowledgeOriginHumanAuthored
	}
	if origin != KnowledgeOriginHumanAuthored {
		return LoreEntry{}, validationError("create origin must be human_authored")
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	evidence := in.Evidence
	if evidence == nil {
		evidence = []EvidenceReference{}
	}

	return LoreEntry{
		ID:                 id,
		Statement:          statement,
		Scope:              in.Scope,
		Origin:             origin,
		VerificationStatus: VerificationUnverified,
		Evidence:           evidence,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// NewObservationalLoreEntry creates unverified repository_observation lore.
// Git and PR ingest MUST use this path; NewLoreEntry remains human-authored only.
func NewObservationalLoreEntry(in NewLoreEntryInput) (LoreEntry, error) {
	statement := strings.TrimSpace(in.Statement)
	createdBy := strings.TrimSpace(in.CreatedBy)

	if statement == "" {
		return LoreEntry{}, validationError("statement must be non-empty")
	}
	if len(statement) > MaxStatementLength {
		return LoreEntry{}, validationError(
			fmt.Sprintf("statement must be at most %d characters", MaxStatementLength),
		)
	}
	if createdBy == "" {
		return LoreEntry{}, validationError("created_by must be non-empty")
	}
	origin := in.Origin
	if origin == "" {
		origin = KnowledgeOriginRepositoryObservation
	}
	if origin != KnowledgeOriginRepositoryObservation {
		return LoreEntry{}, validationError("observational origin must be repository_observation")
	}

	evidence := in.Evidence
	if evidence == nil {
		evidence = []EvidenceReference{}
	}
	hasObservationalEvidence := false
	for _, ref := range evidence {
		if (ref.Type == EvidenceTypeCommit || ref.Type == EvidenceTypePR) && strings.TrimSpace(ref.Value) != "" {
			hasObservationalEvidence = true
			break
		}
	}
	if !hasObservationalEvidence {
		return LoreEntry{}, validationError("observational lore requires commit or pr evidence")
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	return LoreEntry{
		ID:                 id,
		Statement:          statement,
		Scope:              in.Scope,
		Origin:             origin,
		VerificationStatus: VerificationUnverified,
		Evidence:           evidence,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// NewArchitectureDecisionLoreEntry creates verified architecture_decision lore.
// ADR ingest MUST use this path; NewLoreEntry remains human-authored only.
func NewArchitectureDecisionLoreEntry(in NewLoreEntryInput) (LoreEntry, error) {
	statement := strings.TrimSpace(in.Statement)
	createdBy := strings.TrimSpace(in.CreatedBy)

	if statement == "" {
		return LoreEntry{}, validationError("statement must be non-empty")
	}
	if len(statement) > MaxStatementLength {
		return LoreEntry{}, validationError(
			fmt.Sprintf("statement must be at most %d characters", MaxStatementLength),
		)
	}
	if createdBy == "" {
		return LoreEntry{}, validationError("created_by must be non-empty")
	}
	origin := in.Origin
	if origin == "" {
		origin = KnowledgeOriginArchitectureDecision
	}
	if origin != KnowledgeOriginArchitectureDecision {
		return LoreEntry{}, validationError("architecture decision origin must be architecture_decision")
	}

	evidence := in.Evidence
	if evidence == nil {
		evidence = []EvidenceReference{}
	}
	hasADR := false
	for _, ref := range evidence {
		if ref.Type == EvidenceTypeADR && strings.TrimSpace(ref.Value) != "" {
			hasADR = true
			break
		}
	}
	if !hasADR {
		return LoreEntry{}, validationError("architecture decision lore requires adr evidence")
	}

	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	return LoreEntry{
		ID:                 id,
		Statement:          statement,
		Scope:              in.Scope,
		Origin:             origin,
		VerificationStatus: VerificationVerified,
		Evidence:           evidence,
		CreatedBy:          createdBy,
		CreatedAt:          now,
		UpdatedAt:          now,
		VerifiedBy:         &createdBy,
		VerifiedAt:         &now,
	}, nil
}

// IsSuperseded reports whether the entry has been replaced by a successor.
func IsSuperseded(entry LoreEntry) bool {
	return entry.SupersededByID != nil && strings.TrimSpace(*entry.SupersededByID) != ""
}

// IsCurrent reports whether the entry is eligible for default retrieval.
// Current means not superseded and not invalidated.
func IsCurrent(entry LoreEntry) bool {
	return !IsSuperseded(entry) && entry.VerificationStatus != VerificationInvalidated
}
