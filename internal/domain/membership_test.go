package domain_test

import (
	"errors"
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

func TestAuthorizeScopeAdminBypassesMembership(t *testing.T) {
	p := domain.Principal{Subject: "ops", Role: domain.RoleAdmin}
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "beta")
	err := domain.AuthorizeScope(p, scope, domain.ResolvedScopeAccess{})
	if err != nil {
		t.Fatalf("admin should bypass: %v", err)
	}
}

func TestAuthorizeScopeTeamMembership(t *testing.T) {
	writer := domain.Principal{Subject: "alice", Role: domain.RoleWriter}
	scope, _ := domain.NewScope(domain.ScopeKindTeam, "alpha")

	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{TeamMember: true}); err != nil {
		t.Fatalf("member allow: %v", err)
	}
	err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{TeamMember: false})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("non-member = %v", err)
	}
}

func TestAuthorizeScopeOrganizationEqualsTeam(t *testing.T) {
	writer := domain.Principal{Subject: "alice", Role: domain.RoleWriter}
	scope, _ := domain.NewScope(domain.ScopeKindOrganization, "alpha")
	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{TeamMember: true}); err != nil {
		t.Fatalf("org member allow: %v", err)
	}
	err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("org non-member = %v", err)
	}
}

func TestAuthorizeScopeProjectInheritance(t *testing.T) {
	writer := domain.Principal{Subject: "alice", Role: domain.RoleWriter}
	scope, _ := domain.NewScope(domain.ScopeKindProject, "p1")

	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{ProjectMember: true}); err != nil {
		t.Fatalf("direct project member: %v", err)
	}
	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{ParentTeamMember: true}); err != nil {
		t.Fatalf("parent team member: %v", err)
	}
	err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("orphan non-member = %v", err)
	}
}

func TestAuthorizeScopeChildRequiresBinding(t *testing.T) {
	writer := domain.Principal{Subject: "alice", Role: domain.RoleWriter}
	scope, _ := domain.NewScope(domain.ScopeKindRepository, "github.com/acme/app")

	err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{
		HasBinding: false, ProjectMember: true,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("unbound = %v", err)
	}

	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{
		HasBinding: true, ProjectMember: true,
	}); err != nil {
		t.Fatalf("bound project member: %v", err)
	}
	if err := domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{
		HasBinding: true, ParentTeamMember: true,
	}); err != nil {
		t.Fatalf("bound parent team member: %v", err)
	}
	err = domain.AuthorizeScope(writer, scope, domain.ResolvedScopeAccess{HasBinding: true})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("bound but no membership = %v", err)
	}
}

func TestMapScopeAccessDeniedForGet(t *testing.T) {
	err := domain.ScopeAccessDeniedAsNotFound(&domain.ForbiddenError{Message: "no"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}
