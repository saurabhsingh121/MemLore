package mcpadapter

import (
	"fmt"
	"strings"

	"github.com/memlore/memlore/internal/domain"
)

func mapDomainError(err error) error {
	switch e := err.(type) {
	case *domain.ValidationError:
		return fmt.Errorf("validation_error: %s", e.Message)
	case *domain.NotFoundError:
		return fmt.Errorf("not_found: %s", e.Message)
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
