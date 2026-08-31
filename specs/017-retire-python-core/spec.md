# Feature Specification: Retire legacy Python governance core

**Feature Branch**: `017-retire-python-core`  
**Created**: 2026-08-31  
**Status**: Implemented  
**Depends on**: F106a (Go production path + deprecation), F104–F112, F111, F003  
**Implements**: F106a remainder — remove Python adapters after Go verified  
**Input**: User description: "Retire legacy core (src/memlore/)."

## Goal

Stop shipping a second, stale governance implementation. Operators and agents
use **Go MemLore Core** only. The Graphiti isolation service remains Python.

## Clarifications

### Session 2026-08-31

Encoded from the retirement analysis. No remaining product questions.

- Q: What is removed? → A: Legacy governance app only: `src/memlore/`, root
  `tests/`, `alembic/`, root `pyproject.toml` / `uv.lock`, `alembic.ini`,
  root `.python-version`. Historical Spec Kit artifacts under `specs/001–016`
  stay (they document how the slice was built).
- Q: What stays Python? → A: `graph-service/` (own `pyproject.toml`, tests,
  CI job, Dockerfile).
- Q: Compatibility shim? → A: None. `uv run memlore` is deleted, not aliased.
- Q: Schema? → A: Goose is the only PostgreSQL migration path. Alembic is
  deleted. DSN in `.env.example` becomes a standard `postgresql://` URL.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Developers run Go core, not Python (Priority: P1)

A developer clones the repo, follows the README, and starts REST and MCP
without installing a root Python package named `memlore`. Graph-service
setup remains `uv` inside `graph-service/`.

**Why this priority**: README still advertised `uv run memlore`; that path
must disappear so nobody starts the five-tool stale server.

**Independent Test**: README/setup mention `./bin/memlore` or
`go run ./cmd/memlore`; `uv run memlore` is gone; graph-service docs still
use `uv` in that directory.

**Acceptance Scenarios**:

1. **Given** a fresh clone, **When** the developer follows quick start,
   **Then** they build/run Go `memlore serve` / `mcp` / `migrate` / `worker`.
2. **Given** the same clone, **When** they look for `uv run memlore`,
   **Then** that command is not documented as a supported entrypoint and
   the package no longer exists at the repo root.
3. **Given** graph-service work, **When** they run tests, **Then**
   `graph-service/` still has its own uv/pytest CI job.

---

### User Story 2 — CI and schema are Go-canonical (Priority: P1)

CI no longer lint/tests the deleted Python core. Schema changes are goose
only. Graph-service quality gates remain.

**Why this priority**: Dual CI implied two products; Alembic could not apply
outbox/lifecycle migrations.

**Independent Test**: `.github/workflows/ci.yml` has go-test, go-integration,
and graph-service; no root `uv run pytest` / ruff on `src tests`.

**Acceptance Scenarios**:

1. **Given** a PR, **When** CI runs, **Then** there is no job that
   `uv sync`s the removed root package.
2. **Given** a schema change, **When** docs are read, **Then** they say
   `memlore migrate` / goose, not Alembic.

---

### User Story 3 — Architecture docs match reality (Priority: P2)

Constitution, ADRs, and architecture overview describe Go core + Python
graph-service. They do not claim Python REST/MCP is current.

**Why this priority**: Stale “today = Python FastAPI” docs would undo the
cutover.

**Independent Test**: Overview “today” is Go; constitution layout is
`cmd/`/`internal/`/`graph-service/`; ADR-0005 notes strangler complete.

**Acceptance Scenarios**:

1. **Given** `docs/architecture/overview.md` and `containers.md`, **When**
   a reader checks current vs target, **Then** MemLore Core is Go, not
   “Python adapters still running.”
2. **Given** the constitution, **When** an agent starts a new feature,
   **Then** preferred layout is Go core; Python is graph-service only.

---

### Edge Cases

- Historical specs (`specs/001-scoped-lore-entry` etc.) may still mention
  `src/memlore/` — leave them; they are the historical record.
- Go characterization comments pointing at `src/memlore/...` may remain as
  provenance of test origin; they are not a runtime dependency.
- `bootstrap` still accepts `postgresql+psycopg://` DSN prefix for old env
  files; new example uses `postgresql://`.
- Do not delete `graph-service/`, Neo4j compose service, or goose
  `migrations/`.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST NOT ship a Python `memlore` CLI for serve/mcp.
- **FR-002**: Repository MUST NOT contain `src/memlore/` or root pytest
  `tests/` for the governance app.
- **FR-003**: `graph-service/` MUST remain with its Python toolchain and CI.
- **FR-004**: Canonical developer commands MUST be Go: `serve`, `mcp`,
  `migrate`, `worker` (via `scripts/install-memlore.sh` or `go run`).
- **FR-005**: CI MUST drop the root Python `quality` job; MUST keep
  `go-test`, `go-integration`, and `graph-service`.
- **FR-006**: PostgreSQL migrations docs MUST name goose / `memlore migrate`
  only.
- **FR-007**: Constitution architecture baseline MUST list Go 1.25+ core and
  Python 3.12 graph-service; workflow MUST require goose (not Alembic) for
  governance schema.
- **FR-008**: README, setup, testing, contributing, and architecture
  “current vs target” MUST not instruct `uv run memlore`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero files remain under `src/memlore/` or root `tests/`
  (excluding `graph-service/tests/`).
- **SC-002**: Root `pyproject.toml` and `alembic.ini` are gone.
- **SC-003**: `go test ./...` still passes.
- **SC-004**: CI workflow has no root-package `uv run pytest`.
- **SC-005**: A new contributor following README never installs FastAPI for
  MemLore Core.

## Assumptions

- Go slices F104–F112, F111, F003, F106a, F107 are the production path.
- No external production traffic still depends on Python `:8000`.
- Historical Spec Kit directories are not rewritten.

## Out of Scope

- Rewriting Graphiti into Go
- Changing goose migrations or Go domain behavior
- F010 team/project membership
- Rewriting completed historical specs
- Removing characterization comments in Go tests
