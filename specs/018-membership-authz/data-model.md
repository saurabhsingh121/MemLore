# Data Model: Membership-scoped authorization

## Entities

### User

| Field | Type | Notes |
|-------|------|-------|
| id | UUID / string PK | Internal id |
| subject | string UNIQUE | OIDC `sub` or local actor id |
| created_at | timestamptz | |

### Team

| Field | Type | Notes |
|-------|------|-------|
| id | UUID / string PK | |
| key | string UNIQUE | Matches lore `scope.key` for `team` / `organization` |
| name | string | Optional display; may default to key |
| created_at | timestamptz | |

### Project

| Field | Type | Notes |
|-------|------|-------|
| id | UUID / string PK | |
| key | string UNIQUE | Matches lore `scope.key` for `project` |
| name | string | Optional |
| team_id | FK → teams NULL | Parent team for inheritance |
| created_at | timestamptz | |

### Membership

Two shapes (one table with discriminator **or** two tables). Prefer one table:

| Field | Type | Notes |
|-------|------|-------|
| id | UUID / string PK | |
| user_id | FK → users | |
| target_kind | `team` \| `project` | |
| target_id | UUID | FK logically to team or project |
| created_at | timestamptz | |
| UNIQUE | (user_id, target_kind, target_id) | |

### ScopeBinding

| Field | Type | Notes |
|-------|------|-------|
| id | UUID / string PK | |
| scope_kind | `repository` \| `feature` \| `task` | |
| scope_key | string | |
| project_id | FK → projects | |
| created_at | timestamptz | |
| UNIQUE | (scope_kind, scope_key) | |

## Relationships

```text
User ──< Membership >── Team
User ──< Membership >── Project
Team ──< Project (optional parent)
Project ──< ScopeBinding ── (repository|feature|task, key)
```

## Access resolution (logical)

1. **admin** → allow
2. **team / organization** key K → member of team K
3. **project** key P → member of project P OR member of P.team (if set)
4. **repository / feature / task** → binding → project P → rule (3); no binding → deny
5. Else → deny

## Validation

- Team/project keys: non-empty, ≤ MaxScopeKeyLength, stable slug
- Scope binding kinds: only repository, feature, task
- Membership target must exist
- Delete team/project: restrict or cascade — **v1: RESTRICT** if memberships/
  bindings/projects reference (admin must clean up); document in migration

## State transitions

- Membership: add / remove only (no roles)
- Binding: create / delete only
- Users: create-on-first-seen / ensure; no delete in v1 API

## Persistence notes

- Goose `00004_membership.sql`
- sqlc in `db/queries/membership.sql`
- Lore tables unchanged; scopes remain free strings on lore rows — ACL is
  evaluated at read/write time against membership data
