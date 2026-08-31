package ingest

import (
	"regexp"
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

var (
	mergeSubject = regexp.MustCompile(`(?i)^merge(\s+pull request|\s+branch)?\s`)
	conventional = regexp.MustCompile(`^(?i)(chore|ci|style|build)(\(.+\))?!?:\s`)
	versionBump  = regexp.MustCompile(`(?i)^(v?\d+\.\d+(\.\d+)?)(\s.*)?$|^(bump version|release)\b`)
	rationaleCue = regexp.MustCompile(`(?i)\b(because|so that|workaround|migration|breaking|to fix|instead of|avoid|constraint)\b|^why:`)
)

// ExtractResult is a conservative candidate decision for one commit.
type ExtractResult struct {
	OK         bool
	Statement  string
	SkipReason string
}

// ExtractCandidate decides whether a commit supports observational lore.
// Skip rules win over rationale cues. At most the commit message is used;
// paths never invent a decision.
func ExtractCandidate(commit domain.GitCommitSnapshot) ExtractResult {
	message := strings.TrimSpace(commit.Message)
	if message == "" {
		message = strings.TrimSpace(commit.Subject + "\n\n" + commit.Body)
	}
	if message == "" {
		return ExtractResult{SkipReason: domain.IngestSkipEmpty}
	}
	if len(message) > domain.MaxStatementLength {
		return ExtractResult{SkipReason: domain.IngestSkipTooLong}
	}
	subject := strings.TrimSpace(commit.Subject)
	if subject == "" {
		subject = firstLine(message)
	}
	body := strings.TrimSpace(commit.Body)
	if body == "" {
		body = restAfterFirstLine(message)
	}

	if commit.ParentCount > 1 || mergeSubject.MatchString(subject) {
		return ExtractResult{SkipReason: domain.IngestSkipMerge}
	}
	if conventional.MatchString(subject) || versionBump.MatchString(subject) {
		return ExtractResult{SkipReason: domain.IngestSkipNoisy}
	}
	if !rationaleCue.MatchString(subject) && !rationaleCue.MatchString(body) && !rationaleCue.MatchString(message) {
		return ExtractResult{SkipReason: domain.IngestSkipNoRationale}
	}
	return ExtractResult{OK: true, Statement: message}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

func restAfterFirstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[i+1:])
	}
	return ""
}
