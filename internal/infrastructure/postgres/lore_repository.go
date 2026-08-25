package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

// LoreRepository implements lore persistence via sqlc.
type LoreRepository struct {
	q *sqlc.Queries
}

func NewLoreRepository(q *sqlc.Queries) *LoreRepository {
	return &LoreRepository{q: q}
}

func (r *LoreRepository) Add(ctx context.Context, entry domain.LoreEntry) error {
	params, err := loreEntryToInsertParams(entry)
	if err != nil {
		return err
	}
	return r.q.InsertLoreEntry(ctx, params)
}

func (r *LoreRepository) Get(ctx context.Context, id string) (domain.LoreEntry, error) {
	row, err := r.q.GetLoreEntry(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LoreEntry{}, &domain.NotFoundError{Message: "lore entry not found"}
		}
		return domain.LoreEntry{}, err
	}
	return loreEntryFromRow(row)
}

func (r *LoreRepository) Save(ctx context.Context, entry domain.LoreEntry) error {
	params, err := loreEntryToUpdateParams(entry)
	if err != nil {
		return err
	}
	return r.q.UpdateLoreEntry(ctx, params)
}

func (r *LoreRepository) ListByScope(ctx context.Context, scope domain.Scope) ([]domain.LoreEntry, error) {
	rows, err := r.q.ListLoreEntriesByScope(ctx, sqlc.ListLoreEntriesByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.LoreEntry, 0, len(rows))
	for _, row := range rows {
		entry, convErr := loreEntryFromRow(row)
		if convErr != nil {
			return nil, convErr
		}
		out = append(out, entry)
	}
	return out, nil
}
