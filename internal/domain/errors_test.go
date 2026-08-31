package domain_test

import (
	"errors"
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

func TestValidationErrorIsValidationSentinel(t *testing.T) {
	err := &domain.ValidationError{Message: "scope key must be non-empty"}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatal("expected ValidationError to match ErrValidation")
	}
}

func TestConflictErrorIsConflictSentinel(t *testing.T) {
	err := &domain.ConflictError{Message: "ingest already running"}
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatal("expected ConflictError to match ErrConflict")
	}
}
