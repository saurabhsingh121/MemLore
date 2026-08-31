package ports

import (
	"context"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// PullRequestQuery selects merged PRs from GitHub for a repository.
type PullRequestQuery struct {
	Owner         string
	Repo          string
	AfterMergedAt *time.Time
	Number        int // if >0, fetch that PR only
	MaxPRs        int
}

// PullRequestReader lists pull requests from GitHub.
type PullRequestReader interface {
	ListPullRequests(ctx context.Context, q PullRequestQuery) ([]domain.PullRequestSnapshot, error)
}

// GitHubError is a GitHub API or configuration failure.
type GitHubError struct {
	Message    string
	StatusCode int
}

func (e *GitHubError) Error() string {
	if e.Message == "" {
		return "github request failed"
	}
	return e.Message
}
