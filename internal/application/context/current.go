package context

import "github.com/memlore/memlore/internal/domain"

// FilterCurrent returns only current (not superseded, not invalidated) entries.
func FilterCurrent(entries []domain.LoreEntry) []domain.LoreEntry {
	out := make([]domain.LoreEntry, 0, len(entries))
	for _, entry := range entries {
		if domain.IsCurrent(entry) {
			out = append(out, entry)
		}
	}
	return out
}
