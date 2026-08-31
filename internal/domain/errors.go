package domain

import "errors"

// Sentinel errors for domain validation and lookup failures.
var (
	ErrValidation   = errors.New("validation error")
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
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

// UnauthorizedError indicates missing or invalid authentication.
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	if e.Message == "" {
		return "unauthorized"
	}
	return e.Message
}

func (e *UnauthorizedError) Is(target error) bool {
	return target == ErrUnauthorized
}

// ForbiddenError indicates authenticated caller lacks permission.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	if e.Message == "" {
		return "forbidden"
	}
	return e.Message
}

func (e *ForbiddenError) Is(target error) bool {
	return target == ErrForbidden
}

func validationError(message string) error {
	return &ValidationError{Message: message}
}
