package ingest

import (
	"regexp"
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

var (
	prConventional = regexp.MustCompile(`^(?i)(chore|ci|style|build|deps)(\(.+\))?!?:\s`)
	issueRef       = regexp.MustCompile(`(?i)(?:fixes|closes|resolves)\s+(?:https://github\.com/[^/]+/[^/]+/issues/(\d+)|#(\d+))`)
	issueURL       = regexp.MustCompile(`https://github\.com/[^/]+/[^/]+/issues/\d+`)
)

// PRExtractResult is a conservative candidate decision for one pull request.
type PRExtractResult struct {
	OK              bool
	Statement       string
	SkipReason      string
	UsedCommentURLs []string
}

// ExtractPRCandidate decides whether a PR supports observational lore.
// Unmerged and bot skips win over rationale cues. At most one candidate per PR.
func ExtractPRCandidate(pr domain.PullRequestSnapshot) PRExtractResult {
	if !pr.Merged {
		return PRExtractResult{SkipReason: domain.IngestSkipUnmerged}
	}
	if isNoiseBot(pr.AuthorLogin) {
		return PRExtractResult{SkipReason: domain.IngestSkipBot}
	}

	title := strings.TrimSpace(pr.Title)
	body := strings.TrimSpace(pr.Body)
	humanComments := humanRationaleComments(pr)

	if title == "" && body == "" && len(humanComments) == 0 {
		return PRExtractResult{SkipReason: domain.IngestSkipEmpty}
	}

	titleBody := title
	if body != "" {
		if titleBody != "" {
			titleBody += "\n\n" + body
		} else {
			titleBody = body
		}
	}

	noisyTitle := prConventional.MatchString(title) || versionBump.MatchString(title)
	titleBodyHasCue := rationaleCue.MatchString(title) || rationaleCue.MatchString(body)
	commentsHaveCue := len(humanComments) > 0

	if noisyTitle && !titleBodyHasCue && !commentsHaveCue {
		return PRExtractResult{SkipReason: domain.IngestSkipNoisy}
	}
	if !titleBodyHasCue && !commentsHaveCue {
		return PRExtractResult{SkipReason: domain.IngestSkipNoRationale}
	}

	statement := titleBody
	used := make([]string, 0)
	if !titleBodyHasCue {
		var b strings.Builder
		for i, c := range humanComments {
			if i > 0 {
				b.WriteString("\n\n")
			}
			b.WriteString(strings.TrimSpace(c.Body))
			if strings.TrimSpace(c.HTMLURL) != "" {
				used = append(used, strings.TrimSpace(c.HTMLURL))
			}
		}
		statement = b.String()
	} else {
		for _, c := range humanComments {
			if strings.Contains(statement, strings.TrimSpace(c.Body)) && strings.TrimSpace(c.HTMLURL) != "" {
				used = append(used, strings.TrimSpace(c.HTMLURL))
			}
		}
	}
	if len(statement) > domain.MaxStatementLength {
		return PRExtractResult{SkipReason: domain.IngestSkipTooLong}
	}
	if strings.TrimSpace(statement) == "" {
		return PRExtractResult{SkipReason: domain.IngestSkipEmpty}
	}
	return PRExtractResult{OK: true, Statement: statement, UsedCommentURLs: used}
}

// LinkedIssueURLs returns issue URLs from the snapshot or parsed from the body.
func LinkedIssueURLs(pr domain.PullRequestSnapshot) []string {
	if len(pr.LinkedIssueURLs) > 0 {
		return uniqueNonEmpty(pr.LinkedIssueURLs)
	}
	return parseLinkedIssues(pr.Owner, pr.Repo, pr.Body)
}

func parseLinkedIssues(owner, repo, body string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range issueURL.FindAllString(body, -1) {
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}
	for _, m := range issueRef.FindAllStringSubmatch(body, -1) {
		n := m[1]
		if n == "" {
			n = m[2]
		}
		if n == "" || owner == "" || repo == "" {
			continue
		}
		u := "https://github.com/" + owner + "/" + repo + "/issues/" + n
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func uniqueNonEmpty(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func humanRationaleComments(pr domain.PullRequestSnapshot) []domain.PullRequestComment {
	var out []domain.PullRequestComment
	for _, c := range append(append([]domain.PullRequestComment{}, pr.ReviewComments...), pr.IssueComments...) {
		if c.AuthorIsBot || isNoiseBot(c.AuthorLogin) {
			continue
		}
		if !rationaleCue.MatchString(c.Body) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func isNoiseBot(login string) bool {
	n := strings.ToLower(strings.TrimSpace(login))
	n = strings.ReplaceAll(n, "[bot]", "")
	switch n {
	case "dependabot", "renovate", "github-actions":
		return true
	}
	return false
}
