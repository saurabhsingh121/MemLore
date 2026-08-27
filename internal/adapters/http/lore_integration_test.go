//go:build integration

package httpadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	"github.com/memlore/memlore/internal/bootstrap"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestHTTPPostgresCreateVerifyFlow(t *testing.T) {
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

	_, _ = pool.Exec(ctx, `
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
	`)
	_, _ = pool.Exec(ctx, "TRUNCATE audit_records, lore_entries")

	begin := bootstrap.PostgresUnitOfWorkFactory(pool)
	fixed := clock.FixedClock{Instant: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)}
	graph := &memory.KnowledgeGraph{}
	server := httpadapter.NewHandlers(begin, fixed, graph, "integration").Router()

	createBody, _ := json.Marshal(map[string]any{
		"statement": "Outbox required",
		"scope":     map[string]string{"kind": "repository", "key": "github.com/acme/app"},
	})
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", createRec.Code, createRec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	entryID := created["id"].(string)

	verifyRec := httptest.NewRecorder()
	verifyReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+entryID+"/verify", nil)
	verifyReq.Header.Set("X-Memlore-Actor", "alice")
	server.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify status = %d", verifyRec.Code)
	}

	auditsRec := httptest.NewRecorder()
	auditsReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+entryID+"/audits", nil)
	server.ServeHTTP(auditsRec, auditsReq)
	var audits map[string]any
	_ = json.Unmarshal(auditsRec.Body.Bytes(), &audits)
	if len(audits["items"].([]any)) != 2 {
		t.Fatalf("audits = %v", audits)
	}
}
