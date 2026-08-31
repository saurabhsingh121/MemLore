package httpadapter

import (
	"context"
	"net/http"
	"strings"

	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/domain"
)

type principalCtxKey struct{}

func withPrincipal(ctx context.Context, p domain.Principal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

func principalFromContext(ctx context.Context) (domain.Principal, bool) {
	p, ok := ctx.Value(principalCtxKey{}).(domain.Principal)
	return p, ok
}

func (h *Handlers) authEnabled() bool {
	return h.Auth != nil && h.Auth.Config.Enabled()
}

func (h *Handlers) oidcMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.authEnabled() {
			next.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			handleDomainError(w, &domain.UnauthorizedError{Message: "bearer token required"})
			return
		}
		raw := strings.TrimSpace(header[len("Bearer "):])
		p, err := h.Auth.AuthenticateBearer(r.Context(), raw)
		if err != nil {
			handleDomainError(w, err)
			return
		}
		if err := appauth.Authorize(p, domain.PermRead); err != nil {
			handleDomainError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
	})
}

// actorFor resolves the acting subject and checks permission.
func (h *Handlers) actorFor(r *http.Request, perm domain.Permission) (string, error) {
	if h.authEnabled() {
		p, ok := principalFromContext(r.Context())
		if !ok {
			return "", &domain.UnauthorizedError{Message: "authentication required"}
		}
		if err := appauth.Authorize(p, perm); err != nil {
			return "", err
		}
		return p.Subject, nil
	}
	actor, err := requireActor(r)
	if err != nil {
		return "", err
	}
	return actor, nil
}

// ensureReadAuthorized checks read permission when OIDC is on (middleware already
// authenticated). No-op in local mode for open reads.
func (h *Handlers) ensureReadAuthorized(r *http.Request) error {
	if !h.authEnabled() {
		return nil
	}
	p, ok := principalFromContext(r.Context())
	if !ok {
		return &domain.UnauthorizedError{Message: "authentication required"}
	}
	return appauth.Authorize(p, domain.PermRead)
}
