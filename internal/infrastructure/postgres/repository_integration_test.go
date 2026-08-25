//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres"
)

// Characterization: tests/integration/test_lore_postgres.py

func testDatabaseURL() string {
	if dsn := os.Getenv("MEMLORE_TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	return "postgresql://memlore:memlore@localhost:15432/memlore"
}

func setupPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testDatabaseURL())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		t.Skipf("postgres ping failed: %v", pingErr)
	}
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS lore_entries (
			id VARCHAR(36) PRIMARY KEY,
			statement TEXT NOT NULL,
			scope_kind VARCHAR(64) NOT NULL,
			scope_key VARCHAR(512) NOT NULL,
			origin VARCHAR(64) NOT NULL,
			verification_status VARCHAR(32) NOT NULL,
			evidence JSONB NOT NULL,
			created_by VARCHAR(256) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			verified_by VARCHAR(256),
			verified_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ NOT NULL
		);
		CREATE TABLE IF NOT EXISTS audit_records (
			id VARCHAR(36) PRIMARY KEY,
			target_id VARCHAR(36) NOT NULL,
			action VARCHAR(32) NOT NULL,
			actor_id VARCHAR(256) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		);
	`); err != nil {
		pool.Close()
		t.Fatalf("ensure schema: %v", err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE audit_records, lore_entries"); err != nil {
		pool.Close()
		t.Fatalf("truncate tables: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestPostgresCreateVerifyListAudits(t *testing.T) {
	ctx := context.Background()
	pool := setupPool(t)

	uow, err := postgres.BeginUnitOfWork(ctx, pool)
	if err != nil {
		t.Fatalf("BeginUnitOfWork: %v", err)
	}
	defer uow.Rollback(ctx)

	scope, err := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/app")
	if err != nil {
		t.Fatalf("NewScope: %v", err)
	}
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	entry, err := domain.NewLoreEntry(domain.NewLoreEntryInput{
		Statement: "Outbox required",
		Scope:     scope,
		CreatedBy: "alice",
		Now:       now,
	})
	if err != nil {
		t.Fatalf("NewLoreEntry: %v", err)
	}
	if err := uow.LoreEntries().Add(ctx, entry); err != nil {
		t.Fatalf("Add lore: %v", err)
	}
	createAudit, err := domain.NewAuditRecord(domain.NewAuditRecordInput{
		TargetID:  entry.ID,
		Action:    domain.AuditActionCreate,
		ActorID:   "alice",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("NewAuditRecord: %v", err)
	}
	if err := uow.Audits().Add(ctx, createAudit); err != nil {
		t.Fatalf("Add create audit: %v", err)
	}

	verifyNow := time.Date(2026, 8, 25, 13, 0, 0, 0, time.UTC)
	verified, verifyAudit, err := domain.ApplyVerification(entry, "alice", verifyNow)
	if err != nil {
		t.Fatalf("ApplyVerification: %v", err)
	}
	if err := uow.LoreEntries().Save(ctx, verified); err != nil {
		t.Fatalf("Save lore: %v", err)
	}
	if verifyAudit != nil {
		if err := uow.Audits().Add(ctx, *verifyAudit); err != nil {
			t.Fatalf("Add verify audit: %v", err)
		}
	}
	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	readUow, err := postgres.BeginUnitOfWork(ctx, pool)
	if err != nil {
		t.Fatalf("BeginUnitOfWork read: %v", err)
	}
	defer readUow.Rollback(ctx)

	got, err := readUow.LoreEntries().Get(ctx, entry.ID)
	if err != nil {
		t.Fatalf("Get lore: %v", err)
	}
	if got.VerificationStatus != domain.VerificationVerified {
		t.Fatalf("status = %q", got.VerificationStatus)
	}

	audits, err := readUow.Audits().ListByTarget(ctx, entry.ID)
	if err != nil {
		t.Fatalf("List audits: %v", err)
	}
	if len(audits) != 2 {
		t.Fatalf("audit count = %d", len(audits))
	}
	if audits[0].Action != domain.AuditActionCreate || audits[1].Action != domain.AuditActionVerify {
		t.Fatalf("actions = %v, %v", audits[0].Action, audits[1].Action)
	}

	listed, err := readUow.LoreEntries().ListByScope(ctx, scope)
	if err != nil {
		t.Fatalf("ListByScope: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != entry.ID {
		t.Fatalf("listed = %+v", listed)
	}
}

func TestPostgresGetMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	pool := setupPool(t)

	uow, err := postgres.BeginUnitOfWork(ctx, pool)
	if err != nil {
		t.Fatalf("BeginUnitOfWork: %v", err)
	}
	defer uow.Rollback(ctx)

	_, err = uow.LoreEntries().Get(ctx, "missing-id")
	var nf *domain.NotFoundError
	if err == nil {
		t.Fatal("expected not found")
	}
	if !errors.As(err, &nf) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}
