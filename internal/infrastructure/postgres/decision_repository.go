package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.DecisionRepository = (*DecisionRepository)(nil)

// DecisionRepository persists human-recorded Decisions.
type DecisionRepository struct {
	q *sqlc.Queries
}

func NewDecisionRepository(q *sqlc.Queries) *DecisionRepository {
	return &DecisionRepository{q: q}
}

func (r *DecisionRepository) Add(ctx context.Context, decision domain.Decision) error {
	if err := r.q.InsertDecision(ctx, decisionToInsertParams(decision)); err != nil {
		if isUniqueViolation(err) {
			return &domain.ConflictError{Message: "decision already exists"}
		}
		return err
	}
	return r.replaceChildren(ctx, decision)
}

func (r *DecisionRepository) Get(ctx context.Context, id string) (domain.Decision, error) {
	row, err := r.q.GetDecision(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Decision{}, &domain.NotFoundError{Message: fmt.Sprintf("decision %s not found", id)}
		}
		return domain.Decision{}, err
	}
	return r.decisionFromRow(ctx, row)
}

func (r *DecisionRepository) Save(ctx context.Context, decision domain.Decision) error {
	if err := r.q.UpdateDecision(ctx, decisionToUpdateParams(decision)); err != nil {
		return err
	}
	if err := r.q.DeleteDecisionAlternatives(ctx, decision.ID); err != nil {
		return err
	}
	if err := r.q.DeleteDecisionComponents(ctx, decision.ID); err != nil {
		return err
	}
	return r.replaceChildren(ctx, decision)
}

func (r *DecisionRepository) ListByScope(ctx context.Context, scope domain.Scope) ([]domain.Decision, error) {
	rows, err := r.q.ListDecisionsByScope(ctx, sqlc.ListDecisionsByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Decision, 0, len(rows))
	for _, row := range rows {
		got, err := r.decisionFromRow(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, got)
	}
	return out, nil
}

func (r *DecisionRepository) replaceChildren(ctx context.Context, decision domain.Decision) error {
	for i, alt := range decision.Alternatives {
		if err := r.q.InsertDecisionAlternative(ctx, sqlc.InsertDecisionAlternativeParams{
			DecisionID: decision.ID,
			Position:   int32(i),
			Label:      alt.Label,
			Note:       alt.Note,
		}); err != nil {
			return err
		}
	}
	for i, name := range decision.AffectedComponents {
		if err := r.q.InsertDecisionComponent(ctx, sqlc.InsertDecisionComponentParams{
			DecisionID: decision.ID,
			Position:   int32(i),
			Name:       name,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *DecisionRepository) decisionFromRow(ctx context.Context, row sqlc.Decision) (domain.Decision, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.Decision{}, err
	}
	alts, err := r.q.ListDecisionAlternatives(ctx, row.ID)
	if err != nil {
		return domain.Decision{}, err
	}
	comps, err := r.q.ListDecisionComponents(ctx, row.ID)
	if err != nil {
		return domain.Decision{}, err
	}
	alternatives := make([]domain.DecisionAlternative, 0, len(alts))
	for _, a := range alts {
		alternatives = append(alternatives, domain.DecisionAlternative{Label: a.Label, Note: a.Note})
	}
	components := make([]string, 0, len(comps))
	for _, c := range comps {
		components = append(components, c.Name)
	}
	return domain.Decision{
		ID:                 row.ID,
		Scope:              domain.Scope{Kind: kind, Key: row.ScopeKey},
		Question:           row.Question,
		Choice:             row.Choice,
		Rationale:          row.Rationale,
		Alternatives:       alternatives,
		Consequences:       row.Consequences,
		Owner:              row.Owner,
		DecidedAt:          timeFromTimestamptz(row.DecidedAt),
		AffectedComponents: components,
		Evidence:           []domain.EvidenceReference{},
		SourceKind:         domain.DecisionSourceKind(row.SourceKind),
		SupersededByID:     ptrFromText(row.SupersededByID),
		CreatedBy:          row.CreatedBy,
		CreatedAt:          timeFromTimestamptz(row.CreatedAt),
	}, nil
}

func decisionToInsertParams(d domain.Decision) sqlc.InsertDecisionParams {
	return sqlc.InsertDecisionParams{
		ID:             d.ID,
		ScopeKind:      string(d.Scope.Kind),
		ScopeKey:       d.Scope.Key,
		Question:       d.Question,
		Choice:         d.Choice,
		Rationale:      d.Rationale,
		Consequences:   d.Consequences,
		Owner:          d.Owner,
		DecidedAt:      timestamptzFromTime(d.DecidedAt),
		SourceKind:     string(d.SourceKind),
		SupersededByID: textFromPtr(d.SupersededByID),
		CreatedBy:      d.CreatedBy,
		CreatedAt:      timestamptzFromTime(d.CreatedAt),
		UpdatedAt:      timestamptzFromTime(d.CreatedAt),
	}
}

func decisionToUpdateParams(d domain.Decision) sqlc.UpdateDecisionParams {
	return sqlc.UpdateDecisionParams{
		ID:             d.ID,
		Question:       d.Question,
		Choice:         d.Choice,
		Rationale:      d.Rationale,
		Consequences:   d.Consequences,
		Owner:          d.Owner,
		DecidedAt:      timestamptzFromTime(d.DecidedAt),
		SourceKind:     string(d.SourceKind),
		SupersededByID: textFromPtr(d.SupersededByID),
		UpdatedAt:      timestamptzFromTime(d.CreatedAt),
	}
}
