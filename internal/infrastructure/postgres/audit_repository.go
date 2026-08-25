package postgres

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

// AuditRepository implements audit persistence via sqlc.
type AuditRepository struct {
	q *sqlc.Queries
}

func NewAuditRepository(q *sqlc.Queries) *AuditRepository {
	return &AuditRepository{q: q}
}

func (r *AuditRepository) Add(ctx context.Context, record domain.AuditRecord) error {
	return r.q.InsertAuditRecord(ctx, auditRecordToInsertParams(record))
}

func (r *AuditRepository) ListByTarget(ctx context.Context, targetID string) ([]domain.AuditRecord, error) {
	rows, err := r.q.ListAuditRecordsByTarget(ctx, targetID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AuditRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, auditRecordFromRow(row))
	}
	return out, nil
}
