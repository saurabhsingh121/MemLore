package domain_test

import (
	"errors"
	"testing"

	"github.com/memlore/memlore/internal/domain"
)

func TestAuthorizeRoleMatrix(t *testing.T) {
	cases := []struct {
		role domain.Role
		perm domain.Permission
		ok   bool
	}{
		{domain.RoleReader, domain.PermRead, true},
		{domain.RoleReader, domain.PermWrite, false},
		{domain.RoleReader, domain.PermVerify, false},
		{domain.RoleWriter, domain.PermRead, true},
		{domain.RoleWriter, domain.PermWrite, true},
		{domain.RoleWriter, domain.PermVerify, false},
		{domain.RoleWriter, domain.PermInvalidate, false},
		{domain.RoleAdmin, domain.PermRead, true},
		{domain.RoleAdmin, domain.PermWrite, true},
		{domain.RoleAdmin, domain.PermVerify, true},
		{domain.RoleAdmin, domain.PermInvalidate, true},
	}
	for _, tc := range cases {
		err := domain.Authorize(tc.role, tc.perm)
		if tc.ok && err != nil {
			t.Fatalf("%s %s: unexpected err %v", tc.role, tc.perm, err)
		}
		if !tc.ok {
			var fe *domain.ForbiddenError
			if !errors.As(err, &fe) {
				t.Fatalf("%s %s: want ForbiddenError, got %v", tc.role, tc.perm, err)
			}
		}
	}
}

func TestHighestRole(t *testing.T) {
	r, ok := domain.HighestRole([]string{"reader", "admin", "nope"})
	if !ok || r != domain.RoleAdmin {
		t.Fatalf("got %q %v", r, ok)
	}
	_, ok = domain.HighestRole([]string{"nope"})
	if ok {
		t.Fatal("expected no role")
	}
}
