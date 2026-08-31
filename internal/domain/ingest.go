package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IngestRunStatus is the lifecycle of a git ingest operation.
type IngestRunStatus string

const (
	IngestRunRunning   IngestRunStatus = "running"
	IngestRunSucceeded IngestRunStatus = "succeeded"
	IngestRunFailed    IngestRunStatus = "failed"
)

// Skip reasons for processed SHAs that did not produce lore.
const (
	IngestSkipMerge       = "merge"
	IngestSkipEmpty       = "empty"
	IngestSkipNoisy       = "noisy"
	IngestSkipNoRationale = "no_rationale"
	IngestSkipTooLong     = "too_long"
	IngestSkipUnmerged    = "unmerged"
	IngestSkipBot         = "bot"
)

// GitCommitSnapshot is a commit read from a local repository.
type GitCommitSnapshot struct {
	SHA         string
	Author      string
	CommittedAt time.Time
	Subject     string
	Body        string
	Message     string
	ParentCount int
	Paths       []string
}

// IngestRun is an observable git ingest operation for one repository.
type IngestRun struct {
	ID               string
	Scope            Scope
	ActorID          string
	LocalPath        string
	Status           IngestRunStatus
	CommitsSeen      int
	CommitsSkipped   int
	CandidatesStored int
	CursorSHA        string
	CursorAt         *time.Time
	ErrorSummary     string
	StartedAt        time.Time
	FinishedAt       *time.Time
}

// IngestCursor is the per-repository watermark of last processed commit.
type IngestCursor struct {
	Scope           Scope
	LastSHA         string
	LastCommittedAt time.Time
	UpdatedAt       time.Time
}

// ProcessedSHA records that a commit was considered for a repository.
type ProcessedSHA struct {
	Scope       Scope
	SHA         string
	LoreEntryID string
	Skipped     bool
	SkipReason  string
	ProcessedAt time.Time
}

// NewIngestRunInput is input for starting a run.
type NewIngestRunInput struct {
	Scope     Scope
	ActorID   string
	LocalPath string
	ID        string
	Now       time.Time
}

// NewIngestRun creates a running ingest run.
func NewIngestRun(in NewIngestRunInput) (IngestRun, error) {
	actor := strings.TrimSpace(in.ActorID)
	path := strings.TrimSpace(in.LocalPath)
	if actor == "" {
		return IngestRun{}, validationError("actor must be non-empty")
	}
	if path == "" {
		return IngestRun{}, validationError("path must be non-empty")
	}
	if in.Scope.Kind != ScopeKindRepository {
		return IngestRun{}, validationError("ingest scope kind must be repository")
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return IngestRun{
		ID:        id,
		Scope:     in.Scope,
		ActorID:   actor,
		LocalPath: path,
		Status:    IngestRunRunning,
		StartedAt: now,
	}, nil
}

// MarkSucceeded completes a successful run.
func (r IngestRun) MarkSucceeded(now time.Time) IngestRun {
	t := now.UTC()
	r.Status = IngestRunSucceeded
	r.FinishedAt = &t
	r.ErrorSummary = ""
	return r
}

// MarkFailed completes a failed run.
func (r IngestRun) MarkFailed(now time.Time, summary string) IngestRun {
	t := now.UTC()
	r.Status = IngestRunFailed
	r.FinishedAt = &t
	r.ErrorSummary = strings.TrimSpace(summary)
	return r
}

// PullRequestComment is a review or conversation comment on a PR.
type PullRequestComment struct {
	HTMLURL     string
	Body        string
	AuthorLogin string
	AuthorIsBot bool
}

// PullRequestSnapshot is a pull request read from GitHub.
type PullRequestSnapshot struct {
	Number          int
	NodeID          string
	Owner           string
	Repo            string
	Title           string
	Body            string
	AuthorLogin     string
	AuthorIsBot     bool
	Merged          bool
	MergedAt        *time.Time
	HTMLURL         string
	Files           []string
	ReviewComments  []PullRequestComment
	IssueComments   []PullRequestComment
	LinkedIssueURLs []string
}

// PREvidenceValue formats evidence type pr: owner/repo#number.
func PREvidenceValue(owner, repo string, number int) string {
	return fmt.Sprintf("%s/%s#%d", strings.TrimSpace(owner), strings.TrimSpace(repo), number)
}

// GitHubRepoFromScopeKey maps github.com/{owner}/{repo} to GitHub owner/repo.
func GitHubRepoFromScopeKey(key string) (owner, repo string, err error) {
	const prefix = "github.com/"
	trimmed := strings.TrimSpace(key)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", "", validationError("repository scope key must be github.com/{owner}/{repo}")
	}
	rest := strings.TrimPrefix(trimmed, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", validationError("repository scope key must be github.com/{owner}/{repo}")
	}
	return parts[0], parts[1], nil
}

// PRIngestRun is an observable pull-request ingest operation for one repository.
type PRIngestRun struct {
	ID               string
	Scope            Scope
	ActorID          string
	PRNumber         int
	Status           IngestRunStatus
	PRsSeen          int
	PRsSkipped       int
	CandidatesStored int
	CursorPR         int
	CursorAt         *time.Time
	ErrorSummary     string
	StartedAt        time.Time
	FinishedAt       *time.Time
}

// PRIngestCursor is the per-repository watermark of last processed merged PR.
type PRIngestCursor struct {
	Scope        Scope
	LastPR       int
	LastMergedAt time.Time
	UpdatedAt    time.Time
}

// ProcessedPR records that a pull request was considered for a repository.
type ProcessedPR struct {
	Scope       Scope
	PRNumber    int
	NodeID      string
	LoreEntryID string
	Skipped     bool
	SkipReason  string
	ProcessedAt time.Time
}

// NewPRIngestRunInput is input for starting a PR ingest run.
type NewPRIngestRunInput struct {
	Scope    Scope
	ActorID  string
	PRNumber int
	ID       string
	Now      time.Time
}

// NewPRIngestRun creates a running PR ingest run.
func NewPRIngestRun(in NewPRIngestRunInput) (PRIngestRun, error) {
	actor := strings.TrimSpace(in.ActorID)
	if actor == "" {
		return PRIngestRun{}, validationError("actor must be non-empty")
	}
	if in.Scope.Kind != ScopeKindRepository {
		return PRIngestRun{}, validationError("ingest scope kind must be repository")
	}
	if _, _, err := GitHubRepoFromScopeKey(in.Scope.Key); err != nil {
		return PRIngestRun{}, err
	}
	if in.PRNumber < 0 {
		return PRIngestRun{}, validationError("pr number must be non-negative")
	}
	id := strings.TrimSpace(in.ID)
	if id == "" {
		id = uuid.NewString()
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return PRIngestRun{
		ID:        id,
		Scope:     in.Scope,
		ActorID:   actor,
		PRNumber:  in.PRNumber,
		Status:    IngestRunRunning,
		StartedAt: now,
	}, nil
}

// MarkSucceeded completes a successful PR run.
func (r PRIngestRun) MarkSucceeded(now time.Time) PRIngestRun {
	t := now.UTC()
	r.Status = IngestRunSucceeded
	r.FinishedAt = &t
	r.ErrorSummary = ""
	return r
}

// MarkFailed completes a failed PR run.
func (r PRIngestRun) MarkFailed(now time.Time, summary string) PRIngestRun {
	t := now.UTC()
	r.Status = IngestRunFailed
	r.FinishedAt = &t
	r.ErrorSummary = strings.TrimSpace(summary)
	return r
}
