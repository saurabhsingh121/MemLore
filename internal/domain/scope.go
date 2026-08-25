package domain

import (
	"fmt"
	"strings"
)

// Scope is a structured scope identity (kind + key).
type Scope struct {
	Kind ScopeKind
	Key  string
}

// NewScope validates and returns a scope. Keys are trimmed.
func NewScope(kind ScopeKind, key string) (Scope, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return Scope{}, validationError("scope key must be non-empty")
	}
	if len(trimmed) > MaxScopeKeyLength {
		return Scope{}, validationError(
			fmt.Sprintf("scope key must be at most %d characters", MaxScopeKeyLength),
		)
	}
	return Scope{Kind: kind, Key: trimmed}, nil
}
