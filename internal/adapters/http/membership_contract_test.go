package httpadapter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authadapter "github.com/memlore/memlore/internal/adapters/auth"
	httpadapter "github.com/memlore/memlore/internal/adapters/http"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func membershipOIDCServer(t *testing.T, cfg appauth.Config, mem *memory.MembershipDirectory) http.Handler {
	t.Helper()
	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)}
	handlers := httpadapter.NewHandlers(begin, fixed, &memory.KnowledgeGraph{}, "test")
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}
	authSvc := appauth.NewService(cfg, verifier)
	handlers.Auth = authSvc
	handlers.Membership = mem
	handlers.Authz = &authz.Gate{Auth: authSvc, Membership: mem}
	return handlers.Router()
}

func TestMembershipWriterIsolatedByTeam(t *testing.T) {
	cfg := appauth.Config{
		Issuer: "http://localhost/memlore", Audience: "memlore",
		HMACSecret: "test-secret", RoleClaim: "memlore_role",
	}
	mem := memory.NewMembershipDirectory()
	ctx := context.Background()
	_ = mem.CreateTeam(ctx, "alpha", "Alpha")
	_ = mem.CreateTeam(ctx, "beta", "Beta")
	_ = mem.AddTeamMember(ctx, "alpha", "alice")

	server := membershipOIDCServer(t, cfg, mem)
	adminTok, _ := authadapter.IssueHMACToken(cfg, "admin", domain.RoleAdmin, time.Hour)
	aliceTok, _ := authadapter.IssueHMACToken(cfg, "alice", domain.RoleWriter, time.Hour)

	// Admin seeds beta lore
	betaBody, _ := json.Marshal(map[string]any{
		"statement": "Beta secret",
		"scope":     map[string]string{"kind": "team", "key": "beta"},
	})
	betaRec := httptest.NewRecorder()
	betaReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(betaBody))
	betaReq.Header.Set("Content-Type", "application/json")
	betaReq.Header.Set("Authorization", "Bearer "+adminTok)
	server.ServeHTTP(betaRec, betaReq)
	if betaRec.Code != http.StatusCreated {
		t.Fatalf("admin create beta = %d %s", betaRec.Code, betaRec.Body.String())
	}
	var betaEntry map[string]any
	_ = json.Unmarshal(betaRec.Body.Bytes(), &betaEntry)
	betaID := betaEntry["id"].(string)

	// Alice create alpha OK
	alphaBody, _ := json.Marshal(map[string]any{
		"statement": "Alpha rule",
		"scope":     map[string]string{"kind": "team", "key": "alpha"},
	})
	alphaRec := httptest.NewRecorder()
	alphaReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(alphaBody))
	alphaReq.Header.Set("Content-Type", "application/json")
	alphaReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(alphaRec, alphaReq)
	if alphaRec.Code != http.StatusCreated {
		t.Fatalf("alice create alpha = %d %s", alphaRec.Code, alphaRec.Body.String())
	}

	// Alice create beta forbidden
	denyCreate := httptest.NewRecorder()
	denyReq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(betaBody))
	denyReq.Header.Set("Content-Type", "application/json")
	denyReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(denyCreate, denyReq)
	if denyCreate.Code != http.StatusForbidden {
		t.Fatalf("alice create beta = %d %s", denyCreate.Code, denyCreate.Body.String())
	}

	// Alice list beta forbidden
	listRec := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=team&scope_key=beta", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusForbidden {
		t.Fatalf("alice list beta = %d %s", listRec.Code, listRec.Body.String())
	}

	// Alice get beta id → not_found (no leak)
	getRec := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries/"+betaID, nil)
	getReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("alice get beta = %d %s", getRec.Code, getRec.Body.String())
	}

	// Alice list alpha OK
	listA := httptest.NewRecorder()
	listAReq := httptest.NewRequest(http.MethodGet, "/v1/lore-entries?scope_kind=team&scope_key=alpha", nil)
	listAReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(listA, listAReq)
	if listA.Code != http.StatusOK {
		t.Fatalf("alice list alpha = %d %s", listA.Code, listA.Body.String())
	}
}

func TestMembershipAdminBypasses(t *testing.T) {
	cfg := appauth.Config{
		Issuer: "http://localhost/memlore", Audience: "memlore",
		HMACSecret: "test-secret", RoleClaim: "memlore_role",
	}
	mem := memory.NewMembershipDirectory()
	server := membershipOIDCServer(t, cfg, mem)
	adminTok, _ := authadapter.IssueHMACToken(cfg, "ops", domain.RoleAdmin, time.Hour)
	writerTok, _ := authadapter.IssueHMACToken(cfg, "bob", domain.RoleWriter, time.Hour)

	body, _ := json.Marshal(map[string]any{
		"statement": "Anywhere",
		"scope":     map[string]string{"kind": "team", "key": "orphan"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("admin create = %d %s", rec.Code, rec.Body.String())
	}

	deny := httptest.NewRecorder()
	dreq := httptest.NewRequest(http.MethodPost, "/v1/lore-entries", bytes.NewReader(body))
	dreq.Header.Set("Content-Type", "application/json")
	dreq.Header.Set("Authorization", "Bearer "+writerTok)
	server.ServeHTTP(deny, dreq)
	if deny.Code != http.StatusForbidden {
		t.Fatalf("writer without membership = %d", deny.Code)
	}
}

func TestAdminMembershipAPIs(t *testing.T) {
	cfg := appauth.Config{
		Issuer: "http://localhost/memlore", Audience: "memlore",
		HMACSecret: "test-secret", RoleClaim: "memlore_role",
	}
	mem := memory.NewMembershipDirectory()
	server := membershipOIDCServer(t, cfg, mem)
	adminTok, _ := authadapter.IssueHMACToken(cfg, "ops", domain.RoleAdmin, time.Hour)
	writerTok, _ := authadapter.IssueHMACToken(cfg, "bob", domain.RoleWriter, time.Hour)

	teamBody, _ := json.Marshal(map[string]string{"key": "alpha", "name": "Alpha"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/teams", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminTok)
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create team = %d %s", rec.Code, rec.Body.String())
	}

	deny := httptest.NewRecorder()
	dreq := httptest.NewRequest(http.MethodPost, "/v1/admin/teams", bytes.NewReader(teamBody))
	dreq.Header.Set("Content-Type", "application/json")
	dreq.Header.Set("Authorization", "Bearer "+writerTok)
	server.ServeHTTP(deny, dreq)
	if deny.Code != http.StatusForbidden {
		t.Fatalf("writer admin API = %d", deny.Code)
	}

	memBody, _ := json.Marshal(map[string]string{"subject": "alice"})
	mrec := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodPost, "/v1/admin/teams/alpha/members", bytes.NewReader(memBody))
	mreq.Header.Set("Content-Type", "application/json")
	mreq.Header.Set("Authorization", "Bearer "+adminTok)
	server.ServeHTTP(mrec, mreq)
	if mrec.Code != http.StatusCreated {
		t.Fatalf("add member = %d %s", mrec.Code, mrec.Body.String())
	}
}
