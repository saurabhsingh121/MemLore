//go:build integration

package migrations_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestLoreAuditMigrationCreatesTablesOnPostgres(t *testing.T) {
	dsn := os.Getenv("MEMLORE_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgresql://memlore:memlore@localhost:15432/memlore"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	defer conn.Close(ctx)

	dbName := "memlore_f101_" + strings.ReplaceAll(time.Now().Format("150405.000"), ".", "")
	_, err = conn.Exec(ctx, "CREATE DATABASE "+dbName)
	if err != nil {
		t.Skipf("cannot create test database: %v", err)
	}
	defer func() {
		cleanup, cErr := pgx.Connect(ctx, dsn)
		if cErr == nil {
			_, _ = cleanup.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName)
			cleanup.Close(ctx)
		}
	}()

	testDSN := strings.Replace(dsn, "/memlore", "/"+dbName, 1)
	testConn, err := pgx.Connect(ctx, testDSN)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	defer testConn.Close(ctx)

	upSQL := extractGooseUpSQL(t)
	_, err = testConn.Exec(ctx, upSQL)
	if err != nil {
		t.Fatalf("apply migration sql: %v", err)
	}

	for _, table := range []string{"lore_entries", "audit_records"} {
		var exists bool
		err = testConn.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`, table).Scan(&exists)
		if err != nil {
			t.Fatalf("check table %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("table %s not found after migration", table)
		}
	}
}
