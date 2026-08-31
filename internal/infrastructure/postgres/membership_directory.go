package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/memlore/memlore/internal/application/ports"
	"github.com/memlore/memlore/internal/domain"
	"github.com/memlore/memlore/internal/infrastructure/postgres/sqlc"
)

var _ ports.MembershipDirectory = (*MembershipDirectory)(nil)

// MembershipDirectory implements tenancy persistence via sqlc.
type MembershipDirectory struct {
	q *sqlc.Queries
}

func NewMembershipDirectory(pool *pgxpool.Pool) *MembershipDirectory {
	return &MembershipDirectory{q: sqlc.New(pool)}
}

func NewMembershipDirectoryFromQueries(q *sqlc.Queries) *MembershipDirectory {
	return &MembershipDirectory{q: q}
}

func ts(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func (d *MembershipDirectory) EnsureUser(ctx context.Context, subject string) error {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return &domain.ValidationError{Message: "subject is required"}
	}
	_, err := d.q.EnsureUser(ctx, sqlc.EnsureUserParams{
		ID:        uuid.NewString(),
		Subject:   subject,
		CreatedAt: ts(time.Now()),
	})
	return err
}

func (d *MembershipDirectory) CreateTeam(ctx context.Context, key, name string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return &domain.ValidationError{Message: "team key is required"}
	}
	if strings.TrimSpace(name) == "" {
		name = key
	}
	return d.q.InsertTeam(ctx, sqlc.InsertTeamParams{
		ID:        uuid.NewString(),
		Key:       key,
		Name:      name,
		CreatedAt: ts(time.Now()),
	})
}

func (d *MembershipDirectory) CreateProject(ctx context.Context, key, name, teamKey string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return &domain.ValidationError{Message: "project key is required"}
	}
	if strings.TrimSpace(name) == "" {
		name = key
	}
	var teamID pgtype.Text
	if tk := strings.TrimSpace(teamKey); tk != "" {
		team, err := d.q.GetTeamByKey(ctx, tk)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &domain.NotFoundError{Message: "team not found"}
			}
			return err
		}
		teamID = pgtype.Text{String: team.ID, Valid: true}
	}
	return d.q.InsertProject(ctx, sqlc.InsertProjectParams{
		ID:        uuid.NewString(),
		Key:       key,
		Name:      name,
		TeamID:    teamID,
		CreatedAt: ts(time.Now()),
	})
}

func (d *MembershipDirectory) ensureUserRow(ctx context.Context, subject string) (sqlc.User, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return sqlc.User{}, &domain.ValidationError{Message: "subject is required"}
	}
	return d.q.EnsureUser(ctx, sqlc.EnsureUserParams{
		ID:        uuid.NewString(),
		Subject:   subject,
		CreatedAt: ts(time.Now()),
	})
}

func (d *MembershipDirectory) AddTeamMember(ctx context.Context, teamKey, subject string) error {
	team, err := d.q.GetTeamByKey(ctx, strings.TrimSpace(teamKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.NotFoundError{Message: "team not found"}
		}
		return err
	}
	user, err := d.ensureUserRow(ctx, subject)
	if err != nil {
		return err
	}
	return d.q.InsertMembership(ctx, sqlc.InsertMembershipParams{
		ID:         uuid.NewString(),
		UserID:     user.ID,
		TargetKind: "team",
		TargetID:   team.ID,
		CreatedAt:  ts(time.Now()),
	})
}

func (d *MembershipDirectory) RemoveTeamMember(ctx context.Context, teamKey, subject string) error {
	team, err := d.q.GetTeamByKey(ctx, strings.TrimSpace(teamKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.NotFoundError{Message: "team not found"}
		}
		return err
	}
	user, err := d.q.GetUserBySubject(ctx, strings.TrimSpace(subject))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return d.q.DeleteMembership(ctx, sqlc.DeleteMembershipParams{
		UserID:     user.ID,
		TargetKind: "team",
		TargetID:   team.ID,
	})
}

func (d *MembershipDirectory) AddProjectMember(ctx context.Context, projectKey, subject string) error {
	project, err := d.q.GetProjectByKey(ctx, strings.TrimSpace(projectKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.NotFoundError{Message: "project not found"}
		}
		return err
	}
	user, err := d.ensureUserRow(ctx, subject)
	if err != nil {
		return err
	}
	return d.q.InsertMembership(ctx, sqlc.InsertMembershipParams{
		ID:         uuid.NewString(),
		UserID:     user.ID,
		TargetKind: "project",
		TargetID:   project.ID,
		CreatedAt:  ts(time.Now()),
	})
}

func (d *MembershipDirectory) RemoveProjectMember(ctx context.Context, projectKey, subject string) error {
	project, err := d.q.GetProjectByKey(ctx, strings.TrimSpace(projectKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.NotFoundError{Message: "project not found"}
		}
		return err
	}
	user, err := d.q.GetUserBySubject(ctx, strings.TrimSpace(subject))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	return d.q.DeleteMembership(ctx, sqlc.DeleteMembershipParams{
		UserID:     user.ID,
		TargetKind: "project",
		TargetID:   project.ID,
	})
}

func (d *MembershipDirectory) BindScope(ctx context.Context, scopeKind, scopeKey, projectKey string) error {
	kind := strings.TrimSpace(scopeKind)
	if kind != string(domain.ScopeKindRepository) && kind != string(domain.ScopeKindFeature) && kind != string(domain.ScopeKindTask) {
		return &domain.ValidationError{Message: "invalid scope kind for binding"}
	}
	project, err := d.q.GetProjectByKey(ctx, strings.TrimSpace(projectKey))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.NotFoundError{Message: "project not found"}
		}
		return err
	}
	return d.q.InsertScopeBinding(ctx, sqlc.InsertScopeBindingParams{
		ID:        uuid.NewString(),
		ScopeKind: kind,
		ScopeKey:  strings.TrimSpace(scopeKey),
		ProjectID: project.ID,
		CreatedAt: ts(time.Now()),
	})
}

func (d *MembershipDirectory) UnbindScope(ctx context.Context, scopeKind, scopeKey string) error {
	return d.q.DeleteScopeBinding(ctx, sqlc.DeleteScopeBindingParams{
		ScopeKind: strings.TrimSpace(scopeKind),
		ScopeKey:  strings.TrimSpace(scopeKey),
	})
}

func (d *MembershipDirectory) ResolveScopeAccess(ctx context.Context, subject string, scope domain.Scope) (domain.ResolvedScopeAccess, error) {
	var out domain.ResolvedScopeAccess
	user, err := d.q.GetUserBySubject(ctx, strings.TrimSpace(subject))
	hasUser := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}

	hasMembership := func(kind, targetID string) (bool, error) {
		if !hasUser {
			return false, nil
		}
		return d.q.HasMembership(ctx, sqlc.HasMembershipParams{
			UserID:     user.ID,
			TargetKind: kind,
			TargetID:   targetID,
		})
	}

	fillProjectAccess := func(project sqlc.Project) error {
		var mErr error
		out.ProjectMember, mErr = hasMembership("project", project.ID)
		if mErr != nil {
			return mErr
		}
		if project.TeamID.Valid {
			out.ParentTeamMember, mErr = hasMembership("team", project.TeamID.String)
		}
		return mErr
	}

	switch scope.Kind {
	case domain.ScopeKindTeam, domain.ScopeKindOrganization:
		team, err := d.q.GetTeamByKey(ctx, scope.Key)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return out, nil
			}
			return out, err
		}
		out.TeamMember, err = hasMembership("team", team.ID)
		return out, err

	case domain.ScopeKindProject:
		project, err := d.q.GetProjectByKey(ctx, scope.Key)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return out, nil
			}
			return out, err
		}
		return out, fillProjectAccess(project)

	case domain.ScopeKindRepository, domain.ScopeKindFeature, domain.ScopeKindTask:
		binding, err := d.q.GetScopeBinding(ctx, sqlc.GetScopeBindingParams{
			ScopeKind: string(scope.Kind),
			ScopeKey:  scope.Key,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				out.HasBinding = false
				return out, nil
			}
			return out, err
		}
		out.HasBinding = true
		project, err := d.q.GetProjectByID(ctx, binding.ProjectID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return out, nil
			}
			return out, err
		}
		return out, fillProjectAccess(project)
	default:
		return out, nil
	}
}
