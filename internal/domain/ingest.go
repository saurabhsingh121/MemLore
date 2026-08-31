package domain

import (
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
