# Feature Specification: OIDC authentication + RBAC (F111)

**Feature Branch**: `015-oidc-rbac`  
**Created**: 2026-08-31  
**Status**: Ready  
**Depends on**: F106a (Go governance), F110 (lifecycle), F112 (retrieval)  
**Implements**: Product F010 (partial — authn + coarse RBAC; team/project membership deferred)  
**Input**: User description: "OIDC / RBAC for MemLore REST and MCP — replace
trust-on-header actor model with authenticated identity and authorization."

## Goal

Ensure humans and agents accessing MemLore are **authenticated** (who they are)
and **authorized** (what they may do), so audit provenance reflects real
identity and lore mutations cannot be performed by forging an actor string.

## Clarifications

### Session 2026-08-31

- Q1: Auth enablement? → **A**: Optional until configured — if OIDC issuer/JWKS
  unset, keep `X-Memlore-Actor` / MCP `actor_id`; if configured, Bearer required
  and spoofable actor fields ignored for identity.
- Q2: Role model? → **A**: Three roles from JWT claim: `reader` / `writer` /
  `admin`. Reader = get/list/search/compile/explain; writer = reader +
  create/remember/supersede; admin = writer + verify/invalidate. No team/project
  membership tables in F111 (F010 remainder).
- Q3: MCP auth? → **A**: Same Bearer token when OIDC on; `actor_id` ignored for
  identity when OIDC on; local mode keeps `actor_id`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Authenticated mutations with real actor identity (Priority: P1)

An engineer calls a mutating REST operation with a valid identity token. The
system accepts the request, records the authenticated subject as the actor on
audits, and rejects the same call without valid credentials.

**Why this priority**: Today any client can set `X-Memlore-Actor`; audits and
lifecycle are only as trustworthy as an unverified header.

**Independent Test**: Valid test token → create lore; audit actor matches
subject; without token → unauthorized.

**Acceptance Scenarios**:

1. **Given** OIDC is configured and a valid access token for subject `alice`,
   **When** she creates lore, **Then** the entry and create audit use `alice`
   as actor.
2. **Given** OIDC is configured and no / invalid token, **When** a mutating
   endpoint is called, **Then** the system returns unauthorized and writes no
   lore / audit.
3. **Given** OIDC is configured, **When** a client sends both a valid token and
   a conflicting `X-Memlore-Actor`, **Then** the token subject is the actor
   (header ignored for identity).

---

### User Story 2 — Role-based authorization on lore operations (Priority: P1)

A reader can search and get lore but cannot verify, invalidate, or supersede.
A writer can create and supersede. An admin can verify and invalidate.

**Why this priority**: Authentication alone does not satisfy least privilege.

**Independent Test**: Role matrix unit/contract tests for allow/deny.

**Acceptance Scenarios**:

1. **Given** a principal with role `reader`, **When** they call verify or
   invalidate, **Then** forbidden; no state change.
2. **Given** a principal with role `writer`, **When** they create or supersede,
   **Then** success (subject to other validation); verify/invalidate still
   forbidden.
3. **Given** a principal with role `admin`, **When** they verify or invalidate,
   **Then** success.
4. **Given** a valid token with no recognized role, **When** a protected
   operation is attempted, **Then** forbidden.

---

### User Story 3 — Local / dogfood without a full IdP (Priority: P2)

Developers and CI run without Auth0/Okta while production enforces OIDC.

**Independent Test**: Unconfigured OIDC → existing actor header / `actor_id`
flows pass; configured OIDC → header-only mutate fails.

**Acceptance Scenarios**:

1. **Given** OIDC not configured, **When** contract tests use `X-Memlore-Actor`
   / `actor_id`, **Then** they pass.
2. **Given** OIDC configured, **When** only `X-Memlore-Actor` is sent,
   **Then** mutating calls fail unauthorized.

---

### User Story 4 — MCP shares identity model (Priority: P2)

MCP tools use Bearer when OIDC is on; local mode keeps `actor_id`.

**Acceptance Scenarios**:

1. **Given** OIDC configured, **When** `memlore.remember` is called with a
   valid Bearer identity, **Then** audit actor matches subject.
2. **Given** OIDC configured, **When** credentials are missing, **Then** the
   tool returns an auth error and does not mutate.
3. **Given** OIDC not configured, **When** `actor_id` is supplied as today,
   **Then** behavior is unchanged.

### Edge Cases

- Expired / wrong audience / wrong issuer → unauthorized
- Authenticated but wrong role → forbidden (distinct from unauthorized)
- `/health` always unauthenticated
- Read paths require at least `reader` when OIDC is configured
- Concurrent requests do not leak principal across requests

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When OIDC is configured, system MUST validate Bearer JWT access
  tokens (issuer, audience, signature via JWKS, expiry).
- **FR-002**: When OIDC is configured, system MUST derive `actor_id` from the
  token subject (`sub`) and MUST ignore client-supplied actor header / `actor_id`
  for identity.
- **FR-003**: When OIDC is not configured, system MUST retain current
  `X-Memlore-Actor` / MCP `actor_id` behavior (local/CI mode).
- **FR-004**: System MUST authorize operations by role: `reader`, `writer`,
  `admin` with the matrix in Clarifications Q2.
- **FR-005**: Role MUST be taken from a configurable JWT claim (default
  documented, e.g. `memlore_role` or `roles`); missing/unknown role → forbidden
  when OIDC configured.
- **FR-006**: Unauthorized vs forbidden MUST be distinct API/MCP errors.
- **FR-007**: `/health` MUST remain unauthenticated.
- **FR-008**: REST and MCP MUST share the same principal resolution and
  permission checks.
- **FR-009**: Secrets and raw tokens MUST NOT be logged.
- **FR-010**: Team/project membership persistence and scope ACL completed by
  `018-membership-authz` (F114).

### Key Entities

- **Principal**: `{ subject, role }` from validated token (or local actor string
  in local mode with an implicit elevated role for compatibility — document:
  local mode treats provided actor as `admin` for existing tests, OR maps
  missing role to writer+admin equivalent only in local mode).
- **Role**: `reader` | `writer` | `admin`
- **Auth configuration**: issuer, audience, JWKS URL (or discovery), role claim
  name; “configured” iff issuer+JWKS (or discovery) present

## Local-mode role rule

In local mode (OIDC unset), the caller-supplied actor is trusted and treated as
**admin** so existing dogfood/contract tests keep working without role claims.

## Out of Scope

- Team/project membership tables and scope-bound ACL
- User invite / SCIM
- Graph-service OIDC
- Browser login / device-code UI
- New MCP login tools

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With OIDC configured, mutating calls without Bearer fail closed
  (no state change) in contract tests.
- **SC-002**: Audit `actor_id` equals token `sub` on successful OIDC mutations.
- **SC-003**: Reader cannot verify/invalidate/supersede; writer cannot
  verify/invalidate; admin can — proven by automated tests.
- **SC-004**: Full `go test ./...` passes with OIDC unset (local mode).
- **SC-005**: Enabling OIDC is configuration-only (env/flags), not handler rewrites.

## Assumptions

- Generic OIDC JWT + JWKS (no proprietary IdP SDK required).
- Single role per token for v1 (if claim is an array, first recognized role wins
  with precedence admin > writer > reader).
- Knowledge-search and compile require `reader` when OIDC on.
- F010 remaining work: membership-scoped authorization.
