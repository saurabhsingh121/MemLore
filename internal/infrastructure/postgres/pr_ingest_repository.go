package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.PRIngestRepository = (*PRIngestRepository)(nil)

// PRIngestRepository persists PR ingest runs, cursors, and processed PRs.
type PRIngestRepository struct {
	q *sqlc.Queries
}

func NewPRIngestRepository(q *sqlc.Queries) *PRIngestRepository {
	return &PRIngestRepository{q: q}
}

func (r *PRIngestRepository) InsertRun(ctx context.Context, run domain.PRIngestRun) error {
	err := r.q.InsertPRIngestRun(ctx, prIngestRunToInsertParams(run))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "pr ingest already running for repository"}
	}
	return err
}

func (r *PRIngestRepository) UpdateRun(ctx context.Context, run domain.PRIngestRun) error {
	return r.q.UpdatePRIngestRun(ctx, prIngestRunToUpdateParams(run))
}

func (r *PRIngestRepository) GetRun(ctx context.Context, id string) (domain.PRIngestRun, error) {
	row, err := r.q.GetPRIngestRun(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PRIngestRun{}, &domain.NotFoundError{Message: "pr ingest run not found"}
		}
		return domain.PRIngestRun{}, err
	}
	return prIngestRunFromRow(row)
}

func (r *PRIngestRepository) ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.PRIngestRun, error) {
	rows, err := r.q.ListPRIngestRunsByScope(ctx, sqlc.ListPRIngestRunsByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PRIngestRun, 0, len(rows))
	for _, row := range rows {
		run, err := prIngestRunFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func (r *PRIngestRepository) GetCursor(ctx context.Context, scope domain.Scope) (domain.PRIngestCursor, bool, error) {
	row, err := r.q.GetPRIngestCursor(ctx, sqlc.GetPRIngestCursorParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PRIngestCursor{}, false, nil
		}
		return domain.PRIngestCursor{}, false, err
	}
	return prIngestCursorFromRow(row), true, nil
}

func (r *PRIngestRepository) UpsertCursor(ctx context.Context, cursor domain.PRIngestCursor) error {
	return r.q.UpsertPRIngestCursor(ctx, sqlc.UpsertPRIngestCursorParams{
		ScopeKind:    string(cursor.Scope.Kind),
		ScopeKey:     cursor.Scope.Key,
		LastPr:       int32(cursor.LastPR),
		LastMergedAt: timestamptzFromTime(cursor.LastMergedAt),
		UpdatedAt:    timestamptzFromTime(cursor.UpdatedAt),
	})
}

func (r *PRIngestRepository) GetProcessedPR(ctx context.Context, scope domain.Scope, prNumber int) (domain.ProcessedPR, bool, error) {
	row, err := r.q.GetPRIngestPR(ctx, sqlc.GetPRIngestPRParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
		PrNumber:  int32(prNumber),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProcessedPR{}, false, nil
		}
		return domain.ProcessedPR{}, false, err
	}
	return processedPRFromRow(row), true, nil
}

func (r *PRIngestRepository) InsertProcessedPR(ctx context.Context, row domain.ProcessedPR) error {
	err := r.q.InsertPRIngestPR(ctx, processedPRToInsertParams(row))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "pull request already processed"}
	}
	return err
}
