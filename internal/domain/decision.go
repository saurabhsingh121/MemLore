package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// DecisionSourceKind distinguishes human-recorded Decisions from ADR projections.
type DecisionSourceKind string

const (
	DecisionSourceHuman DecisionSourceKind = "human"
	DecisionSourceADR   DecisionSourceKind = "adr"
)

// DecisionAlternative is a considered option on a Decision (F042-as-fields).
type DecisionAlternative struct {
	Label string
	Note  string
}

// Decision is a first-class engineering decision linked to one lore entry.
type Decision struct {
	ID                 string
	Scope              Scope
	Question           string
	Choice             string
	Rationale          string
	Alternatives       []DecisionAlternative
	Consequences       string
	Owner              string
	DecidedAt          time.Time
	AffectedComponents []string
	Evidence           []EvidenceReference
	SourceKind         DecisionSourceKind
	SupersededByID     *string
	Current            bool
	CreatedBy          string
	CreatedAt          time.Time
}

// NewDecisionInput is input for a human-recorded Decision.
type NewDecisionInput struct {
	ID                 string
	Scope              Scope
	Question           string
	Choice             string
	Rationale          string
	Alternatives       []DecisionAlternative
	Consequences       string
	Owner              string
	DecidedAt          time.Time
	AffectedComponents []string
	Evidence           []EvidenceReference
	CreatedBy          string
	Now                time.Time
}

// NewHumanDecision validates and builds a current human-recorded Decision.
func NewHumanDecision(in NewDecisionInput) (Decision, error) {
	if in.Scope.Kind != ScopeKindRepository {
		return Decision{}, validationError("decision scope kind must be repository")
	}
	question := strings.TrimSpace(in.Question)
	choice := strings.TrimSpace(in.Choice)
	owner := strings.TrimSpace(in.Owner)
	createdBy := strings.TrimSpace(in.CreatedBy)
	if question == "" {
		return Decision{}, validationError("question must be non-empty")
	}
	if choice == "" {
		return Decision{}, validationError("decision must be non-empty")
	}
	if len(choice) > MaxStatementLength {
		return Decision{}, validationError("decision must be at most 8000 characters")
	}
	if len(question) > MaxStatementLength {
		return Decision{}, validationError("question must be at most 8000 characters")
	}
	if owner == "" {
		return Decision{}, validationError("owner must be non-empty")
	}
	if createdBy == "" {
		return Decision{}, validationError("created_by must be non-empty")
	}

	alts := make([]DecisionAlternative, 0, len(in.Alternatives))
	for _, alt := range in.Alternatives {
		label := strings.TrimSpace(alt.Label)
		if label == "" {
			return Decision{}, validationError("alternative label must be non-empty")
		}
		alts = append(alts, DecisionAlternative{Label: label, Note: strings.TrimSpace(alt.Note)})
	}
	components := make([]string, 0, len(in.AffectedComponents))
	for _, name := range in.AffectedComponents {
		name = strings.TrimSpace(name)
		if name == "" {
			return Decision{}, validationError("affected component must be non-empty")
		}
		components = append(components, name)
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
	decided := in.DecidedAt
	if decided.IsZero() {
		decided = now
	} else {
		decided = decided.UTC()
	}
	evidence := in.Evidence
	if evidence == nil {
		evidence = []EvidenceReference{}
	}

	return Decision{
		ID:                 id,
		Scope:              in.Scope,
		Question:           question,
		Choice:             choice,
		Rationale:          strings.TrimSpace(in.Rationale),
		Alternatives:       alts,
		Consequences:       strings.TrimSpace(in.Consequences),
		Owner:              owner,
		DecidedAt:          decided,
		AffectedComponents: components,
		Evidence:           evidence,
		SourceKind:         DecisionSourceHuman,
		Current:            true,
		CreatedBy:          createdBy,
		CreatedAt:          now,
	}, nil
}

// CanProjectADRDecision reports whether lore is F032 accepted-ADR (or historical ADR) lore.
func CanProjectADRDecision(entry LoreEntry) bool {
	if entry.Origin != KnowledgeOriginArchitectureDecision {
		return false
	}
	for _, ref := range entry.Evidence {
		if ref.Type == EvidenceTypeADR && strings.TrimSpace(ref.Value) != "" {
			return true
		}
	}
	return false
}

// ProjectADRDecision maps architecture_decision lore with adr evidence to a Decision.
func ProjectADRDecision(entry LoreEntry) (Decision, error) {
	if !CanProjectADRDecision(entry) {
		return Decision{}, validationError("lore entry is not an ADR-backed decision")
	}
	evidence := entry.Evidence
	if evidence == nil {
		evidence = []EvidenceReference{}
	}
	return Decision{
		ID:                 entry.ID,
		Scope:              entry.Scope,
		Choice:             entry.Statement,
		Owner:              entry.CreatedBy,
		DecidedAt:          entry.CreatedAt,
		Alternatives:       []DecisionAlternative{},
		AffectedComponents: []string{},
		Evidence:           evidence,
		SourceKind:         DecisionSourceADR,
		SupersededByID:     entry.SupersededByID,
		Current:            IsCurrent(entry),
		CreatedBy:          entry.CreatedBy,
		CreatedAt:          entry.CreatedAt,
	}, nil
}

// WithSupersededBy returns a copy marked superseded by successorID.
func (d Decision) WithSupersededBy(successorID string) Decision {
	id := strings.TrimSpace(successorID)
	d.SupersededByID = &id
	d.Current = false
	return d
}

// DecisionIsCurrent reports whether a persisted human Decision row is still current
// given its superseded pointer and the linked lore lifecycle.
func DecisionIsCurrent(d Decision, lore LoreEntry) bool {
	if d.SupersededByID != nil && strings.TrimSpace(*d.SupersededByID) != "" {
		return false
	}
	return IsCurrent(lore)
}
