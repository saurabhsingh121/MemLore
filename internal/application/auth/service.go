package auth

import (
	"context"
	"os"
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

// Config holds OIDC settings. Enabled when Issuer is set and a key source exists.
type Config struct {
	Issuer     string
	Audience   string
	JWKSURL    string
	HMACSecret string
	RoleClaim  string
}

// ConfigFromEnv loads auth config from environment.
func ConfigFromEnv() Config {
	roleClaim := strings.TrimSpace(os.Getenv("MEMLORE_OIDC_ROLE_CLAIM"))
	if roleClaim == "" {
		roleClaim = "memlore_role"
	}
	return Config{
		Issuer:     strings.TrimSpace(os.Getenv("MEMLORE_OIDC_ISSUER")),
		Audience:   strings.TrimSpace(os.Getenv("MEMLORE_OIDC_AUDIENCE")),
		JWKSURL:    strings.TrimSpace(os.Getenv("MEMLORE_OIDC_JWKS_URL")),
		HMACSecret: strings.TrimSpace(os.Getenv("MEMLORE_OIDC_HMAC_SECRET")),
		RoleClaim:  roleClaim,
	}
}

// Enabled reports whether OIDC enforcement is active.
func (c Config) Enabled() bool {
	if c.Issuer == "" {
		return false
	}
	return c.JWKSURL != "" || c.HMACSecret != ""
}

// TokenVerifier validates a Bearer access token into a principal.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (domain.Principal, error)
}

// Service resolves principals for HTTP/MCP.
type Service struct {
	Config   Config
	Verifier TokenVerifier
}

// NewService builds an auth service. Verifier may be nil when disabled.
func NewService(cfg Config, verifier TokenVerifier) *Service {
	return &Service{Config: cfg, Verifier: verifier}
}

// AuthenticateBearer validates a bearer token when OIDC is enabled.
func (s *Service) AuthenticateBearer(ctx context.Context, rawToken string) (domain.Principal, error) {
	if s == nil || !s.Config.Enabled() {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "authentication required"}
	}
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "bearer token required"}
	}
	if s.Verifier == nil {
		return domain.Principal{}, &domain.UnauthorizedError{Message: "auth verifier not configured"}
	}
	return s.Verifier.Verify(ctx, token)
}

// LocalPrincipal builds an admin principal from a trusted local actor id.
func LocalPrincipal(actorID string) (domain.Principal, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return domain.Principal{}, &domain.ValidationError{Message: "actor_id is required"}
	}
	return domain.Principal{Subject: actor, Role: domain.RoleAdmin}, nil
}

// Authorize checks permission for a principal.
func Authorize(p domain.Principal, perm domain.Permission) error {
	return domain.Authorize(p.Role, perm)
}
