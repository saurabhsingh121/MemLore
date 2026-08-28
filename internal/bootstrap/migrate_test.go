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

	_, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS outbox_events, audit_records, lore_entries, goose_db_version CASCADE")

	if err := bootstrap.RunMigrations(dsn); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	for _, table := range []string{"lore_entries", "audit_records", "outbox_events"} {
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
	for _, col := range []string{"superseded_by_id", "invalidated_by", "invalidated_at"} {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'lore_entries' AND column_name = $1
			)`, col).Scan(&exists)
		if err != nil {
			t.Fatalf("check column %s: %v", col, err)
		}
		if !exists {
			t.Fatalf("column lore_entries.%s missing after migrate", col)
		}
	}
}
