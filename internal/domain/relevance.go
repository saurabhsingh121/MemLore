package domain

import (
	"sort"
	"strings"
	"unicode"
)

// SignificantQueryTokens returns lowercased tokens of length >= 2.
func SignificantQueryTokens(query string) []string {
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(query)), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
	out := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, tok := range fields {
		if len(tok) < 2 {
			continue
		}
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	return out
}

// StatementMatchesQuery reports whether statement is relevant to query.
// Rules: empty query → false; case-insensitive — statement contains the full
// trimmed query, OR at least one significant token (len ≥ 2) appears.
func StatementMatchesQuery(statement, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return false
	}
	s := strings.ToLower(statement)
	full := strings.ToLower(q)
	if strings.Contains(s, full) {
		return true
	}
	tokens := SignificantQueryTokens(q)
	if len(tokens) == 0 {
		return false
	}
	for _, tok := range tokens {
		if strings.Contains(s, tok) {
			return true
		}
	}
	return false
}

// RelevanceScore returns a coarse score for ranking (higher is better).
func RelevanceScore(statement, query string) int {
	if !StatementMatchesQuery(statement, query) {
		return 0
	}
	s := strings.ToLower(statement)
	full := strings.ToLower(strings.TrimSpace(query))
	score := 0
	if strings.Contains(s, full) {
		score += 100
	}
	for _, tok := range SignificantQueryTokens(query) {
		if strings.Contains(s, tok) {
			score += 10
		}
	}
	return score
}

// SortLoreByRelevance sorts entries in place: relevance desc, verified first,
// then created_at desc. Non-matching entries sort last.
func SortLoreByRelevance(entries []LoreEntry, query string) {
	sort.SliceStable(entries, func(i, j int) bool {
		si := RelevanceScore(entries[i].Statement, query)
		sj := RelevanceScore(entries[j].Statement, query)
		if si != sj {
			return si > sj
		}
		vi := entries[i].VerificationStatus == VerificationVerified
		vj := entries[j].VerificationStatus == VerificationVerified
		if vi != vj {
			return vi && !vj
		}
		return entries[i].CreatedAt.After(entries[j].CreatedAt)
	})
}
