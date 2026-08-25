package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

// UnitOfWork runs lore and audit repositories in a single pgx transaction.
type UnitOfWork struct {
	tx      pgx.Tx
	queries *sqlc.Queries
	lore    ports.LoreRepository
	audits  ports.AuditRepository
}

// BeginUnitOfWork starts a transaction-bound unit of work.
func BeginUnitOfWork(ctx context.Context, pool *pgxpool.Pool) (*UnitOfWork, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	q := sqlc.New(tx)
	return &UnitOfWork{
		tx:      tx,
		queries: q,
		lore:    NewLoreRepository(q),
		audits:  NewAuditRepository(q),
	}, nil
}

func (u *UnitOfWork) LoreEntries() ports.LoreRepository {
	return u.lore
}

func (u *UnitOfWork) Audits() ports.AuditRepository {
	return u.audits
}

func (u *UnitOfWork) Commit(ctx context.Context) error {
	return u.tx.Commit(ctx)
}

func (u *UnitOfWork) Rollback(ctx context.Context) error {
	return u.tx.Rollback(ctx)
}
