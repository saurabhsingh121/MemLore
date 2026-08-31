package ports

import (
	"context"
	"fmt"
	"time"

	"github.com/memlore/memlore/internal/domain"
)

// GitLogQuery selects commits from a local working copy.
type GitLogQuery struct {
	Path       string
	After      *time.Time
	MaxCommits int
}

// GitReader lists commits from a local git directory.
type GitReader interface {
	ListCommits(ctx context.Context, q GitLogQuery) ([]domain.GitCommitSnapshot, error)
}

// GitNotRepositoryError is returned when path is not a git working copy.
type GitNotRepositoryError struct {
	Path string
}

func (e *GitNotRepositoryError) Error() string {
	if e.Path == "" {
		return "path is not a git repository"
	}
	return fmt.Sprintf("path is not a git repository: %s", e.Path)
}
