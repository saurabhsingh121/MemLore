package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.IngestRepository = (*IngestRepository)(nil)

// IngestRepository persists git ingest runs, cursors, and processed SHAs.
type IngestRepository struct {
	q *sqlc.Queries
}

func NewIngestRepository(q *sqlc.Queries) *IngestRepository {
	return &IngestRepository{q: q}
}

func (r *IngestRepository) InsertRun(ctx context.Context, run domain.IngestRun) error {
	err := r.q.InsertGitIngestRun(ctx, ingestRunToInsertParams(run))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "ingest already running for repository"}
	}
	return err
}

func (r *IngestRepository) UpdateRun(ctx context.Context, run domain.IngestRun) error {
	return r.q.UpdateGitIngestRun(ctx, ingestRunToUpdateParams(run))
}

func (r *IngestRepository) GetRun(ctx context.Context, id string) (domain.IngestRun, error) {
	row, err := r.q.GetGitIngestRun(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IngestRun{}, &domain.NotFoundError{Message: "ingest run not found"}
		}
		return domain.IngestRun{}, err
	}
	return ingestRunFromRow(row)
}

func (r *IngestRepository) ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.IngestRun, error) {
	rows, err := r.q.ListGitIngestRunsByScope(ctx, sqlc.ListGitIngestRunsByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.IngestRun, 0, len(rows))
	for _, row := range rows {
		run, err := ingestRunFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func (r *IngestRepository) GetCursor(ctx context.Context, scope domain.Scope) (domain.IngestCursor, bool, error) {
	row, err := r.q.GetGitIngestCursor(ctx, sqlc.GetGitIngestCursorParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.IngestCursor{}, false, nil
		}
		return domain.IngestCursor{}, false, err
	}
	return ingestCursorFromRow(row), true, nil
}

func (r *IngestRepository) UpsertCursor(ctx context.Context, cursor domain.IngestCursor) error {
	return r.q.UpsertGitIngestCursor(ctx, sqlc.UpsertGitIngestCursorParams{
		ScopeKind:       string(cursor.Scope.Kind),
		ScopeKey:        cursor.Scope.Key,
		LastSha:         cursor.LastSHA,
		LastCommittedAt: timestamptzFromTime(cursor.LastCommittedAt),
		UpdatedAt:       timestamptzFromTime(cursor.UpdatedAt),
	})
}

func (r *IngestRepository) GetProcessedSHA(ctx context.Context, scope domain.Scope, sha string) (domain.ProcessedSHA, bool, error) {
	row, err := r.q.GetGitIngestSHA(ctx, sqlc.GetGitIngestSHAParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
		Sha:       sha,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProcessedSHA{}, false, nil
		}
		return domain.ProcessedSHA{}, false, err
	}
	return processedSHAFromRow(row), true, nil
}

func (r *IngestRepository) InsertProcessedSHA(ctx context.Context, row domain.ProcessedSHA) error {
	err := r.q.InsertGitIngestSHA(ctx, processedSHAToInsertParams(row))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "commit already processed"}
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
