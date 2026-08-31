package ports

import (
	"context"

	"github.com/memlore/memlore/internal/domain"
)

// MembershipDirectory persists tenancy and resolves scope access facts.
type MembershipDirectory interface {
	EnsureUser(ctx context.Context, subject string) error
	CreateTeam(ctx context.Context, key, name string) error
	CreateProject(ctx context.Context, key, name, teamKey string) error
	AddTeamMember(ctx context.Context, teamKey, subject string) error
	RemoveTeamMember(ctx context.Context, teamKey, subject string) error
	AddProjectMember(ctx context.Context, projectKey, subject string) error
	RemoveProjectMember(ctx context.Context, projectKey, subject string) error
	BindScope(ctx context.Context, scopeKind, scopeKey, projectKey string) error
	UnbindScope(ctx context.Context, scopeKind, scopeKey string) error
	ResolveScopeAccess(ctx context.Context, subject string, scope domain.Scope) (domain.ResolvedScopeAccess, error)
}
