package domain

// Role is a coarse authorization role for F111.
type Role string

const (
	RoleReader Role = "reader"
	RoleWriter Role = "writer"
	RoleAdmin  Role = "admin"
)

// Permission is an authorization permission on lore operations.
type Permission string

const (
	PermRead       Permission = "read"
	PermWrite      Permission = "write"
	PermVerify     Permission = "verify"
	PermInvalidate Permission = "invalidate"
)

// Principal is an authenticated caller.
type Principal struct {
	Subject string
	Role    Role
}

// ParseRole parses a role string.
func ParseRole(raw string) (Role, bool) {
	switch Role(raw) {
	case RoleReader, RoleWriter, RoleAdmin:
		return Role(raw), true
	default:
		return "", false
	}
}

// Authorize reports whether role may perform permission.
func Authorize(role Role, perm Permission) error {
	switch perm {
	case PermRead:
		if role == RoleReader || role == RoleWriter || role == RoleAdmin {
			return nil
		}
	case PermWrite:
		if role == RoleWriter || role == RoleAdmin {
			return nil
		}
	case PermVerify, PermInvalidate:
		if role == RoleAdmin {
			return nil
		}
	}
	return &ForbiddenError{Message: "insufficient role for operation"}
}

// HighestRole returns the strongest recognized role from a list.
func HighestRole(roles []string) (Role, bool) {
	var found Role
	rank := map[Role]int{RoleReader: 1, RoleWriter: 2, RoleAdmin: 3}
	best := 0
	for _, raw := range roles {
		if r, ok := ParseRole(raw); ok && rank[r] > best {
			found = r
			best = rank[r]
		}
	}
	return found, best > 0
}
