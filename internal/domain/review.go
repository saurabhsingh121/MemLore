package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReviewStatus is the durable outcome of a suggested-lore review.
type ReviewStatus string

const (
	ReviewStatusAccepted ReviewStatus = "accepted"
	ReviewStatusRejected ReviewStatus = "rejected"
)

// ExtractIdentity uniquely identifies an observational extract for review.
type ExtractIdentity struct {
	Scope             Scope
	EvidenceType      EvidenceType
	EvidenceValue     string
	StatementChecksum string
}

// Key is a stable lookup key for the extract identity.
func (id ExtractIdentity) Key() string {
	return strings.Join([]string{
		string(id.Scope.Kind),
		id.Scope.Key,
		string(id.EvidenceType),
		id.EvidenceValue,
		id.StatementChecksum,
	}, "\x1f")
}

// ReviewDecision is a durable Accept/Reject overlay for one extract identity.
type ReviewDecision struct {
	ID                string
	Scope             Scope
	EvidenceType      EvidenceType
	EvidenceValue     string
	StatementChecksum string
	LoreEntryID       string
	SuccessorLoreID   *string
	Status            ReviewStatus
	ActorID           string
	DecidedAt         time.Time
}

// AcceptResult is the supersession plus the accepted review decision.
type AcceptResult struct {
	SupersessionResult
	Decision ReviewDecision
}

// StatementChecksum returns SHA-256 hex of the trimmed statement.
func StatementChecksum(statement string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(statement)))
	return hex.EncodeToString(sum[:])
}

// PrimaryObservationalEvidence returns the first commit evidence, else the first pr evidence.
func PrimaryObservationalEvidence(entry LoreEntry) (EvidenceReference, error) {
	var firstPR *EvidenceReference
	for i := range entry.Evidence {
		ref := entry.Evidence[i]
		if strings.TrimSpace(ref.Value) == "" {
			continue
		}
		if ref.Type == EvidenceTypeCommit {
			return ref, nil
		}
		if ref.Type == EvidenceTypePR && firstPR == nil {
			copyRef := ref
			firstPR = &copyRef
		}
	}
	if firstPR != nil {
		return *firstPR, nil
	}
	return EvidenceReference{}, validationError("observational extract requires commit or pr evidence")
}

// ExtractIdentityOf derives the review identity of an observational lore entry.
func ExtractIdentityOf(entry LoreEntry) (ExtractIdentity, error) {
	ev, err := PrimaryObservationalEvidence(entry)
	if err != nil {
		return ExtractIdentity{}, err
	}
	return ExtractIdentity{
		Scope:             entry.Scope,
		EvidenceType:      ev.Type,
		EvidenceValue:     strings.TrimSpace(ev.Value),
		StatementChecksum: StatementChecksum(entry.Statement),
	}, nil
}

// IsReviewEligible reports whether an entry may appear in the pending review queue.
func IsReviewEligible(entry LoreEntry) bool {
	if !IsCurrent(entry) {
		return false
	}
	if entry.Origin != KnowledgeOriginRepositoryObservation {
		return false
	}
	_, err := PrimaryObservationalEvidence(entry)
	return err == nil
}

// ReviewSourceType is "commit" or "pr" from primary observational evidence.
func ReviewSourceType(entry LoreEntry) string {
	ev, err := PrimaryObservationalEvidence(entry)
	if err != nil {
		return ""
	}
	return string(ev.Type)
}

// AcceptSuggestedLore promotes an observational extract by superseding it.
// Empty or whitespace-equal statement is Accept-as-stated (human_verified).
// A different statement is Edit-then-Accept (verified human_authored).
func AcceptSuggestedLore(predecessor LoreEntry, statement, actorID string, now time.Time) (AcceptResult, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return AcceptResult{}, validationError("actor must be non-empty")
	}
	if !IsReviewEligible(predecessor) {
		return AcceptResult{}, validationError("lore entry is not a pending suggested-lore item")
	}

	trimmed := strings.TrimSpace(statement)
	asStated := trimmed == "" || trimmed == strings.TrimSpace(predecessor.Statement)
	successorStatement := strings.TrimSpace(predecessor.Statement)
	if !asStated {
		successorStatement = trimmed
	}

	evidence := append([]EvidenceReference(nil), predecessor.Evidence...)
	in := NewLoreEntryInput{
		Statement: successorStatement,
		Scope:     predecessor.Scope,
		CreatedBy: actor,
		Evidence:  evidence,
		Now:       now,
	}

	var successor LoreEntry
	var err error
	if asStated {
		successor, err = NewHumanVerifiedLoreEntry(in)
	} else {
		successor, err = NewVerifiedHumanAuthoredLoreEntry(in)
	}
	if err != nil {
		return AcceptResult{}, err
	}

	super, err := ApplySupersessionWithSuccessor(predecessor, successor, actor, now)
	if err != nil {
		return AcceptResult{}, err
	}

	identity, err := ExtractIdentityOf(predecessor)
	if err != nil {
		return AcceptResult{}, err
	}
	succID := super.Successor.ID
	decision := ReviewDecision{
		ID:                uuid.NewString(),
		Scope:             identity.Scope,
		EvidenceType:      identity.EvidenceType,
		EvidenceValue:     identity.EvidenceValue,
		StatementChecksum: identity.StatementChecksum,
		LoreEntryID:       predecessor.ID,
		SuccessorLoreID:   &succID,
		Status:            ReviewStatusAccepted,
		ActorID:           actor,
		DecidedAt:         EnsureUTC(now),
	}
	return AcceptResult{SupersessionResult: super, Decision: decision}, nil
}

// RejectSuggestedLore records a negative decision without mutating the observation.
func RejectSuggestedLore(entry LoreEntry, actorID string, now time.Time) (ReviewDecision, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return ReviewDecision{}, validationError("actor must be non-empty")
	}
	if !IsReviewEligible(entry) {
		return ReviewDecision{}, validationError("lore entry is not a pending suggested-lore item")
	}
	identity, err := ExtractIdentityOf(entry)
	if err != nil {
		return ReviewDecision{}, err
	}
	return ReviewDecision{
		ID:                uuid.NewString(),
		Scope:             identity.Scope,
		EvidenceType:      identity.EvidenceType,
		EvidenceValue:     identity.EvidenceValue,
		StatementChecksum: identity.StatementChecksum,
		LoreEntryID:       entry.ID,
		Status:            ReviewStatusRejected,
		ActorID:           actor,
		DecidedAt:         EnsureUTC(now),
	}, nil
}
