package mcpadapter_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	authadapter "github.com/memlore/memlore/internal/adapters/auth"
	mcpadapter "github.com/memlore/memlore/internal/adapters/mcp"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/clock"
	"github.com/memlore/memlore/internal/infrastructure/memory"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPMembershipRememberDenyForeignTeam(t *testing.T) {
	cfg := appauth.Config{
		Issuer: "http://localhost/memlore", Audience: "memlore",
		HMACSecret: "test-secret", RoleClaim: "memlore_role",
	}
	mem := memory.NewMembershipDirectory()
	ctx := context.Background()
	_ = mem.CreateTeam(ctx, "alpha", "Alpha")
	_ = mem.AddTeamMember(ctx, "alpha", "alice")

	uow := memory.NewUnitOfWork()
	tools := mcpadapter.NewTools(memory.BeginFactory(uow), clock.FixedClock{Instant: time.Now().UTC()}, &memory.KnowledgeGraph{})
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}
	authSvc := appauth.NewService(cfg, verifier)
	tools.Auth = authSvc
	tools.Membership = mem
	tools.Authz = &authz.Gate{Auth: authSvc, Membership: mem}

	server := mcpadapter.NewServerFromTools(tools, "test", slog.Default())
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	aliceTok, _ := authadapter.IssueHMACToken(cfg, "alice", domain.RoleWriter, time.Hour)

	ok := callTool(t, session, "memlore.remember", map[string]any{
		"statement":    "ok",
		"scope":        map[string]string{"kind": "team", "key": "alpha"},
		"actor_id":     "ignored",
		"access_token": aliceTok,
	})
	if ok.IsError {
		t.Fatalf("remember alpha: %s", toolText(ok))
	}

	deny := callTool(t, session, "memlore.remember", map[string]any{
		"statement":    "nope",
		"scope":        map[string]string{"kind": "team", "key": "beta"},
		"actor_id":     "ignored",
		"access_token": aliceTok,
	})
	if !deny.IsError {
		t.Fatal("expected forbidden for beta")
	}
	if !strings.Contains(toolText(deny), "forbidden") {
		t.Fatalf("deny text = %q", toolText(deny))
	}

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tools) != 9 {
		t.Fatalf("tool count = %d", len(listed.Tools))
	}
}
