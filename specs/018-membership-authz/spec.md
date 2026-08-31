# Feature Specification: Membership-scoped authorization (F010 remainder)

**Feature Branch**: `018-membership-authz`  
**Created**: 2026-08-31  
**Status**: Implemented  
**Depends on**: F111 / `015-oidc-rbac` (authn + coarse RBAC)  
**Implements**: Product F010 remainder (team/project membership + scope ACL)  
**Input**: User description: "Persist identity org units and memberships;
bind lore scopes to those units; enforce tenant-aware authorization in Go core
so a principal may only read/write lore in scopes they belong to (in addition
to the F111 verb matrix)."

## Goal

Close the tenancy gap left by F111: roles today are global, so a writer who can
name any `scope.kind` + `scope.key` can mutate that lore. Membership-scoped
authorization makes access **tenant-aware** — role answers *what verbs*,
membership answers *which scopes*. Both MUST pass when membership enforcement
is on.

## Clarifications

### Session 2026-08-31

- Q: Project access via parent team? → A: Member of parent team **or**
  direct project member may access that project’s lore (Option A).
- Q: Child scopes (`repository` / `feature` / `task`)? → A: Explicit
  scope-binding table maps each child scope key to a parent project; access
  equals membership on that project (including parent-team inheritance from
  Q1). Unbound child scopes fail closed for non-admin (Option A).
- Q: `organization` scope access? → A: Team-equivalent — `organization`/`K`
  requires membership on team `K` (same rule as `team`/`K`); no separate
  organizations table in v1 (Option A).
- Q: Who manages scope bindings? → A: REST admin CRUD for scope bindings
  (JWT/local `admin` only), part of the membership control plane; MCP unchanged
  (Option A).

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Tenant isolation on lore access (Priority: P1)

An authenticated writer who is a member of team A can create and read lore
scoped to team A. The same writer cannot get, list, search, compile, or mutate
lore scoped to team B. Cross-tenant get-by-id does not reveal that foreign lore
exists.

**Why this priority**: Without membership, F111 RBAC alone cannot prevent
cross-tenant lore access — this is the core product gap.

**Independent Test**: Seed memberships for subject in team A only; OIDC writer
token; allow team A create/get/list; deny team B get/list/create with correct
error semantics (no existence leak on get-by-id).

**Acceptance Scenarios**:

1. **Given** membership enforcement is on and subject `alice` is a member of
   team `alpha` (not `beta`), with role `writer`,
   **When** she creates lore with scope `team`/`alpha`,
   **Then** the create succeeds (role + membership both pass).
2. **Given** the same principal,
   **When** she attempts create or list with scope `team`/`beta`,
   **Then** the operation is forbidden; no lore is written / no `beta` lore
   is returned.
3. **Given** lore exists under `team`/`beta` and `alice` is not a member,
   **When** she gets that entry by id,
   **Then** the response is indistinguishable from a missing entry
   (`not_found`), not `forbidden`.
4. **Given** `alice` is a member of `alpha` but has role `reader`,
   **When** she attempts create on `team`/`alpha`,
   **Then** the response is `forbidden` (verb denied; membership alone is
   insufficient).

---

### User Story 2 — List/search/compile never leak other tenants (Priority: P1)

Search, list-by-scope, compile / get_for_task, explain (where scoped), and
knowledge_search only return lore the principal may read. Results from scopes
outside membership are omitted, not errored as a bulk dump.

**Why this priority**: Read-path leaks are the most common multi-tenant failure
mode; every read must be authorization-aware.

**Independent Test**: Seed lore in team A and team B; principal member of A
only; list/search/compile return only A; empty where no accessible matches.

**Acceptance Scenarios**:

1. **Given** lore in `alpha` and `beta`, principal member of `alpha` only,
   **When** they search or compile without an explicit foreign scope,
   **Then** only `alpha` lore appears in results.
2. **Given** the same principal,
   **When** they list explicitly by `team`/`beta`,
   **Then** they receive forbidden (or empty equivalent that does not confirm
   foreign lore — prefer `forbidden` when the caller names a scope they
   cannot access).
3. **Given** a principal with no memberships and role `writer`,
   **When** they search,
   **Then** results are empty (or forbidden on scoped list); no global lore
   dump.

---

### User Story 3 — Platform admin bypasses membership (Priority: P1)

A principal with global JWT role `admin` may access any scope for lore
operations without being a membership row holder. `reader` and `writer` MUST
have membership for the target scope when enforcement is on.

**Why this priority**: Operators need a break-glass / platform path; encoding
this explicitly prevents silent opposite policy choices.

**Independent Test**: OIDC admin token with no membership rows can get/list
team B lore; OIDC writer with no membership cannot.

**Acceptance Scenarios**:

1. **Given** membership enforcement on and subject with role `admin` and no
   memberships,
   **When** they get or list lore in any team/project scope,
   **Then** membership does not block them (verb matrix still applies —
   admin already has all verbs).
2. **Given** role `writer` with no membership for scope S,
   **When** they attempt any lore operation on S,
   **Then** access is denied per isolation rules.

---

### User Story 4 — Local mode unchanged for dogfood/CI (Priority: P1)

When OIDC is unset (local mode), the trusted actor remains `admin` and
membership checks are **off**, so existing F111 contract tests stay green
without seeding membership rows.

**Why this priority**: Regression of local dogfood would block the whole suite
and daily development.

**Independent Test**: OIDC unset; existing HTTP/MCP lore contract suites pass
with no membership seed data.

**Acceptance Scenarios**:

1. **Given** OIDC unset,
   **When** contract tests use `X-Memlore-Actor` / MCP `actor_id`,
   **Then** behavior matches F111 (admin, no membership required).
2. **Given** OIDC configured with test tokens and seeded memberships,
   **When** the dedicated membership contract suite runs,
   **Then** allow/deny and leak-prevention scenarios pass.

---

### User Story 5 — Operators manage teams, projects, and members (Priority: P2)

An admin provisions teams, projects, scope bindings, and members via REST-only
admin APIs. Agents keep the existing nine MCP tools; membership is not exposed
as MCP tools in v1.

**Why this priority**: Enforcement needs durable membership and binding data;
admin APIs are the control plane for that data.

**Independent Test**: Admin creates team + project, binds a repository scope,
adds member; that member gains access to bound child scope; remove member or
binding → access lost. Non-admin caller cannot manage membership or bindings.

**Acceptance Scenarios**:

1. **Given** a JWT `admin` (or local admin),
   **When** they create a team with key `alpha` and add subject `alice`,
   **Then** `alice` becomes a member of `alpha`.
2. **Given** a JWT `admin`,
   **When** they bind `repository`/`repo-x` to project `p1` and add `alice`
   (directly or via parent team),
   **Then** `alice` may access lore scoped to `repository`/`repo-x` subject to
   her role verbs.
3. **Given** a JWT `writer`,
   **When** they call membership or binding admin endpoints,
   **Then** the call is forbidden.
4. **Given** MCP clients,
   **When** they list available tools,
   **Then** tool count remains nine (no new membership tools).

---

### User Story 6 — Shared policy across REST and MCP (Priority: P2)

REST and MCP apply the **same** principal resolution, role check, and
membership/scope check. A writer denied on REST for team B is also denied on
the equivalent MCP tool.

**Why this priority**: Constitution and F111 require one authorization model
across adapters.

**Independent Test**: Same subject/role/membership fixtures against HTTP and
MCP contract suites yield matching allow/deny outcomes.

**Acceptance Scenarios**:

1. **Given** membership enforcement on and a non-member writer,
   **When** they call REST create and MCP `memlore.remember` for the same
   foreign scope,
   **Then** both deny access.
2. **Given** a member writer for scope S,
   **When** they remember/get via MCP for S,
   **Then** both succeed subject to the verb matrix.

### Edge Cases

- Unknown / unbound scope keys when enforcement is on → fail closed
  (`forbidden`), except local mode (membership off).
- User subject seen for the first time → create-on-first-seen user row is OK;
  no membership until an admin adds them.
- Principal member of a project but not its parent team → still has access to
  that project’s lore (direct membership suffices).
- Principal member of a team but not listed on a child project → still has
  access to that child project’s lore via parent-team inheritance.
- Orphan project (no parent team) → only direct project members (plus admin
  bypass) may access.
- Unbound `repository` / `feature` / `task` scope (no binding to a project) →
  fail closed (`forbidden`) for non-admin when enforcement is on.
- Bound child scope → access follows the bound project’s membership rules
  (direct project member or parent-team member).
- `organization`/`K` → same membership check as `team`/`K` (team-equivalent;
  no separate org table).
- `/health` remains unauthenticated and membership-agnostic.
- Secrets and raw tokens MUST NOT be logged (F111 preserved).
- Removing membership mid-flight → subsequent requests deny; in-flight already
  authorized requests are not retroactively rolled back.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST evaluate authorization as **role AND membership**
  when membership enforcement is on: F111 role grants verbs; membership grants
  scope access. Role alone is insufficient for non-admin principals.
- **FR-002**: System MUST persist users (keyed by subject string), teams,
  projects, and memberships (user↔team and/or user↔project) in the governance
  store. Membership in v1 is boolean (member vs not) — no second role system
  on membership.
- **FR-003**: Team and project stable keys MUST match lore `scope.key` when
  `scope.kind` is `team` or `project`.
- **FR-004**: For `scope.kind=team`, key=T, a non-admin principal MUST be a
  member of team T to access lore in that scope.
- **FR-005**: For `scope.kind=project`, key=P, a non-admin principal MUST be
  either a direct member of project P **or** a member of P’s parent team (when
  the project has a parent team). Orphan projects (no parent team) require
  direct project membership.
- **FR-006**: For `scope.kind` in (`repository`, `feature`, `task`), the system
  MUST resolve access via an explicit binding from that scope key to a parent
  project. Access equals membership on the bound project (including
  parent-team inheritance per FR-005). Unbound child scopes MUST fail closed
  (`forbidden`) for non-admin principals when enforcement is on.
- **FR-006a**: For `scope.kind=organization`, key=K, a non-admin principal
  MUST satisfy the same membership rule as `scope.kind=team`, key=K
  (team-equivalent; no separate organizations entity in v1).
- **FR-007**: When membership enforcement is on, unknown or unbound scopes
  MUST fail closed (`forbidden`) for non-admin principals.
- **FR-008**: Global JWT role `admin` MAY bypass membership checks for lore
  operations (platform operator). Roles `reader` and `writer` MUST NOT bypass
  membership.
- **FR-009**: When OIDC is unset (local mode), membership checks MUST be off
  and the trusted actor remains `admin`, preserving F111 contract behavior
  without membership seed data.
- **FR-010**: When OIDC is configured, membership enforcement MUST be on for
  lore read/write paths (dedicated contract suite with test tokens + seeded
  memberships).
- **FR-011**: Create / remember MUST check membership on the **request** scope.
- **FR-012**: Get / explain / verify / invalidate / supersede MUST check
  membership on the **persisted entry’s** scope. Cross-tenant get-by-id MUST
  return `not_found` (no existence leak). When the caller is a member of the
  scope but lacks the verb, return `forbidden`.
- **FR-013**: List-by-scope / search / compile / get_for_task / knowledge_search
  MUST filter to scopes the principal may read and MUST NOT return other
  tenants’ lore. Listing a scope the caller cannot access MUST return
  `forbidden` (not a silent empty list that could be confused with “no lore”).
- **FR-014**: REST-only admin APIs MUST allow creating teams/projects,
  adding/removing members, and creating/deleting scope bindings
  (`repository` / `feature` / `task` → project); callers MUST have role
  `admin` (JWT or local).
- **FR-015**: MCP tool count MUST remain nine unless this spec is amended;
  membership management is REST-only in v1.
- **FR-016**: REST and MCP MUST share the same principal resolution, role
  checks, and membership/scope policy.
- **FR-017**: `/health` MUST remain unauthenticated.
- **FR-018**: Graph-service auth and Neo4j ACLs are out of scope; core MUST
  authorize before retrieval/mutation and MUST NOT use the graph worker as a
  substitute for ACL.
- **FR-019**: Unauthorized vs forbidden vs not_found MUST remain distinct per
  existing error contracts (REST `{error:{code,message,details}}`; MCP
  `{code}: {message}`).
- **FR-020**: Users MAY be created on first authenticated sighting; SCIM /
  invite flows are out of scope.

### Key Entities

- **User**: Identity row keyed by subject (`OIDC sub` or local actor id).
- **Team**: Named org unit with stable key used as lore `scope.key` for
  `scope.kind=team`.
- **Project**: Named org unit with stable key for `scope.kind=project`; MAY
  belong to a parent team.
- **Membership**: Boolean association of a user to a team or to a project.
- **Scope binding**: Maps a child lore scope (`repository` / `feature` /
  `task` + key) to a parent project for ACL inheritance.
- **Scope access decision**: Outcome of role check + membership/inheritance
  check for a principal and a lore scope (allow, forbidden, or not_found for
  cross-tenant get-by-id).
- **Principal** (from F111): `{ subject, role }` — unchanged structurally;
  membership is looked up by subject when enforcement is on.

## Out of Scope

- SCIM / invite emails / browser login / device-code UI
- Changing the F111 role verb matrix or making OIDC mandatory
- Graph-service OIDC / Neo4j ACLs
- New MCP membership or login tools
- F006 remainder (deeper semantic search)
- Authority factor model (F003) changes
- Nested ABAC / OPA policy language
- Full org-chart product / repository as first-class org table UI
- Reviving deleted Python `src/memlore/` or Alembic

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An OIDC `writer` who is a member of team A and not team B cannot
  get, list, or compile team B lore in automated tests (including no
  existence leak on get-by-id).
- **SC-002**: The same writer can create and search team A lore when the verb
  is allowed, proven by automated tests.
- **SC-003**: An OIDC `admin` with no membership rows can access lore across
  scopes (bypass), proven by automated tests.
- **SC-004**: With OIDC unset, existing F111 HTTP and MCP lore contract suites
  pass without any membership seed data.
- **SC-005**: Dedicated membership enforcement suite (OIDC on + seeded
  memberships) proves allow/deny for create, get, list, and at least one
  search/compile path.
- **SC-006**: Non-admin callers cannot create teams/projects, alter
  memberships, or manage scope bindings; admin callers can — proven by
  automated tests.
- **SC-007**: MCP remains at nine tools; membership admin is REST-only.
- **SC-008**: Full core test suite passes after the feature; graph-service is
  unchanged.

## Assumptions

- F111 optional OIDC + role matrix remains the authentication / verb layer;
  this feature adds tenancy only.
- Projects MAY reference a parent team; team membership inherits access to
  child projects (Clarifications Q1 / Option A).
- Child scopes use an explicit binding table to a project (Clarifications Q2 /
  Option A); v1 will not ship a first-class repository org-table product.
- `organization` scopes are team-equivalent by key (Clarifications Q3 /
  Option A); no organizations table in v1.
- Create-on-first-seen for users is acceptable; no invite workflow in v1.
- Membership enforcement is tied to “OIDC configured” (same gate as F111
  Bearer mode), not a separate env flag, unless Clarifications introduce one.
- Prefer smallest schema that enforces tenant isolation over speculative
  org-chart completeness.
- Documentation updates (feature tracker, REST API, security architecture)
  are part of Done for this feature.
- Brand remains MemLore (ADR 0004); Go core + goose/sqlc governance path
  (ADR 0005).
