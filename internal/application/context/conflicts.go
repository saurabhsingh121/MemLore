package context

import (
	"sort"

	"github.com/memlore/memlore/internal/domain"
)

// ConflictGroup is an ephemeral set of disagreeing current lore in one scope.
type ConflictGroup struct {
	Scope      domain.Scope
	EntryIDs   []string
	Statements []string
}

// DetectConflicts finds scopes with two or more distinct normalized statements
// among the given entries (expected to already be current-only).
func DetectConflicts(entries []domain.LoreEntry) []ConflictGroup {
	type scopeKey struct {
		kind domain.ScopeKind
		key  string
	}
	type bucket struct {
		scope      domain.Scope
		byNorm     map[string]string // normalized -> first original statement
		entryIDs   []string
	}

	grouped := map[scopeKey]*bucket{}
	for _, entry := range entries {
		sk := scopeKey{kind: entry.Scope.Kind, key: entry.Scope.Key}
		b, ok := grouped[sk]
		if !ok {
			b = &bucket{
				scope:  entry.Scope,
				byNorm: map[string]string{},
			}
			grouped[sk] = b
		}
		norm := NormalizeStatement(entry.Statement)
		if _, exists := b.byNorm[norm]; !exists {
			b.byNorm[norm] = entry.Statement
		}
		b.entryIDs = append(b.entryIDs, entry.ID)
	}

	keys := make([]scopeKey, 0, len(grouped))
	for k := range grouped {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].kind != keys[j].kind {
			return keys[i].kind < keys[j].kind
		}
		return keys[i].key < keys[j].key
	})

	out := make([]ConflictGroup, 0)
	for _, k := range keys {
		b := grouped[k]
		if len(b.byNorm) < 2 {
			continue
		}
		statements := make([]string, 0, len(b.byNorm))
		for _, stmt := range b.byNorm {
			statements = append(statements, stmt)
		}
		sort.Strings(statements)
		ids := append([]string(nil), b.entryIDs...)
		sort.Strings(ids)
		out = append(out, ConflictGroup{
			Scope:      b.scope,
			EntryIDs:   ids,
			Statements: statements,
		})
	}
	return out
}
