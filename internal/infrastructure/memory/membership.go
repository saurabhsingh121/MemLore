package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
)

var _ ports.MembershipDirectory = (*MembershipDirectory)(nil)

// MembershipDirectory is an in-memory tenancy store for tests.
type MembershipDirectory struct {
	mu       sync.RWMutex
	users    map[string]string            // subject -> id
	teams    map[string]teamRec           // key -> rec
	projects map[string]projectRec        // key -> rec
	members  map[string]map[string]bool   // "kind:targetID" -> subject set? better: userID|kind|targetID
	memberSet map[string]bool             // userID + "|" + kind + "|" + targetID
	bindings map[string]string            // kind+"|"+key -> projectID
}

type teamRec struct {
	ID   string
	Key  string
	Name string
}

type projectRec struct {
	ID     string
	Key    string
	Name   string
	TeamID string
}

func NewMembershipDirectory() *MembershipDirectory {
	return &MembershipDirectory{
		users:     make(map[string]string),
		teams:     make(map[string]teamRec),
		projects:  make(map[string]projectRec),
		memberSet: make(map[string]bool),
		bindings:  make(map[string]string),
	}
}

func (d *MembershipDirectory) EnsureUser(_ context.Context, subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return &domain.ValidationError{Message: "subject is required"}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.users[subject]; !ok {
		d.users[subject] = uuid.NewString()
	}
	return nil
}

func (d *MembershipDirectory) CreateTeam(_ context.Context, key, name string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return &domain.ValidationError{Message: "team key is required"}
	}
	if name == "" {
		name = key
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.teams[key]; ok {
		return &domain.ValidationError{Message: "team already exists"}
	}
	d.teams[key] = teamRec{ID: uuid.NewString(), Key: key, Name: name}
	return nil
}

func (d *MembershipDirectory) CreateProject(_ context.Context, key, name, teamKey string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return &domain.ValidationError{Message: "project key is required"}
	}
	if name == "" {
		name = key
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.projects[key]; ok {
		return &domain.ValidationError{Message: "project already exists"}
	}
	var teamID string
	if tk := strings.TrimSpace(teamKey); tk != "" {
		t, ok := d.teams[tk]
		if !ok {
			return &domain.NotFoundError{Message: "team not found"}
		}
		teamID = t.ID
	}
	d.projects[key] = projectRec{ID: uuid.NewString(), Key: key, Name: name, TeamID: teamID}
	return nil
}

func (d *MembershipDirectory) userIDLocked(subject string) (string, error) {
	id, ok := d.users[subject]
	if !ok {
		id = uuid.NewString()
		d.users[subject] = id
	}
	return id, nil
}

func memberKey(userID, kind, targetID string) string {
	return userID + "|" + kind + "|" + targetID
}

func (d *MembershipDirectory) AddTeamMember(_ context.Context, teamKey, subject string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	t, ok := d.teams[strings.TrimSpace(teamKey)]
	if !ok {
		return &domain.NotFoundError{Message: "team not found"}
	}
	uid, _ := d.userIDLocked(strings.TrimSpace(subject))
	d.memberSet[memberKey(uid, "team", t.ID)] = true
	return nil
}

func (d *MembershipDirectory) RemoveTeamMember(_ context.Context, teamKey, subject string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	t, ok := d.teams[strings.TrimSpace(teamKey)]
	if !ok {
		return &domain.NotFoundError{Message: "team not found"}
	}
	uid, ok := d.users[strings.TrimSpace(subject)]
	if !ok {
		return nil
	}
	delete(d.memberSet, memberKey(uid, "team", t.ID))
	return nil
}

func (d *MembershipDirectory) AddProjectMember(_ context.Context, projectKey, subject string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.projects[strings.TrimSpace(projectKey)]
	if !ok {
		return &domain.NotFoundError{Message: "project not found"}
	}
	uid, _ := d.userIDLocked(strings.TrimSpace(subject))
	d.memberSet[memberKey(uid, "project", p.ID)] = true
	return nil
}

func (d *MembershipDirectory) RemoveProjectMember(_ context.Context, projectKey, subject string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.projects[strings.TrimSpace(projectKey)]
	if !ok {
		return &domain.NotFoundError{Message: "project not found"}
	}
	uid, ok := d.users[strings.TrimSpace(subject)]
	if !ok {
		return nil
	}
	delete(d.memberSet, memberKey(uid, "project", p.ID))
	return nil
}

func (d *MembershipDirectory) BindScope(_ context.Context, scopeKind, scopeKey, projectKey string) error {
	kind := strings.TrimSpace(scopeKind)
	key := strings.TrimSpace(scopeKey)
	if kind != string(domain.ScopeKindRepository) && kind != string(domain.ScopeKindFeature) && kind != string(domain.ScopeKindTask) {
		return &domain.ValidationError{Message: "invalid scope kind for binding"}
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.projects[strings.TrimSpace(projectKey)]
	if !ok {
		return &domain.NotFoundError{Message: "project not found"}
	}
	d.bindings[kind+"|"+key] = p.ID
	return nil
}

func (d *MembershipDirectory) UnbindScope(_ context.Context, scopeKind, scopeKey string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.bindings, strings.TrimSpace(scopeKind)+"|"+strings.TrimSpace(scopeKey))
	return nil
}

func (d *MembershipDirectory) ResolveScopeAccess(_ context.Context, subject string, scope domain.Scope) (domain.ResolvedScopeAccess, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out domain.ResolvedScopeAccess
	uid, ok := d.users[strings.TrimSpace(subject)]
	if !ok {
		// no user → no memberships
		switch scope.Kind {
		case domain.ScopeKindRepository, domain.ScopeKindFeature, domain.ScopeKindTask:
			_, out.HasBinding = d.bindings[string(scope.Kind)+"|"+scope.Key]
		}
		return out, nil
	}

	hasTeam := func(teamKey string) bool {
		t, ok := d.teams[teamKey]
		if !ok {
			return false
		}
		return d.memberSet[memberKey(uid, "team", t.ID)]
	}
	hasProject := func(projectKey string) bool {
		p, ok := d.projects[projectKey]
		if !ok {
			return false
		}
		return d.memberSet[memberKey(uid, "project", p.ID)]
	}
	parentTeamMember := func(projectKey string) bool {
		p, ok := d.projects[projectKey]
		if !ok || p.TeamID == "" {
			return false
		}
		return d.memberSet[memberKey(uid, "team", p.TeamID)]
	}

	switch scope.Kind {
	case domain.ScopeKindTeam, domain.ScopeKindOrganization:
		out.TeamMember = hasTeam(scope.Key)
	case domain.ScopeKindProject:
		out.ProjectMember = hasProject(scope.Key)
		out.ParentTeamMember = parentTeamMember(scope.Key)
	case domain.ScopeKindRepository, domain.ScopeKindFeature, domain.ScopeKindTask:
		pid, ok := d.bindings[string(scope.Kind)+"|"+scope.Key]
		out.HasBinding = ok
		if !ok {
			return out, nil
		}
		for _, p := range d.projects {
			if p.ID == pid {
				out.ProjectMember = d.memberSet[memberKey(uid, "project", p.ID)]
				if p.TeamID != "" {
					out.ParentTeamMember = d.memberSet[memberKey(uid, "team", p.TeamID)]
				}
				break
			}
		}
	}
	return out, nil
}
