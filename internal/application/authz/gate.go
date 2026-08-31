package authz

import (
	"context"

	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

// Gate enforces role ∩ membership for lore scopes when OIDC membership is on.
type Gate struct {
	Auth       *appauth.Service
	Membership ports.MembershipDirectory
}

// Enforced reports whether membership checks apply (OIDC configured).
func (g *Gate) Enforced() bool {
	return g != nil && g.Auth != nil && g.Auth.Config.Enabled()
}

// AuthorizeLoreScope checks membership when enforced. Admin bypass is in domain.
func (g *Gate) AuthorizeLoreScope(ctx context.Context, p domain.Principal, scope domain.Scope) error {
	if !g.Enforced() {
		return nil
	}
	if p.Role == domain.RoleAdmin {
		return domain.AuthorizeScope(p, scope, domain.ResolvedScopeAccess{})
	}
	if g.Membership == nil {
		return &domain.ForbiddenError{Message: "not a member of scope"}
	}
	access, err := g.Membership.ResolveScopeAccess(ctx, p.Subject, scope)
	if err != nil {
		return err
	}
	return domain.AuthorizeScope(p, scope, access)
}

// FilterAccessible returns entries whose scopes the principal may read.
func (g *Gate) FilterAccessible(ctx context.Context, p domain.Principal, entries []domain.LoreEntry) ([]domain.LoreEntry, error) {
	if !g.Enforced() || p.Role == domain.RoleAdmin {
		return entries, nil
	}
	out := make([]domain.LoreEntry, 0, len(entries))
	for _, e := range entries {
		if err := g.AuthorizeLoreScope(ctx, p, e.Scope); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}
