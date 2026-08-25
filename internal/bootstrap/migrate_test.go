//go:build integration

package bootstrap_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/bootstrap"
)

func TestRunMigrationsCreatesTables(t *testing.T) {
	dsn := os.Getenv("MEMLORE_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://memlore:memlore@localhost:15432/memlore"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, bootstrap.NormalizePostgresDSN(dsn))
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if pingErr := pool.Ping(ctx); pingErr != nil {
		t.Skipf("postgres ping failed: %v", pingErr)
	}
	t.Cleanup(pool.Close)

	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS audit_records, lore_entries, goose_db_version")

	if err := bootstrap.RunMigrations(dsn); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	for _, table := range []string{"lore_entries", "audit_records"} {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("check %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("table %s missing after migrate", table)
		}
	}
}
