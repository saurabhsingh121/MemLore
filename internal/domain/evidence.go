package domain

import (
	"fmt"
	"strings"
)

// EvidenceReference links lore to supporting material.
type EvidenceReference struct {
	Type  EvidenceType
	Value string
}

// NewEvidenceReference validates and returns an evidence reference.
func NewEvidenceReference(evidenceType EvidenceType, value string) (EvidenceReference, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return EvidenceReference{}, validationError("evidence value must be non-empty")
	}
	if len(trimmed) > MaxEvidenceValueLength {
		return EvidenceReference{}, validationError(
			fmt.Sprintf("evidence value must be at most %d characters", MaxEvidenceValueLength),
		)
	}
	return EvidenceReference{Type: evidenceType, Value: trimmed}, nil
}
