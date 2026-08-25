# Research: Scoped Human-Authored Lore Entry

**Feature**: `001-scoped-lore-entry`  
**Date**: 2026-08-25

## R1 — Persistence for this slice

**Decision**: Store lore entries, evidence, and audit records in PostgreSQL only
(governance plane). Do not write to Graphiti/Neo4j in this feature.

**Rationale**: Spec and ADR 0001 place authority/verification/audit in Postgres.
Graph sync via outbox is a later concern; dual-write now would violate “no
distributed transactions” and expand scope.

**Alternatives considered**:
- Dual-write Postgres + Graphiti — rejected (out of scope, complexity).
- In-memory only — rejected (does not satisfy durable provenance).

## R2 — SQLAlchemy / Alembic / driver

**Decision**: SQLAlchemy 2.x style mappings + Alembic migrations; `psycopg`
(v3) sync driver; session-per-request in the REST adapter/bootstrap.

**Rationale**: Matches stack ADR; sync sessions keep the first slice simpler for
TDD and TestClient. Alembic is mandatory per constitution for schema changes.

**Alternatives considered**:
- Async SQLAlchemy + asyncpg — viable later; unnecessary complexity for CRUD MVP.
- Raw SQL only — loses maintainability for evolving governance schema.

## R3 — Identifier scheme

**Decision**: Use UUID version 4 strings (canonical 36-char form) as lore entry
and audit record primary keys, generated in the application/domain layer before
persist.

**Rationale**: Stable, opaque, language-agnostic; no sequential leak. Adequate
for MVP scale.

**Alternatives considered**:
- ULID / UUIDv7 — nicer sortability; defer unless listing-by-time needs it.
- DB serial integers — weaker for public API opacity.

## R4 — Actor identity transport

**Decision**: Require HTTP header `X-Memlore-Actor` (non-empty string) on
mutating endpoints (create, verify). Read endpoints do not require it for this
slice.

**Rationale**: Spec soft-auth; explicit actor without OIDC. Header keeps actor
out of body for GETs and matches audit needs on writes.

**Alternatives considered**:
- Actor in JSON body only — works but inconsistent for verify-with-empty-body.
- Fake auth middleware with defaults — hides missing-actor failures.

## R5 — Evidence and scope storage

**Decision**: Persist scope as columns `scope_kind`, `scope_key`. Persist
evidence as JSONB array of `{type, value}` objects with CHECK/app validation for
allowed types.

**Rationale**: Exact list filter on kind+key is indexed and simple; JSONB avoids
evidence join table for MVP while remaining queryable later.

**Alternatives considered**:
- Normalized `lore_evidence` table — cleaner relationally; add if evidence
  querying becomes first-class.
- Opaque scope string — rejected in clarification (structured kind+key).

## R6 — Verification semantics

**Decision**: `verification_status` enum `unverified` | `verified`. First
successful verify sets `verified_by`, `verified_at`. Subsequent verify is
idempotent no-op (no second audit `verify` row? — see below).

**Rationale**: Spec requires idempotent re-verify preserving original metadata.

**Audit nuance**: Re-verify no-op MUST NOT append another `verify` audit record
(keeps trail meaningful). First verify appends exactly one `verify` audit.

**Alternatives considered**:
- Reject re-verify with 409 — allowed by earlier draft; clarification/assumption
  chose no-op.
- Multi-state `rejected` — deferred.

## R7 — Testing strategy

**Decision**:
- Unit: domain models + application services with in-memory fakes.
- Integration: real Postgres via Docker Compose (or testcontainers) for
  repository + migration tests.
- Contract: HTTP tests against app with test DB (or transactional fixtures)
  covering OpenAPI paths/status codes.

**Rationale**: Constitution test pyramid; Postgres behavior matters for this
feature.

**Alternatives considered**:
- SQLite for tests — rejected (JSONB/enum drift risk vs Postgres).

## R8 — New dependencies

**Decision**: Add `sqlalchemy`, `alembic`, `psycopg[binary]` (or `psycopg`) as
runtime deps; `testcontainers` optional later—prefer compose-backed integration
for v1.

**Rationale**: Justified by governance persistence requirements; no Graphiti yet.

## R9 — Observability for this slice

**Decision**: Structured logging on create/verify/get/list/audit with
`operation`, `lore_entry_id`, `actor_id` (when present). Keep `/health`. Full
OpenTelemetry export can follow without blocking acceptance.

**Rationale**: Constitution VIII — instrument important ops without
over-building.

## R10 — API shape

**Decision**: REST under `/v1/lore-entries` as documented in
`contracts/rest-lore-entries.md`. No MCP in this feature.

**Rationale**: Spec primary interface is human/automation REST; MCP deferred.
