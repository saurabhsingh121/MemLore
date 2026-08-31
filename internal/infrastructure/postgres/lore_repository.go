package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/application/ports"
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

func (r *LoreRepository) SearchRelevant(ctx context.Context, opts ports.SearchRelevantOpts) ([]domain.LoreEntry, error) {
	pattern := ilikePattern(opts.Query)
	if pattern == "" {
		return []domain.LoreEntry{}, nil
	}
	fetchLimit := int32(opts.Limit)
	if fetchLimit <= 0 {
		fetchLimit = 50
	} else {
		// Over-fetch so Go multi-token filter still has room.
		fetchLimit = fetchLimit * 5
		if fetchLimit < 20 {
			fetchLimit = 20
		}
		if fetchLimit > 200 {
			fetchLimit = 200
		}
	}

	var rows []sqlc.LoreEntry
	var err error
	if opts.Scope != nil {
		rows, err = r.q.SearchLoreEntriesByStatementScoped(ctx, sqlc.SearchLoreEntriesByStatementScopedParams{
			ScopeKind: string(opts.Scope.Kind),
			ScopeKey:  opts.Scope.Key,
			Column3:   pattern,
			Limit:     fetchLimit,
		})
	} else {
		rows, err = r.q.SearchLoreEntriesByStatementAll(ctx, sqlc.SearchLoreEntriesByStatementAllParams{
			Column1: pattern,
			Limit:   fetchLimit,
		})
	}
	if err != nil {
		return nil, err
	}
	out := make([]domain.LoreEntry, 0, len(rows))
	for _, row := range rows {
		entry, convErr := loreEntryFromRow(row)
		if convErr != nil {
			return nil, convErr
		}
		if !domain.StatementMatchesQuery(entry.Statement, opts.Query) {
			continue
		}
		out = append(out, entry)
	}
	domain.SortLoreByRelevance(out, opts.Query)
	if opts.Limit > 0 && len(out) > opts.Limit {
		out = out[:opts.Limit]
	}
	return out, nil
}

func ilikePattern(query string) string {
	tokens := domain.SignificantQueryTokens(query)
	if len(tokens) == 0 {
		return strings.TrimSpace(query)
	}
	// Prefer longest token for broad SQL candidate fetch.
	best := tokens[0]
	for _, tok := range tokens[1:] {
		if len(tok) > len(best) {
			best = tok
		}
	}
	return best
}
