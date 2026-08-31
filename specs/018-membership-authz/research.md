# Research: Membership-scoped authorization

## R1 — Enforcement tied to OIDC Enabled

**Decision**: Membership checks run iff `auth.Config.Enabled()` (same gate as
F111 Bearer mode). Local mode: membership off.

**Rationale**: Spec assumption; keeps existing contract suites green without
seed data; one configuration surface.

**Alternatives considered**: Separate `MEMLORE_MEMBERSHIP_ENFORCE=1` flag —
more flexible for OIDC-on dogfood without tenancy, but splits auth modes and
risks misconfiguration. Deferred.

## R2 — Schema: users, teams, projects, memberships, scope_bindings

**Decision**: Five tables in goose `00004_membership.sql`. No `organizations`
table (`organization` ≡ team by key). Boolean membership only. Projects
optional `team_id` FK. Scope bindings unique on `(scope_kind, scope_key)`.

**Rationale**: Clarifications Q1–Q3; smallest isolatable schema.

**Alternatives considered**: Full org chart with repositories table — out of
scope. Key-convention child scopes — rejected in Q2 for fail-closed clarity.

## R3 — Domain AuthorizeScope vs application gate

**Decision**: Pure `domain.AuthorizeScope(principal, scope, view MembershipView)`
(or callback/port interface defined at application boundary and passed as a
narrow domain-facing snapshot). Prefer: application loads a
`MembershipSnapshot` / uses `ports.MembershipDirectory`, then calls domain
with concrete booleans/keys so domain never imports ports/pgx.

**Rationale**: Constitution III — domain MUST NOT import infrastructure.
Membership lookup is I/O → application/port; decision rules are domain.

**Alternatives considered**: All logic in application — weaker reuse/testability
of inheritance rules.

## R4 — Admin bypass

**Decision**: `RoleAdmin` skips membership for lore operations. Admin still
required for admin REST. Reader/writer never bypass.

**Rationale**: Spec User Story 3 / FR-008.

## R5 — Error semantics (404 vs 403)

**Decision**:

| Situation | Code |
|-----------|------|
| Unauthenticated (OIDC on) | `unauthorized` |
| Authenticated, wrong verb, in-scope | `forbidden` |
| Authenticated, names inaccessible scope (list/create) | `forbidden` |
| Authenticated, get-by-id for inaccessible or missing entry | `not_found` |
| Unbound child scope (non-admin) | `forbidden` |

**Rationale**: FR-012 / FR-013; prevent tenant existence leaks.

## R6 — Where to enforce

**Decision**: Shared application helper invoked from HTTP and MCP adapters
after principal resolution and role check, before command/query handlers that
need scope (or inside handlers with principal+scope passed in). Prefer
adapter-level gate for create/list and post-load gate for get-by-id so handlers
stay mostly unchanged — or thin wrapper in application commands/queries.

**Recommended split**:

- **Create / list / compile (scoped) / remember**: check request scope before
  handler body.
- **Get / explain / verify / invalidate / supersede / audits**: load entry,
  then AuthorizeScope on entry.Scope; map deny → NotFound for get/explain/
  audits; map deny → NotFound also for mutate-by-id when out of tenant (same
  leak rule); verb failure after access → Forbidden.
- **Search / knowledge_search / compile multi**: filter entries by
  `CanAccessScope` after retrieval or constrain query to allowed scope set.

**Alternatives considered**: SQL RLS — powerful but opaque to Go tests and
cross-cuts hexagonal style; deferred.

## R7 — Ensure user

**Decision**: Upsert user row by subject on add-member and optionally on first
authenticated request when enforcement is on. No SCIM.

**Rationale**: FR-020.

## R8 — sqlc / goose

**Decision**: Migration `00004_membership.sql`; queries in
`db/queries/membership.sql`; regenerate committed
`internal/infrastructure/postgres/sqlc` via `sqlc generate`.

**Rationale**: ADR 0005 / F103 pattern.

## R9 — MCP tools

**Decision**: No new MCP tools; REST admin only.

**Rationale**: Clarification Q4 / FR-015.

## R10 — Graph-service

**Decision**: Untouched. Authorization happens in Go core before lore write
and before returning reads; outbox inherits authorized writes only.

**Rationale**: Spec FR-018.
