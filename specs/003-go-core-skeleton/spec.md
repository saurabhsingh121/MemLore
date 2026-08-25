# Feature Specification: Go Core Project Skeleton (F101)

**Feature Branch**: `003-go-core-skeleton`  
**Created**: 2026-08-25  
**Status**: Draft  
**Input**: User description: "Introduce MemLore Core Go project skeleton with build tooling, database migration scaffold, and CI gates — no production traffic switch"

**ADR**: [ADR-0005 Go for MemLore Core](../../docs/adr/0005-go-memlore-core.md)

## Clarifications

### Session 2026-08-25

- Q: Does this feature migrate lore REST/MCP to Go? → A: **No** — skeleton and
  tooling only (F104/F105 are separate). Python remains the runtime for agents
  until later slices.
- Q: Must Go schema match existing Postgres tables today? → A: Yes — port
  `lore_entries` and `audit_records` from Alembic `0001` verbatim for sqlc/goose
  readiness; no new columns in F101.
- Q: Is graph-service included? → A: No — `graph-service/` directory placeholder
  is optional; no Graphiti code in F101.
- Q: What proves the skeleton is done? → A: `go test ./...` and `go vet ./...`
  pass locally and in CI; goose migration applies cleanly to empty Postgres;
  Python tests still pass unchanged.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Core toolchain health (Priority: P1)

A contributor clones the repository and can verify MemLore Core module health
with standard Go commands without starting databases or changing Python behavior.

**Why this priority**: Without a buildable Go module, no migration slice can
proceed safely.

**Independent Test**: Run `go test ./...` and `go vet ./...` from repository
root; both succeed with zero failing packages.

**Acceptance Scenarios**:

1. **Given** a clean checkout with Go 1.25+ installed, **When** the contributor
   runs core module tests, **Then** all packages pass with no database required
   for unit-level checks.
2. **Given** the repository CI pipeline on a pull request, **When** Go quality
   gates run, **Then** they pass alongside existing Python ruff/mypy/pytest jobs.
3. **Given** the Python application, **When** F101 merges, **Then** `uv run
   pytest` still passes (no regression to F001/F002).

---

### User Story 2 - Database migration scaffold (Priority: P1)

A contributor can apply MemLore Core database migrations to a fresh PostgreSQL
instance and obtain the same governance tables already used by the Python app.

**Why this priority**: Go persistence slices (F103/F104) depend on goose + sqlc
aligned with the live schema.

**Independent Test**: Start Postgres, run goose up, confirm `lore_entries` and
`audit_records` exist with expected columns and indexes.

**Acceptance Scenarios**:

1. **Given** an empty PostgreSQL database, **When** goose migrations are applied,
   **Then** `lore_entries` and `audit_records` tables exist matching Alembic
   `0001_lore_audit` semantics.
2. **Given** an already-migrated Python dev database, **When** goose `0001` is
   applied to a **fresh** database, **Then** the resulting DDL is compatible
   with what Python expects (no drift in column names/types for existing tables).
3. **Given** goose down on the initial migration, **When** executed on a test DB,
   **Then** tables are removed cleanly (dev/test only).

---

### User Story 3 - Project layout contract (Priority: P2)

A contributor can locate MemLore Core code using a documented, hexagonal directory
layout that mirrors the migration target without empty speculative packages.

**Why this priority**: Establishes boundaries before domain code lands in F102+.

**Independent Test**: Inspect `cmd/`, `internal/`, `db/`, `migrations/` against
`specs/003-go-core-skeleton/contracts/go-module-layout.md`; only directories
with real files exist.

**Acceptance Scenarios**:

1. **Given** the layout contract, **When** a contributor lists the Go tree,
   **Then** `cmd/memlore/main.go` exists as a minimal health/binary entrypoint
   and `internal/` follows domain/application/adapters/infrastructure split.
2. **Given** sqlc configuration, **When** code generation runs, **Then** typed
   query stubs generate without error (queries may be minimal/no-op selects in
   F101).
3. **Given** domain packages, **When** imported from adapters, **Then** domain
   packages do not import HTTP, PostgreSQL, or MCP SDK packages (enforced by
   package structure and review; automated import lint optional in F101).

---

### Edge Cases

- Contributor has only Go 1.24 installed → documented minimum version failure
  with clear message.
- Postgres unavailable for manual migration test → integration test skipped in
  CI with documented opt-in (testcontainers or compose).
- Both Alembic and goose manage the same dev DB → document: use separate DB or
  migrate once; F101 does not auto-run goose against Python-managed DB.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Repository MUST contain a Go module for MemLore Core with
  documented module path and Go 1.25+ minimum.
- **FR-002**: Repository MUST provide `go test ./...` and `go vet ./...` clean
  results for the new module.
- **FR-003**: CI MUST run Go test and vet on pull requests without removing
  Python quality gates.
- **FR-004**: Repository MUST include goose migration(s) reproducing governance
  tables from Alembic `0001_lore_audit`.
- **FR-005**: Repository MUST include sqlc configuration and at least one
  generated query package wired to compile (smoke query acceptable).
- **FR-006**: Repository MUST include a minimal `cmd/memlore` entrypoint
  (health or version command sufficient; HTTP server optional in F101).
- **FR-007**: F101 MUST NOT change Python REST/MCP runtime behavior or default
  developer workflow (`uv run memlore serve` / `memlore mcp` unchanged).
- **FR-008**: Documentation MUST record ADR-0005 and update migration inventory
  and feature tracker for F101 status.

### Key Entities *(include if feature involves data)*

- **Go module**: Buildable unit for MemLore Core; owns `go.mod`, `cmd/`,
  `internal/`.
- **Goose migration**: Versioned SQL applying governance DDL.
- **sqlc package**: Typed database access generated from `db/queries/*.sql`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new contributor completes Go toolchain verification (`go test`,
  `go vet`) in under 5 minutes on a machine with Go installed (excluding module
  download time).
- **SC-002**: CI pipeline includes Go jobs that pass on `main` after merge.
- **SC-003**: Fresh Postgres + goose up produces `lore_entries` and
  `audit_records` matching Python schema expectations (column/table parity).
- **SC-004**: Python test suite remains green (51+ tests) with no behavioral
  changes to F001/F002.

## Assumptions

- Go 1.25+ is available locally and in CI (setup action or preinstalled).
- Module path: `github.com/memlore/memlore` (adjust in plan if repository remote
  differs).
- F101 does not publish a production Go binary; Python remains default.
- OpenTelemetry wiring is deferred to a later observability slice.
- `cmd/worker` may be a stub `main` returning immediately or omitted until F107.

## Out of Scope

- Lore CRUD/handlers in Go (F104)
- MCP server in Go (F105)
- Graphiti / graph-service implementation (F106)
- Transactional outbox tables/worker (F107)
- Deleting or disabling Python `src/memlore/`
