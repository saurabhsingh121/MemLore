package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authadapter "github.com/memlore/memlore/internal/adapters/auth"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func oidcTestClient(t *testing.T, cfg appauth.Config) http.Handler {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)}
	graph := &memory.KnowledgeGraph{}
	handlers := httpadapter.NewHandlers(begin, fixed, graph, "test")
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}
	handlers.Auth = appauth.NewService(cfg, verifier)
	return handlers.Router()
}

func TestOIDCCreateUsesTokenSubjectAndRejectsHeaderOnly(t *testing.T) {
	cfg := appauth.Config{
		Issuer:     "http://localhost/memlore",
		Audience:   "memlore",
		HMACSecret: "test-secret",
		RoleClaim:  "memlore_role",
	}
	server := oidcTestClient(t, cfg)

	missing := httptest.NewRecorder()
	reqMissing := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader([]byte(
		`{"statement":"x","scope":{"kind":"repository","key":"r1"}}`,
	)))
	reqMissing.Header.Set("Content-Type", "application/json")
	reqMissing.Header.Set("X-Memlore-Actor", "spoofed")
	server.ServeHTTP(missing, reqMissing)
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("header-only status = %d body=%s", missing.Code, missing.Body.String())
	}

	token, err := authadapter.IssueHMACToken(cfg, "alice", domain.RoleAdmin, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"statement": "Authenticated rule",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Memlore-Actor", "spoofed")
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["created_by"] != "alice" {
		t.Fatalf("created_by = %v", resp["created_by"])
	}
}

func TestOIDCReaderForbiddenOnVerify(t *testing.T) {
	cfg := appauth.Config{
		Issuer:     "http://localhost/memlore",
		Audience:   "memlore",
		HMACSecret: "test-secret",
		RoleClaim:  "memlore_role",
	}
	server := oidcTestClient(t, cfg)
	adminToken, _ := authadapter.IssueHMACToken(cfg, "admin", domain.RoleAdmin, time.Hour)
	readerToken, _ := authadapter.IssueHMACToken(cfg, "reader", domain.RoleReader, time.Hour)

	createBody, _ := json.Marshal(map[string]any{
		"statement": "Rule",
		"scope":     map[string]string{"kind": "repository", "key": "r1"},
	})
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+adminToken)
	server.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", createRec.Code, createRec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(createRec.Body.Bytes(), &created)
	id := created["id"].(string)

	verifyRec := httptest.NewRecorder()
	verifyReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries/"+id+"/verify", nil)
	verifyReq.Header.Set("Authorization", "Bearer "+readerToken)
	server.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusForbidden {
		t.Fatalf("verify status = %d body=%s", verifyRec.Code, verifyRec.Body.String())
	}
}

func TestHealthOpenWithOIDC(t *testing.T) {
	cfg := appauth.Config{
		Issuer:     "http://localhost/memlore",
		Audience:   "memlore",
		HMACSecret: "test-secret",
		RoleClaim:  "memlore_role",
	}
	server := oidcTestClient(t, cfg)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health = %d", rec.Code)
	}
}
