# MemLore Migration Discovery Report

**Last Updated**: 2026-08-25  
**Branch inspected**: `002-mcp-lore-tools` (uncommitted MCP implementation present)  
**Discovery scope**: Full repository characterization before Go migration or production-behavior changes

---

## Executive Summary

MemLore is an early-stage Python application with a clean hexagonal layout, two
completed Spec Kit features (scoped lore REST slice + MCP lore tools), and
**51 passing tests**. The governance plane (PostgreSQL lore CRUD, verify, audit,
scope list) is **implemented and working**. The knowledge plane (Graphiti,
Neo4j, episodes, semantic retrieval) is **documented and provisioned in Docker
Compose but not implemented in code**. There is **no Go code** in the repository
yet.

The migration target is: **Go MemLore Core** owning governance, orchestration,
MCP/REST, and outbox; **thin Python graph-service** isolating Graphiti. This
discovery report records the baseline so incremental strangler migration can
proceed without guessing.

---

## Current Architecture

```text
                 Humans / Coding Agents
                          |
                +---------+---------+
                |                   |
               MCP                 REST
           (stdio CLI)          (FastAPI)
                |                   |
                +---------+---------+
                          |
              Python MemLore Application
           (domain / application / adapters)
                          |
                          v
                    PostgreSQL
              (lore_entries, audit_records)

     [Planned, not wired]
     Neo4j + Graphiti + Redis + Workers
```

### Layering (implemented)

| Layer | Location | Responsibility |
|-------|----------|----------------|
| Domain | `src/memlore/domain/` | `LoreEntry`, `Scope`, `EvidenceReference`, `AuditRecord`, enums, verification rules |
| Application | `src/memlore/application/` | Commands (`CreateLore`, `VerifyLore`), queries (`GetLore`, `ListLoreByScope`, `ListAudits`), ports |
| Infrastructure | `src/memlore/infrastructure/` | SQLAlchemy Postgres repos, session, UoW, clock, structured logging |
| Adapters | `src/memlore/adapters/` | FastAPI REST, MCP stdio server, CLI (`memlore serve`, `memlore mcp`) |
| Bootstrap | `src/memlore/bootstrap/` | Settings, DI container |

Dependency direction is mostly correct: domain has no FastAPI/SQLAlchemy imports;
adapters depend on application handlers via `AppContainer`.

### Dual-plane intent vs reality

| Plane | Intended store | Current state |
|-------|----------------|---------------|
| Governance | PostgreSQL | **Implemented** — lore entries + audit records |
| Knowledge | Graphiti + Neo4j | **Not implemented** — Neo4j container only; no Graphiti dependency, no ingestion, no search |

Transactional outbox, workers, and Redis are **planned in ADRs/docs** but have
**zero production code**.

---

## Current Tech Stack

| Area | Technology | Notes |
|------|------------|-------|
| Language | Python 3.12 | `requires-python >= 3.12` |
| Package manager | uv | `uv.lock` present |
| HTTP | FastAPI + Uvicorn | REST on port 8000 |
| MCP | `mcp` SDK 2.x | stdio transport via `memlore mcp` |
| ORM / DB | SQLAlchemy 2 + Alembic + psycopg3 | Not sqlc/pgx (target stack differs) |
| Validation | Pydantic v2 | REST/MCP DTOs in adapters |
| Testing | pytest, pytest-asyncio, httpx | 51 tests collected |
| Lint / types | Ruff, mypy `--strict` | CI enforced |
| Containers | Docker Compose | Postgres 16, Neo4j 5, Redis 7 |
| CI | GitHub Actions | ruff, mypy, pytest on push/PR |
| Spec Kit | `.specify/`, `specs/` | Constitution v1.0.0, two feature specs |
| Go | — | **Not present** |

---

## Current Repository Structure

```text
lore/
├── .cursor/              # Spec Kit skills, workspace rules
├── .github/workflows/    # ci.yml
├── .specify/             # Constitution, templates
├── alembic/              # Migration 0001_lore_audit
├── docs/
│   ├── adr/              # ADR-0001..0004
│   ├── api/              # rest.md, mcp.md
│   ├── architecture/     # overview, containers, authority-model, security
│   ├── concepts/         # lore, provenance, authority
│   ├── development/      # setup, testing, tdd, contributing
│   └── operations/       # migrations, observability
├── specs/
│   ├── 001-scoped-lore-entry/   # REST vertical slice (complete)
│   └── 002-mcp-lore-tools/      # MCP tools (complete on branch)
├── src/memlore/          # 39 Python modules
├── tests/
│   ├── unit/             # domain, application, adapters
│   ├── contract/         # REST + MCP in-memory
│   ├── integration/      # Postgres (skips if unavailable)
│   └── e2e/              # stdio MCP (skips if Postgres unavailable)
├── docker-compose.yml
├── pyproject.toml
└── README.md
```

**Not present** (target architecture placeholders): `cmd/`, `internal/`,
`graph-service/`, `db/queries/`, `migrations/` (goose), `go.mod`, OpenAPI
bundle under `api/`.

---

## Implemented Features

| ID | Feature | Status | Evidence |
|----|---------|--------|----------|
| F001 | Scoped human-authored lore entry (REST) | **DONE** | `specs/001-scoped-lore-entry/`, commit `e342c65` |
| F002 | MCP lore tools (remember/get/verify/explain/search) | **DONE** (branch, uncommitted) | `specs/002-mcp-lore-tools/`, 51 tests pass |

### F001 — REST governance slice

- Create lore with scope (`kind` + `key`), statement, evidence, human authorship
- Get by id, list by exact scope, verify (idempotent, self-verify allowed)
- Audit trail on create/verify; list audits (404 if entry missing)
- Actor via `X-Memlore-Actor` header
- Duplicates allowed in same scope

### F002 — MCP parity

- Five tools: `memlore.remember`, `memlore.get`, `memlore.verify`, `memlore.explain`, `memlore.search`
- stdio server: `uv run memlore mcp`
- Reuses application handlers; payloads align with REST schemas
- No Graphiti/Neo4j tools exposed (contract-tested)

### Explicitly not implemented

- Users, teams, projects, repositories as first-class entities
- OIDC / RBAC / authorization beyond actor header
- Authority scoring / explainable ranking (factors documented only)
- Conflicts, supersession, invalidation
- Agent-authored origins (enums reserved, create rejects non-human)
- Semantic / graph search
- `memlore.get_for_task`, context compilation, token budgeting
- Graphiti episode ingestion
- Transactional outbox and background workers
- OpenTelemetry (logging via `slog`-style structured logger only)

---

## Current Domain Model

### Core types (`src/memlore/domain/models/`)

| Type | Fields / behavior |
|------|-------------------|
| `LoreEntry` | id, statement (≤8000), scope, origin, verification_status, evidence[], created_by, timestamps, verified_by/at |
| `Scope` | `ScopeKind` + key (trimmed, ≤512) |
| `EvidenceReference` | `EvidenceType` + value |
| `AuditRecord` | id, target_id, action, actor_id, created_at |
| `HealthStatus` | service health payload |

### Enums (`enums.py`)

- `ScopeKind`: organization, team, project, repository, feature, task
- `EvidenceType`: url, path, adr
- `KnowledgeOrigin`: human_authored, human_verified, agent_*, repository_observation, imported_source, architecture_decision
- `VerificationStatus`: unverified, verified (no canonical/invalidated yet)
- `AuditAction`: create, verify

### Domain services

- `apply_verification()` — idempotent verify, emits audit on first verify
- `LoreEntry.__post_init__` — enforces human_authored on create

### Application handlers

| Handler | Behavior |
|---------|----------|
| `CreateLoreHandler` | Validates actor, creates entry + create audit, commits |
| `VerifyLoreHandler` | Applies verification domain logic, persists |
| `GetLoreHandler` | NotFound if missing |
| `ListLoreByScopeHandler` | Exact scope match, newest first |
| `ListAuditsHandler` | Requires existing entry |

---

## Persistence Model

### Schema (Alembic `0001_lore_audit`)

**`lore_entries`**

| Column | Type |
|--------|------|
| id | varchar(36) PK |
| statement | text |
| scope_kind | varchar(64) |
| scope_key | varchar(512) |
| origin | varchar(64) |
| verification_status | varchar(32) |
| evidence | jsonb |
| created_by | varchar(256) |
| created_at, verified_at, updated_at | timestamptz |

Index: `(scope_kind, scope_key, created_at)`

**`audit_records`**

| Column | Type |
|--------|------|
| id | varchar(36) PK |
| target_id | varchar(36) |
| action | varchar(32) |
| actor_id | varchar(256) |
| created_at | timestamptz |

Indexes: `target_id`, `(target_id, created_at, id)`

### Access pattern

- SQLAlchemy ORM with explicit repository mapping (`SqlAlchemyLoreRepository`, `SqlAlchemyAuditRepository`)
- Unit of work wraps session commit/rollback
- `build_memory_container()` for in-memory contract tests (fake UoW in `tests/support/fakes.py`)

### Environment

- DSN: `MEMLORE_POSTGRES_DSN` (default `localhost:15432`)
- Neo4j/Redis env vars defined in `.env.example` but unused by application code

---

## Graphiti Usage

**Status: not integrated.**

- No `graphiti` or `graphiti-core` dependency in `pyproject.toml`
- No Python modules import Graphiti or Neo4j drivers
- Architecture docs and ADR-0001 describe Graphiti as the knowledge-plane engine
- Docker Compose runs Neo4j 5 with APOC for local future use
- MCP contract explicitly forbids Graphiti/Neo4j tool exposure
- Specs defer graph sync to outbox in later features

**Implication for migration**: Graphiti extraction is greenfield behind a new
Python `graph-service` contract, not a refactor of existing code.

---

## Current API / MCP Interfaces

### REST (`/v1/lore-entries`)

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | — |
| POST | `/v1/lore-entries` | `X-Memlore-Actor` |
| GET | `/v1/lore-entries/{id}` | — |
| POST | `/v1/lore-entries/{id}/verify` | `X-Memlore-Actor` |
| GET | `/v1/lore-entries?scope_kind&scope_key` | — |
| GET | `/v1/lore-entries/{id}/audits` | — |

Contract: `specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`

### MCP (stdio)

| Tool | Mutating | Notes |
|------|----------|-------|
| `memlore.remember` | yes | requires `actor_id` |
| `memlore.get` | no | by id |
| `memlore.verify` | yes | requires `actor_id` |
| `memlore.explain` | no | entry + audit list |
| `memlore.search` | no | exact scope list, not semantic |

Contract: `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md`

Error mapping: `validation_error`, `not_found` as `ToolError` with `{code}: {message}`

---

## Background Processing

**Not implemented.**

- `docs/architecture/containers.md` mentions Dramatiq + Redis ingestion worker (planned)
- Redis container exists in Compose; no consumer code
- No outbox table, no worker binary, no retry/idempotency logic
- Lore writes are synchronous Postgres-only

---

## Existing Tests

| Suite | Count (approx) | Scope | Notes |
|-------|----------------|-------|-------|
| Unit | ~30 | domain, application, MCP adapters | Always run |
| Contract | ~15 | REST + MCP via in-memory container | Always run |
| Integration | 1 file | Postgres round-trip | Skips if DB down |
| E2E | 1 file | stdio MCP full path | Skips if DB down |

**Verified 2026-08-25**: `uv run pytest` → **51 passed** in ~1.8s (integration/e2e
skipped in sandbox without Postgres).

### TDD evidence

Both features have explicit RED→GREEN task lists in `specs/*/tasks.md` with
tests preceding implementation. Constitution mandates TDD for behavioral work.

### Gaps

- No characterization tests exported for Go parity yet
- No contract tests for future graph-service
- No authority evaluator tests (feature not built)
- Integration tests not run in CI against service containers (CI runs pytest only)

---

## Existing Documentation

| Category | Files | Quality |
|----------|-------|---------|
| ADRs | 0001–0004 | Accepted; 0002 (Python stack) will need supersession for Go |
| Architecture | overview, containers, system-context, authority-model, security | Good intent; partially ahead of code |
| Concepts | lore, provenance, authority | Baseline |
| API | rest.md, mcp.md | Aligned with implemented slice |
| Development | setup, testing, tdd, contributing | Present |
| Operations | migrations, observability | Migrations doc exists; observability mostly planned |
| Spec Kit | constitution, two feature specs with plans/tasks/contracts | Strong |

**Missing before this discovery**: `MIGRATION_DISCOVERY.md`, `migration-inventory.md`,
`FEATURE_DEVELOPMENT.md` (now being added).

---

## Spec Kit Artifacts

| Spec | Branch | Status |
|------|--------|--------|
| `001-scoped-lore-entry` | `001-scoped-lore-entry` | Implemented, merged to history |
| `002-mcp-lore-tools` | `002-mcp-lore-tools` | Implemented on branch; polish tasks marked complete |

Constitution: `.specify/memory/constitution.md` v1.0.0 — TDD, spec-driven,
dual-plane, authority/provenance, no distributed transactions.

---

## Technical Debt Observed

1. **ADR-0002 assumes Python as primary stack** — conflicts with Go migration target; needs ADR-0001-style decision for Go core.
2. **SQLAlchemy ORM** — works but diverges from target sqlc + explicit SQL; mapping layer will be rewritten in Go.
3. **Actor identity is a header/argument only** — no real authn/authz.
4. **VerificationStatus incomplete** — missing canonical, invalidated per target domain model.
5. **Authority model documented but not implemented** — no factors persisted, no scoring, no retrieval ranking.
6. **Neo4j/Redis provisioned but unused** — operational noise for new contributors.
7. **Integration/e2e tests skip silently** — easy to miss Postgres-dependent regressions in CI.
8. **Duplicate DTO mapping** — REST routes and MCP tools both map domain → response (acceptable for now).
9. **MCP work uncommitted** on `002-mcp-lore-tools` — branch hygiene before migration kickoff.
10. **No OpenAPI artifact** — REST contract lives in spec markdown only.

---

## Migration Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Behavior drift during Python→Go rewrite | High | Characterization tests from existing handlers; spec contracts as source of truth |
| Premature Graphiti coupling in Go | High | Strict `KnowledgeGraph` port; HTTP contract to Python service only |
| Big-bang rewrite temptation | High | Strangler per vertical slice; keep Python running until Go slice verified |
| ADR/doc stack confusion (Python vs Go) | Medium | New ADRs; mark 0002 superseded when Go skeleton lands |
| Schema migration tooling change (Alembic→goose) | Medium | Port `0001` schema to goose; single source of truth in SQL |
| CI gap for integration tests | Medium | Add testcontainers-go / Compose service in CI for migrated slices |
| Uncommitted MCP feature | Low | Commit/merge 002 before starting F101 |
| Lost audit ordering guarantees | Low | Characterize `ListAudits` ordering before Go port |

---

## Reusable Components

| Asset | Reuse strategy |
|-------|----------------|
| Spec Kit specs (`001`, `002`) | Acceptance criteria for Go ports |
| Domain enums and invariants | Port to Go typed constants; characterization tests |
| REST/MCP contracts | Contract tests against Go adapters |
| Alembic schema `0001` | Basis for goose migration + sqlc queries |
| Hexagonal package layout | Mirror in `internal/domain`, `application`, `adapters` |
| `tests/support/fakes.py` patterns | In-memory ports for Go unit tests |
| ADR-0001 dual-plane, ADR-0003 domain MCP, ADR-0004 brand | Remain valid |
| Docker Compose Postgres | Shared dev/CI dependency |
| Constitution + TDD workflow | Unchanged process |

---

## Components to Replace

| Component | Current | Target |
|-----------|---------|--------|
| Application core runtime | Python FastAPI app | Go `cmd/memlore` + chi |
| Persistence access | SQLAlchemy repositories | sqlc + pgx |
| Migrations | Alembic | goose (SQL remains visible) |
| MCP server | Python MCP SDK | Go MCP SDK |
| REST server | FastAPI | net/http + chi |
| DI / bootstrap | Python `AppContainer` | Explicit Go wiring |
| Graphiti integration | (none) | New Python `graph-service` |

---

## Components to Preserve

| Component | Rationale |
|-----------|-----------|
| PostgreSQL schema semantics | Governance source of truth |
| Spec Kit feature specs | Define WHAT; plans evolve for Go HOW |
| MCP tool names and payloads | Agent contract stability |
| REST paths and JSON shapes | Client compatibility during strangler |
| Domain rules (verify idempotency, duplicates, scope matching) | Product behavior |
| Dual-plane architecture | ADR-0001 still applies |
| Brand (`memlore.*`) | ADR-0004 |

During strangler migration, **keep Python REST/MCP running** until each Go slice
passes contract + characterization tests.

---

## Recommended Migration Sequence

Deviations from the generic phase list are noted where the repository state
warrants them.

### Phase 0 — Repository characterization (this document)

- [x] Discovery report
- [x] Migration inventory
- [x] Feature development tracker
- [ ] Commit/merge `002-mcp-lore-tools`
- [ ] ADR: Go for MemLore Core (proposed ADR-0005)

### Phase 1 — Go project skeleton (F101 prep)

- `go.mod`, `cmd/memlore`, `cmd/worker` stubs
- goose + sqlc scaffold; port `0001` schema
- CI: `go test`, `go vet`, golangci-lint (optional)
- No traffic switch yet

### Phase 2 — Domain primitives in Go (F102)

- Port enums, `Scope`, `EvidenceReference`, `LoreEntry`, validation
- Table-driven unit tests from Python characterization

### Phase 3 — Authority + provenance foundations (F103)

- Extend verification model (canonical, invalidated)
- Persist authority **factors** (not opaque scores)
- Spec Kit spec before implementation

### Phase 4 — PostgreSQL persistence in Go (F104)

- sqlc queries for lore + audit
- Integration tests with testcontainers-go

### Phase 5 — First vertical slice: Lore CRUD in Go (F105)

- Create + get + list + verify + audit list
- REST **or** MCP first (recommend REST for simpler harness)
- Characterization tests comparing Python vs Go outputs
- Strangler proxy or feature flag optional

### Phase 6 — Extract Graphiti Python graph-service (F106)

- FastAPI thin service; MemLore-oriented contracts
- No Graphiti types in Go domain
- Contract tests Go ↔ Python

### Phase 7 — Transactional outbox (F107)

- Outbox table + worker in Go
- Episode publish to graph-service

### Phase 8 — Graph retrieval orchestration (F108)

- `KnowledgeGraph` port implementation
- Parallel retrieval foundation

### Phase 9 — Context compiler + `get_for_task` (F109)

- Scope resolution, authority enrichment, token budget

### Phase 10 — MCP in Go (F110)

- Port five tools; deprecate Python MCP adapter

### Phase 11 — Auth / team scopes (F111)

### Phase 12 — Conflict + supersession (F112)

### Phase 13 — Observability hardening (F113)

**Rationale for deferring Graphiti**: Current code has no Graphiti to extract;
Go governance foundation can stabilize without knowledge plane (per constitution
and spec 001 research).

---

## Proposed First Vertical Slice (Go)

**F105 — Migrate scoped lore create/retrieve/verify to Go (governance only)**

Matches existing F001 behavior:

- Human-authored create with scope, evidence, actor
- Get by id, list by exact scope, verify (idempotent)
- Audit on create/verify; list audits
- PostgreSQL via sqlc; **no Graphiti**
- REST `/v1/lore-entries` as first adapter (MCP follows in F110 or parallel once REST green)

**Acceptance**: Go implementation passes ported contract tests and
characterization fixtures derived from Python `tests/contract/` and
`tests/unit/application/`.

---

## Verification Checklist (Discovery Complete When)

- [x] Repository structure mapped
- [x] Git status and branches reviewed
- [x] README and dependencies read
- [x] Docker Compose inspected
- [x] Migrations inspected
- [x] Tests collected and executed
- [x] API/MCP contracts read
- [x] Graphiti usage confirmed absent
- [x] Domain model documented
- [x] Background jobs confirmed absent
- [x] Spec Kit and ADRs reviewed
- [x] Migration inventory created
- [x] Feature tracker created

---

## Related Documents

- [migration-inventory.md](migration-inventory.md) — capability-level old → new tracking
- [FEATURE_DEVELOPMENT.md](FEATURE_DEVELOPMENT.md) — feature ledger and TDD status
- [../architecture/overview.md](../architecture/overview.md)
- [../adr/README.md](../adr/README.md)
