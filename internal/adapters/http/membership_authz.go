package httpadapter

import (
	"net/http"

	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/domain"
)

func (h *Handlers) gate() *authz.Gate {
	if h.Authz != nil {
		return h.Authz
	}
	return &authz.Gate{Auth: h.Auth, Membership: h.Membership}
}

func (h *Handlers) currentPrincipal(r *http.Request) (domain.Principal, bool) {
	if !h.authEnabled() {
		return domain.Principal{}, false
	}
	return principalFromContext(r.Context())
}

// requireScopeAccess enforces membership when OIDC is on. asNotFound maps deny to not_found.
func (h *Handlers) requireScopeAccess(r *http.Request, scope domain.Scope, asNotFound bool) error {
	if !h.gate().Enforced() {
		return nil
	}
	p, ok := h.currentPrincipal(r)
	if !ok {
		return &domain.UnauthorizedError{Message: "authentication required"}
	}
	err := h.gate().AuthorizeLoreScope(r.Context(), p, scope)
	if err != nil && asNotFound {
		return domain.ScopeAccessDeniedAsNotFound(err)
	}
	return err
}

func (h *Handlers) requireAdmin(r *http.Request) error {
	if _, err := h.actorFor(r, domain.PermVerify); err != nil {
		return err
	}
	if h.authEnabled() {
		p, ok := principalFromContext(r.Context())
		if !ok || p.Role != domain.RoleAdmin {
			return &domain.ForbiddenError{Message: "admin role required"}
		}
	}
	return nil
}
