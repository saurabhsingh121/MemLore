package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

// Characterization: tests/unit/domain/test_scope_evidence.py

func TestScopeTrimsKeyAndAcceptsRepository(t *testing.T) {
	scope, err := domain.NewScope(domain.ScopeKindRepository, "  github.com/acme/app  ")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	if scope.Key != "github.com/acme/app" {
		t.Fatalf("key = %q", scope.Key)
	}
}

func TestScopeRejectsBlankKey(t *testing.T) {
	_, err := domain.NewScope(domain.ScopeKindTeam, "   ")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "scope key must be non-empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScopeRejectsOversizedKey(t *testing.T) {
	_, err := domain.NewScope(domain.ScopeKindTeam, strings.Repeat("a", domain.MaxScopeKeyLength+1))
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEvidenceReferenceTrimsValue(t *testing.T) {
	ref, err := domain.NewEvidenceReference(domain.EvidenceTypeADR, " 0001-dual-plane ")
	if err != nil {
		t.Fatalf("NewEvidenceReference: %v", err)
	}
	if ref.Value != "0001-dual-plane" {
		t.Fatalf("value = %q", ref.Value)
	}
}

func TestEvidenceRejectsBlankValue(t *testing.T) {
	_, err := domain.NewEvidenceReference(domain.EvidenceTypeURL, "")
	if err == nil {
		t.Fatal("expected validation error")
	}
	var ve *domain.ValidationError
	if !errors.As(err, &ve) || ve.Message != "evidence value must be non-empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}
