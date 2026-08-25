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
