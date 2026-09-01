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

func TestDecisionMembershipDeniesForeignRepository(t *testing.T) {
	cfg := appauth.Config{
		Issuer: "http://localhost/memlore", Audience: "memlore",
		HMACSecret: "test-secret", RoleClaim: "memlore_role",
	}
	mem := memory.NewMembershipDirectory()
	ctx := context.Background()
	_ = mem.CreateTeam(ctx, "alpha", "Alpha")
	_ = mem.CreateProject(ctx, "pay", "Pay", "alpha")
	_ = mem.AddTeamMember(ctx, "alpha", "alice")
	_ = mem.BindScope(ctx, "repository", "github.com/acme/payments", "pay")
	_ = mem.CreateTeam(ctx, "beta", "Beta")
	_ = mem.CreateProject(ctx, "other", "Other", "beta")
	_ = mem.BindScope(ctx, "repository", "github.com/acme/other", "other")

	uow := memory.NewUnitOfWork()
	begin := memory.BeginFactory(uow)
	fixed := clock.FixedClock{Instant: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)}
	handlers := httpadapter.NewHandlers(begin, fixed, &memory.KnowledgeGraph{}, "test")
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}
	authSvc := appauth.NewService(cfg, verifier)
	handlers.Auth = authSvc
	handlers.Membership = mem
	handlers.Authz = &authz.Gate{Auth: authSvc, Membership: mem}
	server := handlers.Router()

	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/other")
	adrEv, _ := domain.NewEvidenceReference(domain.EvidenceTypeADR, "0001-use-postgres")
	adr, _ := domain.NewArchitectureDecisionLoreEntry(domain.NewLoreEntryInput{
		Statement: "Use PostgreSQL.", Scope: scope, CreatedBy: "ingest",
		Evidence: []domain.EvidenceReference{adrEv}, ID: "foreign-adr",
	})
	_ = uow.LoreEntries().Add(ctx, adr)

	aliceTok, _ := authadapter.IssueHMACToken(cfg, "alice", domain.RoleWriter, time.Hour)
	list := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/v1/decisions?scope_kind=repository&scope_key=github.com/acme/other", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(list, listReq)
	if list.Code != http.StatusForbidden {
		t.Fatalf("list foreign = %d %s", list.Code, list.Body.String())
	}

	get := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/v1/decisions/foreign-adr", nil)
	getReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(get, getReq)
	if get.Code != http.StatusNotFound {
		t.Fatalf("get foreign = %d %s", get.Code, get.Body.String())
	}

	body, _ := json.Marshal(map[string]any{
		"scope":    map[string]string{"kind": "repository", "key": "github.com/acme/other"},
		"question": "Q", "decision": "C", "owner": "alice",
	})
	create := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/v1/decisions", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(create, createReq)
	if create.Code != http.StatusForbidden {
		t.Fatalf("create foreign = %d %s", create.Code, create.Body.String())
	}
}
