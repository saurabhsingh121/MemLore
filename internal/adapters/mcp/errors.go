package mcpadapter

import (
	"context"
	"fmt"
	"os"
	"strings"

	appauth "github.com/memlore/memlore/internal/application/auth"
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

func (t *Tools) resolveActor(ctx context.Context, actorID, accessToken string, perm domain.Permission) (string, error) {
	if t.authEnabled() {
		token := strings.TrimSpace(accessToken)
		if token == "" {
			token = strings.TrimSpace(os.Getenv("MEMLORE_ACCESS_TOKEN"))
		}
		p, err := t.Auth.AuthenticateBearer(ctx, token)
		if err != nil {
			return "", err
		}
		if err := appauth.Authorize(p, perm); err != nil {
			return "", err
		}
		return p.Subject, nil
	}
	return requireActor(actorID)
}

func (t *Tools) resolveRead(ctx context.Context, actorID, accessToken string) error {
	if !t.authEnabled() {
		// Local mode: knowledge_search / get_for_task still required actor_id historically.
		if strings.TrimSpace(actorID) == "" && strings.TrimSpace(accessToken) == "" {
			// get/search/explain don't require actor today; only tools that called requireActor.
			return nil
		}
		_, err := requireActor(actorID)
		return err
	}
	_, err := t.resolveActor(ctx, actorID, accessToken, domain.PermRead)
	return err
}
