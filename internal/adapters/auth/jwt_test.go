package authadapter_test

import (
	"context"
	"testing"
	"time"

	authadapter "github.com/memlore/memlore/internal/adapters/auth"
	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/domain"
)

func TestHMACVerifierAcceptsValidToken(t *testing.T) {
	cfg := appauth.Config{
		Issuer:     "http://localhost/memlore",
		Audience:   "memlore",
		HMACSecret: "test-secret",
		RoleClaim:  "memlore_role",
	}
	verifier, err := authadapter.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := authadapter.IssueHMACToken(cfg, "alice", domain.RoleWriter, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	p, err := verifier.Verify(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.Subject != "alice" || p.Role != domain.RoleWriter {
		t.Fatalf("principal = %+v", p)
	}
}

func TestHMACVerifierRejectsBadToken(t *testing.T) {
	cfg := appauth.Config{
		Issuer:     "http://localhost/memlore",
		Audience:   "memlore",
		HMACSecret: "test-secret",
		RoleClaim:  "memlore_role",
	}
	verifier, _ := authadapter.NewVerifier(cfg)
	_, err := verifier.Verify(context.Background(), "not-a-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if _, ok := err.(*domain.UnauthorizedError); !ok {
		t.Fatalf("got %T", err)
	}
}
