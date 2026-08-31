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
	"github.com/memlore/memlore/internal/application/commands"
	"github.com/memlore/memlore/internal/application/queries"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
)

func TestIngestPRMembershipDeniesForeignRepository(t *testing.T) {
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
	handlers.IngestPR = commands.NewIngestPullRequestsHandler(begin, fixed, &stubPR{})
	handlers.ListPRIngestRuns = queries.NewListPRIngestRunsHandler(begin)
	server := handlers.Router()

	aliceTok, _ := authadapter.IssueHMACToken(cfg, "alice", domain.RoleWriter, time.Hour)
	body, _ := json.Marshal(map[string]any{
		"scope": map[string]string{"kind": "repository", "key": "github.com/acme/other"},
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("trigger foreign = %d %s", rec.Code, rec.Body.String())
	}

	list := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/v1/ingest/pr-runs?scope_kind=repository&scope_key=github.com/acme/other", nil)
	listReq.Header.Set("Authorization", "Bearer "+aliceTok)
	server.ServeHTTP(list, listReq)
	if list.Code != http.StatusForbidden {
		t.Fatalf("list foreign = %d %s", list.Code, list.Body.String())
	}
}
