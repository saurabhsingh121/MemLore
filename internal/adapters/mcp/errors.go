package mcpadapter

import (
	"context"
	"fmt"
	"os"
	"strings"

	appauth "github.com/memlore/memlore/internal/application/auth"
	"github.com/memlore/memlore/internal/application/authz"
	"github.com/memlore/memlore/internal/domain"
)

func mapDomainError(err error) error {
	switch e := err.(type) {
	case *domain.ValidationError:
		return fmt.Errorf("validation_error: %s", e.Message)
	case *domain.NotFoundError:
		return fmt.Errorf("not_found: %s", e.Message)
	case *domain.UnauthorizedError:
		return fmt.Errorf("unauthorized: %s", e.Message)
	case *domain.ForbiddenError:
		return fmt.Errorf("forbidden: %s", e.Message)
	default:
		return err
	}
}

func requireActor(actorID string) (string, error) {
	actor := strings.TrimSpace(actorID)
	if actor == "" {
		return "", &domain.ValidationError{Message: "actor must be non-empty"}
	}
	return actor, nil
}

func (t *Tools) authEnabled() bool {
	return t.Auth != nil && t.Auth.Config.Enabled()
}

func (t *Tools) gate() *authz.Gate {
	if t.Authz != nil {
		return t.Authz
	}
	return &authz.Gate{Auth: t.Auth, Membership: t.Membership}
}

func (t *Tools) resolveActor(ctx context.Context, actorID, accessToken string, perm domain.Permission) (string, error) {
	p, err := t.resolvePrincipal(ctx, actorID, accessToken, perm)
	if err != nil {
		return "", err
	}
	return p.Subject, nil
}

func (t *Tools) resolvePrincipal(ctx context.Context, actorID, accessToken string, perm domain.Permission) (domain.Principal, error) {
	if t.authEnabled() {
		token := strings.TrimSpace(accessToken)
		if token == "" {
			token = strings.TrimSpace(os.Getenv("MEMLORE_ACCESS_TOKEN"))
		}
		p, err := t.Auth.AuthenticateBearer(ctx, token)
		if err != nil {
			return domain.Principal{}, err
		}
		if err := appauth.Authorize(p, perm); err != nil {
			return domain.Principal{}, err
		}
		return p, nil
	}
	actor, err := requireActor(actorID)
	if err != nil {
		return domain.Principal{}, err
	}
	return domain.Principal{Subject: actor, Role: domain.RoleAdmin}, nil
}

func (t *Tools) requireScope(ctx context.Context, p domain.Principal, scope domain.Scope, asNotFound bool) error {
	if !t.gate().Enforced() {
		return nil
	}
	err := t.gate().AuthorizeLoreScope(ctx, p, scope)
	if err != nil && asNotFound {
		return domain.ScopeAccessDeniedAsNotFound(err)
	}
	return err
}

func (t *Tools) resolveRead(ctx context.Context, actorID, accessToken string) error {
	if !t.authEnabled() {
		if strings.TrimSpace(actorID) == "" && strings.TrimSpace(accessToken) == "" {
			return nil
		}
		_, err := requireActor(actorID)
		return err
	}
	_, err := t.resolveActor(ctx, actorID, accessToken, domain.PermRead)
	return err
}
