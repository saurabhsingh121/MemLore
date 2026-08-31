package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// IngestRepository persists ingest runs, cursors, and processed SHAs.
type IngestRepository interface {
	InsertRun(ctx context.Context, run domain.IngestRun) error
	UpdateRun(ctx context.Context, run domain.IngestRun) error
	GetRun(ctx context.Context, id string) (domain.IngestRun, error)
	ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.IngestRun, error)
	GetCursor(ctx context.Context, scope domain.Scope) (domain.IngestCursor, bool, error)
	UpsertCursor(ctx context.Context, cursor domain.IngestCursor) error
	GetProcessedSHA(ctx context.Context, scope domain.Scope, sha string) (domain.ProcessedSHA, bool, error)
	InsertProcessedSHA(ctx context.Context, row domain.ProcessedSHA) error
}

// PRIngestRepository persists PR ingest runs, cursors, and processed PRs.
type PRIngestRepository interface {
	InsertRun(ctx context.Context, run domain.PRIngestRun) error
	UpdateRun(ctx context.Context, run domain.PRIngestRun) error
	GetRun(ctx context.Context, id string) (domain.PRIngestRun, error)
	ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.PRIngestRun, error)
	GetCursor(ctx context.Context, scope domain.Scope) (domain.PRIngestCursor, bool, error)
	UpsertCursor(ctx context.Context, cursor domain.PRIngestCursor) error
	GetProcessedPR(ctx context.Context, scope domain.Scope, prNumber int) (domain.ProcessedPR, bool, error)
	InsertProcessedPR(ctx context.Context, row domain.ProcessedPR) error
}

// ADRIngestRepository persists ADR ingest runs, cursors, and processed files.
type ADRIngestRepository interface {
	InsertRun(ctx context.Context, run domain.ADRIngestRun) error
	UpdateRun(ctx context.Context, run domain.ADRIngestRun) error
	GetRun(ctx context.Context, id string) (domain.ADRIngestRun, error)
	ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.ADRIngestRun, error)
	GetCursor(ctx context.Context, scope domain.Scope) (domain.ADRIngestCursor, bool, error)
	UpsertCursor(ctx context.Context, cursor domain.ADRIngestCursor) error
	GetProcessedADR(ctx context.Context, scope domain.Scope, relativePath, checksum string) (domain.ProcessedADR, bool, error)
	InsertProcessedADR(ctx context.Context, row domain.ProcessedADR) error
	LatestStoredByPath(ctx context.Context, scope domain.Scope, relativePath string) (domain.ProcessedADR, bool, error)
	LatestStoredByADRID(ctx context.Context, scope domain.Scope, adrID string) (domain.ProcessedADR, bool, error)
}
