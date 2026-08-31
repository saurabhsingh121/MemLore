package context_test

import (
	"testing"
	"time"

	appcontext "github.com/memlore/memlore/internal/application/context"
	"github.com/memlore/memlore/internal/domain"
)

func TestFilterCurrentOmitsSupersededAndInvalidated(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	succ := "succ-1"
	entries := []domain.LoreEntry{
		{ID: "c1", Statement: "current", Scope: scope, VerificationStatus: domain.VerificationUnverified, CreatedAt: now, UpdatedAt: now},
		{ID: "s1", Statement: "old", Scope: scope, VerificationStatus: domain.VerificationVerified, SupersededByID: &succ, CreatedAt: now, UpdatedAt: now},
		{ID: "i1", Statement: "bad", Scope: scope, VerificationStatus: domain.VerificationInvalidated, CreatedAt: now, UpdatedAt: now},
	}
	got := appcontext.FilterCurrent(entries)
	if len(got) != 1 || got[0].ID != "c1" {
		t.Fatalf("got = %+v", got)
	}
}

func TestDetectConflictsDifferentStatementsSameScope(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entries := []domain.LoreEntry{
		{ID: "a", Statement: "Use blue-green", Scope: scope, CreatedAt: now, UpdatedAt: now},
		{ID: "b", Statement: "Use rolling deploys", Scope: scope, CreatedAt: now, UpdatedAt: now},
	}
	groups := appcontext.DetectConflicts(entries)
	if len(groups) != 1 {
		t.Fatalf("groups = %+v", groups)
	}
	if len(groups[0].EntryIDs) != 2 {
		t.Fatalf("entry_ids = %v", groups[0].EntryIDs)
	}
	if len(groups[0].Statements) != 2 {
		t.Fatalf("statements = %v", groups[0].Statements)
	}
}

func TestDetectConflictsIdenticalStatementsNotConflict(t *testing.T) {
	now := time.Now().UTC()
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	entries := []domain.LoreEntry{
		{ID: "a", Statement: "Same Rule", Scope: scope, CreatedAt: now, UpdatedAt: now},
		{ID: "b", Statement: "  same rule  ", Scope: scope, CreatedAt: now, UpdatedAt: now},
	}
	groups := appcontext.DetectConflicts(entries)
	if len(groups) != 0 {
		t.Fatalf("expected no conflict, got %+v", groups)
	}
}

func TestDetectConflictsSeparateScopes(t *testing.T) {
	now := time.Now().UTC()
	s1, _ := domain.NewScope(domain.ScopeKindRepository, "r1")
	s2, _ := domain.NewScope(domain.ScopeKindRepository, "r2")
	entries := []domain.LoreEntry{
		{ID: "a", Statement: "A", Scope: s1, CreatedAt: now, UpdatedAt: now},
		{ID: "b", Statement: "B", Scope: s1, CreatedAt: now, UpdatedAt: now},
		{ID: "c", Statement: "C", Scope: s2, CreatedAt: now, UpdatedAt: now},
	}
	groups := appcontext.DetectConflicts(entries)
	if len(groups) != 1 {
		t.Fatalf("groups = %+v", groups)
	}
	if groups[0].Scope.Key != "r1" {
		t.Fatalf("scope = %+v", groups[0].Scope)
	}
}
