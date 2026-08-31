# Implementation Plan: Membership-scoped authorization (F010 remainder)

**Branch**: `018-membership-authz` | **Date**: 2026-08-31 | **Spec**: [spec.md](./spec.md)

## Summary

Add PostgreSQL-backed users, teams, projects, memberships, and child-scope
bindings; enforce **role ∩ membership** on lore REST/MCP paths when OIDC is
configured. JWT `admin` bypasses membership; local mode (OIDC unset) keeps
F111 admin behavior with membership checks off. REST-only admin APIs manage
tenancy; MCP stays at nine tools.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi, pgx/v5, sqlc, goose, existing JWT/auth stack (no new libs)  
**Storage**: PostgreSQL — goose `00004_membership.sql`; sqlc queries under `db/queries/`  
**Testing**: `go test ./...`; domain policy unit tests; OIDC+HMAC membership contract suite; local-mode contracts unchanged  
**Target Platform**: `memlore serve`, `memlore mcp`  
**Project Type**: CLI + REST + MCP governance service  
**Performance Goals**: Membership checks are simple keyed lookups; list/search filter by allowed scopes without full-table scan of foreign tenants  
**Constraints**: TDD; domain pure (no pgx); graph-service untouched; fail closed; no existence leak on cross-tenant get-by-id  
**Scale/Scope**: Tenancy control plane + shared AuthorizeScope policy; not a full org-chart product

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN→REFACTOR for allow/deny, list filter, cross-tenant get, local mode
- [x] Spec-driven: clarifications Q1–Q4 encoded; measurable SC-001–SC-008
- [x] Architecture: domain policy pure; MembershipDirectory port; Postgres in infrastructure; no Graphiti in this feature
- [x] Documentation: FEATURE_DEVELOPMENT, rest.md, security.md, `.env.example`; note on F111 FR-010 completed by 018
- [x] Authority & provenance: unchanged (actor still subject; audits unchanged)
- [x] Temporal correctness: N/A beyond not rewriting lore history
- [x] Secure by default: tenant isolation before retrieval/mutation; admin bypass explicit; secrets not logged
- [x] Observability: forbidden/not_found as structured errors (no token logging)
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: smallest schema (no orgs table; boolean membership; bindings only for child scopes)

**Post-design re-check**: Pass — design uses one MembershipDirectory port, one
AuthorizeScope domain function, shared HTTP/MCP gate; no speculative ABAC.

## Project Structure

### Documentation (this feature)

```text
specs/018-membership-authz/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/membership-authz.md
└── tasks.md                 # /speckit-tasks (not this command)
```

### Source Code (repository root)

```text
migrations/00004_membership.sql
db/queries/membership.sql
internal/domain/membership.go          # AuthorizeScope (+ helpers); pure
internal/domain/membership_test.go
internal/application/ports/membership.go
internal/application/membership/       # admin commands + EnsureUser
internal/application/auth/             # MembershipEnforced() tied to OIDC Enabled
internal/infrastructure/postgres/      # MembershipDirectory impl + sqlc
internal/adapters/http/                # admin routes + lore gate hooks
internal/adapters/mcp/                 # same AuthorizeScope after role check
docs/api/rest.md
docs/architecture/security.md
docs/development/FEATURE_DEVELOPMENT.md
.env.example
```

**Structure Decision**: Extend Go core hexagonal layout. Do not touch
`graph-service/`. Do not revive `src/memlore/`.

## Complexity Tracking

> No constitution violations requiring justification.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/membership-authz.md](./contracts/membership-authz.md)
- [quickstart.md](./quickstart.md)

## Phase 2 — Tasks

See [tasks.md](./tasks.md) (produced by `/speckit-tasks`).

## Implementation approach (planning notes)

### Enforcement gate

`MembershipEnforced() == auth.Config.Enabled()` (OIDC on). Local mode: skip
membership. No separate env flag in v1.

### Domain policy

```text
AuthorizeScope(principal, scope, lookup) error
```

1. If not enforced (caller/adapters skip when local) — N/A at domain; adapters
   only call when enforced OR pass a flag — prefer adapters/application gate:
   `if svc.MembershipEnforced() { domain.AuthorizeScope(...) }`.
2. If `principal.Role == admin` → allow (membership bypass).
3. Resolve effective project/team key(s) from scope kind:
   - `team` / `organization` → team key = scope.key
   - `project` → project key; allow if direct project member OR parent team member
   - `repository` / `feature` / `task` → lookup binding → project; then same as project
4. Unknown / unbound → `ForbiddenError`
5. Role verb check remains separate via existing `Authorize(role, perm)` —
   both must pass. Order: authenticate → role → membership (or role then
   membership; document: role failure → forbidden; membership failure on
   get-by-id after load → not_found if no access).

### Get-by-id leak prevention

Load entry (or check existence internally). If principal cannot access entry
scope → return `NotFoundError` (same as missing). Do not return forbidden for
cross-tenant get. If principal **can** access scope but lacks verb →
`ForbiddenError`.

### List-by-scope

If principal cannot access named scope → `ForbiddenError` (spec FR-013).

### Search / compile / knowledge_search

Filter results to scopes principal may read; never include foreign tenants.
For compile with an explicit task scope the caller cannot access → forbidden.

### Admin REST (admin role only)

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/v1/admin/teams` | Create team |
| `POST` | `/v1/admin/projects` | Create project (optional `team_key`) |
| `POST` | `/v1/admin/teams/{key}/members` | Add team member |
| `DELETE` | `/v1/admin/teams/{key}/members/{subject}` | Remove |
| `POST` | `/v1/admin/projects/{key}/members` | Add project member |
| `DELETE` | `/v1/admin/projects/{key}/members/{subject}` | Remove |
| `POST` | `/v1/admin/scope-bindings` | Bind child scope → project |
| `DELETE` | `/v1/admin/scope-bindings` | Unbind (kind+key query/body) |

Ensure-user on add-member and optionally on first authenticated request.

### Shared wire-up

HTTP `actorFor` / MCP auth path: after role `Authorize`, call membership check
with a shared helper used by both adapters (application-level
`authz.AuthorizeLoreAccess` or extend `auth.Service` with directory port).
