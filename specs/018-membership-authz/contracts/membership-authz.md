# Contract: Membership-scoped authorization

Extends [F111 auth-rbac](../../015-oidc-rbac/contracts/auth-rbac.md).

## Modes

| Mode | Role matrix | Membership |
|------|-------------|------------|
| Local (OIDC unset) | actor → admin | **Off** |
| OIDC configured | reader/writer/admin | **On** (admin bypasses) |

## Scope ACL (OIDC on, non-admin)

| scope.kind | Rule |
|------------|------|
| `team` | Member of team with key = scope.key |
| `organization` | Same as `team` (team-equivalent) |
| `project` | Direct project member **or** member of parent team |
| `repository` / `feature` / `task` | Bound to project P; then project rule |
| unbound child scope | `forbidden` |

## Lore operation semantics

| Operation | Membership check | Deny as |
|-----------|------------------|---------|
| create / remember | request scope | `forbidden` |
| list by scope | named scope | `forbidden` |
| get / explain / audits by id | entry scope | `not_found` if no access |
| verify / invalidate / supersede | entry scope | `not_found` if no access; else verb → `forbidden` |
| search / compile / knowledge_search | filter to allowed scopes; explicit inaccessible task scope → `forbidden` | |

Role check still applies (F111 matrix).

## Admin REST (role `admin` only)

All under `/v1/admin/*`; OIDC Bearer or local admin actor.

### Create team

`POST /v1/admin/teams`

```json
{ "key": "alpha", "name": "Alpha" }
```

### Create project

`POST /v1/admin/projects`

```json
{ "key": "p1", "name": "Project One", "team_key": "alpha" }
```

`team_key` optional.

### Add / remove team member

`POST /v1/admin/teams/{key}/members` — `{ "subject": "alice" }`  
`DELETE /v1/admin/teams/{key}/members/{subject}`

### Add / remove project member

`POST /v1/admin/projects/{key}/members` — `{ "subject": "alice" }`  
`DELETE /v1/admin/projects/{key}/members/{subject}`

### Scope bindings

`POST /v1/admin/scope-bindings`

```json
{ "scope_kind": "repository", "scope_key": "github.com/acme/app", "project_key": "p1" }
```

`DELETE /v1/admin/scope-bindings?scope_kind=repository&scope_key=github.com/acme/app`

### Errors

Same envelope as F111:

```json
{ "error": { "code": "unauthorized|forbidden|not_found|validation_error", "message": "...", "details": [] } }
```

Non-admin → `forbidden` on all admin routes.

## MCP

- Tool count remains **9**
- Same membership policy as REST after principal + role resolution
- No membership admin tools

## `/health`

Unauthenticated; no membership check.
