package domain

import "errors"

// ResolvedScopeAccess is a pure snapshot of membership facts for a lore scope.
// Application/ports load these facts; domain decides allow/deny.
type ResolvedScopeAccess struct {
	TeamMember       bool // team or organization key
	ProjectMember    bool // direct project membership
	ParentTeamMember bool // membership on project's parent team
	HasBinding       bool // child scope has a project binding
}

// AuthorizeScope checks tenancy for a principal against a lore scope.
// RoleAdmin bypasses membership. Caller must still enforce the F111 verb matrix.
func AuthorizeScope(p Principal, scope Scope, access ResolvedScopeAccess) error {
	if p.Role == RoleAdmin {
		return nil
	}
	switch scope.Kind {
	case ScopeKindTeam, ScopeKindOrganization:
		if access.TeamMember {
			return nil
		}
	case ScopeKindProject:
		if access.ProjectMember || access.ParentTeamMember {
			return nil
		}
	case ScopeKindRepository, ScopeKindFeature, ScopeKindTask:
		if !access.HasBinding {
			return &ForbiddenError{Message: "scope is not bound to a project"}
		}
		if access.ProjectMember || access.ParentTeamMember {
			return nil
		}
	default:
		return &ForbiddenError{Message: "unknown scope kind"}
	}
	return &ForbiddenError{Message: "not a member of scope"}
}

// ScopeAccessDeniedAsNotFound maps membership denial to not_found for get-by-id
// leak prevention. Non-forbidden errors pass through unchanged.
func ScopeAccessDeniedAsNotFound(err error) error {
	if err == nil {
		return nil
	}
	var forbidden *ForbiddenError
	if errors.As(err, &forbidden) || errors.Is(err, ErrForbidden) {
		return &NotFoundError{Message: "lore entry not found"}
	}
	return err
}
