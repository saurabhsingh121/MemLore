package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.ReviewDecisionRepository = (*ReviewDecisionRepository)(nil)

// ReviewDecisionRepository persists lore review decisions.
type ReviewDecisionRepository struct {
	q *sqlc.Queries
}

func NewReviewDecisionRepository(q *sqlc.Queries) *ReviewDecisionRepository {
	return &ReviewDecisionRepository{q: q}
}

func (r *ReviewDecisionRepository) Insert(ctx context.Context, decision domain.ReviewDecision) error {
	err := r.q.InsertLoreReviewDecision(ctx, reviewDecisionToInsertParams(decision))
	if isUniqueViolation(err) {
		return &domain.ConflictError{Message: "review decision already exists for extract"}
	}
	return err
}

func (r *ReviewDecisionRepository) GetByIdentity(ctx context.Context, identity domain.ExtractIdentity) (domain.ReviewDecision, bool, error) {
	row, err := r.q.GetLoreReviewDecisionByIdentity(ctx, sqlc.GetLoreReviewDecisionByIdentityParams{
		ScopeKind:         string(identity.Scope.Kind),
		ScopeKey:          identity.Scope.Key,
		EvidenceType:      string(identity.EvidenceType),
		EvidenceValue:     identity.EvidenceValue,
		StatementChecksum: identity.StatementChecksum,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ReviewDecision{}, false, nil
		}
		return domain.ReviewDecision{}, false, err
	}
	got, err := reviewDecisionFromRow(row)
	if err != nil {
		return domain.ReviewDecision{}, false, err
	}
	return got, true, nil
}

func (r *ReviewDecisionRepository) GetByLoreID(ctx context.Context, loreEntryID string) (domain.ReviewDecision, bool, error) {
	row, err := r.q.GetLoreReviewDecisionByLoreID(ctx, loreEntryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ReviewDecision{}, false, nil
		}
		return domain.ReviewDecision{}, false, err
	}
	got, err := reviewDecisionFromRow(row)
	if err != nil {
		return domain.ReviewDecision{}, false, err
	}
	return got, true, nil
}

func (r *ReviewDecisionRepository) ListByScope(ctx context.Context, scope domain.Scope) ([]domain.ReviewDecision, error) {
	rows, err := r.q.ListLoreReviewDecisionsByScope(ctx, sqlc.ListLoreReviewDecisionsByScopeParams{
		ScopeKind: string(scope.Kind),
		ScopeKey:  scope.Key,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ReviewDecision, 0, len(rows))
	for _, row := range rows {
		got, err := reviewDecisionFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, got)
	}
	return out, nil
}

func reviewDecisionToInsertParams(d domain.ReviewDecision) sqlc.InsertLoreReviewDecisionParams {
	return sqlc.InsertLoreReviewDecisionParams{
		ID:                d.ID,
		ScopeKind:         string(d.Scope.Kind),
		ScopeKey:          d.Scope.Key,
		EvidenceType:      string(d.EvidenceType),
		EvidenceValue:     d.EvidenceValue,
		StatementChecksum: d.StatementChecksum,
		LoreEntryID:       d.LoreEntryID,
		SuccessorLoreID:   textFromPtr(d.SuccessorLoreID),
		Status:            string(d.Status),
		ActorID:           d.ActorID,
		DecidedAt:         timestamptzFromTime(d.DecidedAt),
	}
}

func reviewDecisionFromRow(row sqlc.LoreReviewDecision) (domain.ReviewDecision, error) {
	kind, err := domain.ParseScopeKind(row.ScopeKind)
	if err != nil {
		return domain.ReviewDecision{}, err
	}
	et, err := domain.ParseEvidenceType(row.EvidenceType)
	if err != nil {
		return domain.ReviewDecision{}, err
	}
	return domain.ReviewDecision{
		ID:                row.ID,
		Scope:             domain.Scope{Kind: kind, Key: row.ScopeKey},
		EvidenceType:      et,
		EvidenceValue:     row.EvidenceValue,
		StatementChecksum: row.StatementChecksum,
		LoreEntryID:       row.LoreEntryID,
		SuccessorLoreID:   ptrFromText(row.SuccessorLoreID),
		Status:            domain.ReviewStatus(row.Status),
		ActorID:           row.ActorID,
		DecidedAt:         timeFromTimestamptz(row.DecidedAt),
	}, nil
}
