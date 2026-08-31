package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.ADRIngestRepository = (*ADRIngestRepository)(nil)

// ADRIngestRepository persists ADR ingest runs, cursors, and processed files.
type ADRIngestRepository struct {
	q *sqlc.Queries
}

func NewADRIngestRepository(q *sqlc.Queries) *ADRIngestRepository {
	return &ADRIngestRepository{q: q}
}

func (r *ADRIngestRepository) InsertRun(ctx context.Context, run domain.ADRIngestRun) error {
	err := r.q.InsertADRIngestRun(ctx, adrIngestRunToInsertParams(run))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "adr ingest already running for repository"}
	}
	return err
}

func (r *ADRIngestRepository) UpdateRun(ctx context.Context, run domain.ADRIngestRun) error {
	return r.q.UpdateADRIngestRun(ctx, adrIngestRunToUpdateParams(run))
}

func (r *ADRIngestRepository) GetRun(ctx context.Context, id string) (domain.ADRIngestRun, error) {
	row, err := r.q.GetADRIngestRun(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ADRIngestRun{}, &domain.NotFoundError{Message: "adr ingest run not found"}
		}
		return domain.ADRIngestRun{}, err
	}
	return adrIngestRunFromRow(row)
}

func (r *ADRIngestRepository) ListRunsByScope(ctx context.Context, scope domain.Scope) ([]domain.ADRIngestRun, error) {
	rows, err := r.q.ListADRIngestRunsByScope(ctx, sqlc.ListADRIngestRunsByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ADRIngestRun, 0, len(rows))
	for _, row := range rows {
		run, err := adrIngestRunFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func (r *ADRIngestRepository) GetCursor(ctx context.Context, scope domain.Scope) (domain.ADRIngestCursor, bool, error) {
	row, err := r.q.GetADRIngestCursor(ctx, sqlc.GetADRIngestCursorParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ADRIngestCursor{}, false, nil
		}
		return domain.ADRIngestCursor{}, false, err
	}
	return adrIngestCursorFromRow(row), true, nil
}

func (r *ADRIngestRepository) UpsertCursor(ctx context.Context, cursor domain.ADRIngestCursor) error {
	return r.q.UpsertADRIngestCursor(ctx, sqlc.UpsertADRIngestCursorParams{
		ScopeKind:    string(cursor.Scope.Kind),
		ScopeKey:     cursor.Scope.Key,
		LastPath:     textFromString(cursor.LastPath),
		LastChecksum: textFromString(cursor.LastChecksum),
		UpdatedAt:    timestamptzFromTime(cursor.UpdatedAt),
	})
}

func (r *ADRIngestRepository) GetProcessedADR(ctx context.Context, scope domain.Scope, relativePath, checksum string) (domain.ProcessedADR, bool, error) {
	row, err := r.q.GetADRIngestFile(ctx, sqlc.GetADRIngestFileParams{
		ScopeKind:    string(scope.Kind),
		ScopeKey:     scope.Key,
		RelativePath: relativePath,
		Checksum:     checksum,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProcessedADR{}, false, nil
		}
		return domain.ProcessedADR{}, false, err
	}
	return processedADRFromRow(row), true, nil
}

func (r *ADRIngestRepository) InsertProcessedADR(ctx context.Context, row domain.ProcessedADR) error {
	err := r.q.InsertADRIngestFile(ctx, processedADRToInsertParams(row))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "adr file already processed"}
	}
	return err
}

func (r *ADRIngestRepository) LatestStoredByPath(ctx context.Context, scope domain.Scope, relativePath string) (domain.ProcessedADR, bool, error) {
	row, err := r.q.LatestStoredADRByPath(ctx, sqlc.LatestStoredADRByPathParams{
		ScopeKind:    string(scope.Kind),
		ScopeKey:     scope.Key,
		RelativePath: relativePath,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProcessedADR{}, false, nil
		}
		return domain.ProcessedADR{}, false, err
	}
	return processedADRFromRow(row), true, nil
}

func (r *ADRIngestRepository) LatestStoredByADRID(ctx context.Context, scope domain.Scope, adrID string) (domain.ProcessedADR, bool, error) {
	row, err := r.q.LatestStoredADRByID(ctx, sqlc.LatestStoredADRByIDParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
		AdrID:     textFromString(adrID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProcessedADR{}, false, nil
		}
		return domain.ProcessedADR{}, false, err
	}
	return processedADRFromRow(row), true, nil
}
