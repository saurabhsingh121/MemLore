package domain

import "errors"

// Sentinel errors for domain validation and lookup failures.
var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

// ValidationError indicates input failed domain rules.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

// NotFoundError indicates a requested entity does not exist.
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	if e.Message == "" {
		return "not found"
	}
	return e.Message
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

func validationError(message string) error {
	return &ValidationError{Message: message}
}
