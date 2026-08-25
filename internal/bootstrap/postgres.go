package bootstrap

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/infrastructure/postgres"
)

// NormalizePostgresDSN converts SQLAlchemy-style DSNs to pgx format.
func NormalizePostgresDSN(dsn string) string {
	return strings.Replace(dsn, "postgresql+psycopg://", "postgresql://", 1)
}

// PostgresUnitOfWorkFactory returns a factory backed by a connection pool.
func PostgresUnitOfWorkFactory(pool *pgxpool.Pool) ports.UnitOfWorkFactory {
	return func(ctx context.Context) (ports.UnitOfWork, error) {
		return postgres.BeginUnitOfWork(ctx, pool)
	}
}
