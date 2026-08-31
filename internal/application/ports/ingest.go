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
