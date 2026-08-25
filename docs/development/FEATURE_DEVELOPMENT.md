# MemLore Feature Development Tracker

**Last Updated**: 2026-08-25  
**Current Milestone**: M9 — F107 transactional outbox complete; next F108 retrieval  
**Current Release Target**: v0.7.0 governance production-ready; v0.8.0 knowledge plane

---

## Status Legend

| Status | Meaning |
|--------|---------|
| PLANNED | Identified; no spec yet |
| SPECIFYING | Spec Kit specify/clarify in progress |
| READY | Spec + plan + tasks; not coding |
| IN DEVELOPMENT | Active TDD implementation |
| BLOCKED | External dependency or decision needed |
| IN REVIEW | Implementation complete; verification pending |
| DONE | Acceptance criteria met; docs/tests updated |
| DEFERRED | Intentionally postponed |

---

## Feature Summary

| ID | Feature | Status | Spec | Tests | Docs | ADR | Notes |
|----|---------|--------|------|-------|------|-----|-------|
| F001 | Scoped human-authored lore entry (REST) | DONE | ✓ | ✓ | ✓ | 0001 | PostgreSQL governance slice |
| F002 | MCP lore tools (remember/get/verify/explain/search) | DONE | ✓ | ✓ | ✓ | 0003 | Merged to main |
| F003 | Authority factor model + evaluation | PLANNED | — | — | partial | — | Docs only today |
| F004 | Transactional outbox + graph sync | DONE | ✓ | ✓ | ✓ | — | Implemented as F107 |
| F005 | Graph knowledge service (Graphiti isolation) | DONE | ✓ | ✓ | ✓ | — | F106 graph-service |
| F006 | Semantic search + graph retrieval | PLANNED | — | — | — | — | Depends on F005 |
| F007 | Context compiler + `get_for_task` | PLANNED | — | — | partial | — | |
| F008 | Supersession + invalidation | PLANNED | — | — | — | — | |
| F009 | Conflict detection | PLANNED | — | — | — | — | |
| F010 | Auth (OIDC) + team/project scopes | PLANNED | — | — | partial | — | Actor header only today |
| F101 | Go project skeleton + tooling | DONE | ✓ | ✓ | ✓ | 0005 | `go test ./...` green |
| F102 | Go domain primitives (lore/scope/evidence) | DONE | ✓ | ✓ | — | — | Characterization parity with Python |
| F103 | Go PostgreSQL persistence (sqlc/goose) | DONE | ✓ | ✓ | — | — | Repositories + UoW |
| F104 | Migrate lore CRUD/verify REST to Go | DONE | ✓ | ✓ | — | — | `memlore serve` :8080 |
| F105 | Migrate MCP lore tools to Go | DONE | ✓ | ✓ | — | 0003 | `memlore mcp` stdio |
| F106a | Go governance hardening + Python cutover | DONE | ✓ | ✓ | ✓ | — | `memlore migrate`, CI integration |
| F106 | Extract graph-service + contracts | DONE | ✓ | ✓ | ✓ | — | `graph-service/`, Go KnowledgeGraph port |
| F107 | Transactional outbox + graph sync worker | DONE | ✓ | ✓ | ✓ | — | `memlore worker`, outbox migration |

---

## F001 — Scoped Human-Authored Lore Entry (REST)

**Status**: DONE  
**Branch**: `001-scoped-lore-entry` (merged)  
**Commit**: `e342c65`

### Goal

Store, retrieve, verify, and audit scoped human-authored lore on the governance
plane (PostgreSQL) without knowledge-graph coupling.

### Specification

`specs/001-scoped-lore-entry/`

### Acceptance Criteria

- [x] Create with scope (`kind`+`key`), statement, evidence, `human_authored`
- [x] Get by id; not found is clean
- [x] List by exact scope; duplicates allowed
- [x] Verify idempotent; self-verify allowed
- [x] Audit records on create/verify; list audits (404 if entry missing)
- [x] Actor via `X-Memlore-Actor` on mutating REST calls

### TDD Progress

- [x] RED — domain, application, contract tests per `tasks.md`
- [x] GREEN — handlers, repos, REST routes
- [x] REFACTOR — UoW, container wiring

### Implementation

- `src/memlore/domain/models/`
- `src/memlore/application/commands/`, `queries/`
- `src/memlore/infrastructure/postgres/`
- `src/memlore/adapters/rest/`
- `alembic/versions/0001_lore_audit.py`

### Documentation

- `docs/api/rest.md`
- `docs/concepts/lore.md`, `provenance.md`

### Open Questions

- When to extend `VerificationStatus` beyond unverified/verified?

### Next Step

Maintain until Go port (F104) is verified in production; then deprecate Python REST handlers.

---

## F002 — MCP Lore Tools

**Status**: DONE  
**Branch**: merged to `main`  
**Commits**: `144974e`, `d3636cb` (with migration docs)

### Goal

Expose governance-plane lore operations to coding agents via stdio MCP with
domain tool names (no Graphiti leakage).

### Specification

`specs/002-mcp-lore-tools/`

### Acceptance Criteria

- [x] Five tools: remember, get, verify, explain, search
- [x] `actor_id` required on remember/verify
- [x] Payload parity with REST schemas
- [x] `memlore mcp` stdio CLI
- [x] No Graphiti/Neo4j tools on `tools/list`
- [x] Error codes: `validation_error`, `not_found`

### TDD Progress

- [x] RED — unit + contract + e2e tasks T003–T027
- [x] GREEN — `adapters/mcp/`, CLI `mcp` subcommand
- [x] REFACTOR — stderr logging, payload reuse

### Tests Added

- `tests/unit/adapters/test_mcp_*.py`
- `tests/contract/test_mcp_*.py`
- `tests/e2e/test_mcp_stdio.py`
- `tests/support/mcp_client.py`

### Tests Executed

`uv run pytest` → 51 passed (2026-08-25)

### Documentation Updated

- `docs/api/mcp.md`, `README.md`, `docs/architecture/overview.md`, setup guide

### Known Limitations

- Human-authored only on remember (agent origins deferred)
- Search is scope list, not semantic
- No Streamable HTTP MCP
- e2e/integration skip without Postgres

### Next Step

None — complete. Go MCP available via `go run ./cmd/memlore mcp`; Python MCP unchanged until cutover.

---

## F101 — Go Project Skeleton + Tooling

**Status**: DONE  
**Branch**: `003-go-core-skeleton`  
**ADR**: [ADR-0005](../adr/0005-go-memlore-core.md)

### Goal

Additive Go module: `go.mod`, goose/sqlc scaffold, CI gates, layout contract.
No lore handlers; Python REST/MCP unchanged.

### Specification

`specs/003-go-core-skeleton/`

### Acceptance Criteria

- [x] `go test ./...` and `go vet ./...` pass
- [x] CI Go job added
- [x] goose `00001` ports Alembic schema
- [x] Python pytest still green (51 tests)

### TDD Progress

- [x] RED — version, migration parse, layout contract tests
- [x] GREEN — cmd/memlore, migrations, sqlc package, CI job
- [x] REFACTOR — committed sqlc output; integration test behind `integration` tag

### Implementation

- `go.mod`, `cmd/memlore/`, `internal/`, `migrations/`, `db/queries/`
- `internal/infrastructure/postgres/sqlc/` (committed generated code)
- `.github/workflows/ci.yml` — `go-test` job

### Next Step

F103 — Go PostgreSQL persistence (sqlc lore/audit queries)

---

## F102 — Go Domain Primitives

**Status**: DONE  
**Branch**: `004-go-domain-primitives`  
**Spec**: `specs/004-go-domain-primitives/`

### Goal

Port Python governance domain types to Go with characterization test parity.

### Acceptance Criteria

- [x] Enums, scope, evidence, lore, audit, verification in `internal/domain/`
- [x] Validation messages match Python for characterized cases
- [x] Package has no infra imports
- [x] `go test ./internal/domain/...` passes

### Implementation

- `internal/domain/{errors,enums,scope,evidence,lore,audit,verification}.go`
- Characterization tests referencing Python sources

### Next Step

F104 — Migrate lore CRUD/verify REST to Go

---

## F103 — Go PostgreSQL Persistence

**Status**: DONE  
**Branch**: `005-go-postgres-persistence`  
**Spec**: `specs/005-go-postgres-persistence/`

### Goal

sqlc + pgx repositories for lore entries and audit records with transaction UoW.

### Acceptance Criteria

- [x] Application ports in `internal/application/ports/`
- [x] sqlc queries: insert/get/update/list lore; insert/list audit
- [x] Domain ↔ row mapping including JSONB evidence
- [x] `BeginUnitOfWork` with pgx transaction
- [x] Integration tests pass with `-tags=integration` when Postgres is up

### Implementation

- `db/queries/lore.sql`, `audit.sql`
- `internal/infrastructure/postgres/{mapping,lore_repository,audit_repository,unit_of_work}.go`
- sqlc regenerated via `docker run ... sqlc/sqlc:1.28.0 generate`

### Next Step

F104 — application handlers + REST adapter using Go repositories (DONE; see F104/F105)

---

## F003 — Authority Factor Model + Evaluation

**Status**: PLANNED

### Goal

Persist explainable authority factors and compute trust bands for retrieval ranking.

### Specification

Not started — requires `/speckit.specify`

### Acceptance Criteria (draft)

- Verified ADR with evidence → high/canonical trust band
- Unverified agent inference → low trust
- Factors queryable for `explain` flows
- No opaque-only score persistence

### Documentation

- `docs/architecture/authority-model.md` (intent)
- `docs/concepts/authority.md`

### Next Step

Run Spec Kit specify after Go domain skeleton (F102) or in parallel if modeled language-agnostically.

---

## F104 — Migrate Lore CRUD/Verify REST to Go

**Status**: DONE  
**Branch**: `006-go-rest-lore-crud`  
**Spec**: `specs/006-go-rest-lore-crud/`

### Goal

Go REST `/v1/lore-entries` with application handlers and chi adapter.

### Acceptance Criteria

- [x] Create, get, verify, list by scope, list audits
- [x] Error envelope parity (`validation_error`, `not_found`)
- [x] Go contract tests mirror Python contract suite
- [x] `go run ./cmd/memlore serve` (default `:8080`)
- [x] Python `uv run memlore serve` unchanged on `:8000`

### Implementation

- `internal/application/commands/`, `queries/`
- `internal/adapters/http/`
- `internal/infrastructure/memory/` (contract tests)
- `cmd/memlore serve`

### Next Step

Plan Python adapter deprecation and governance hardening.

---

## F105 — Migrate MCP Lore Tools to Go

**Status**: DONE  
**Branch**: `007-go-mcp-lore-tools`  
**Spec**: `specs/007-go-mcp-lore-tools/`

### Goal

Go MCP stdio server with five `memlore.*` lore tools.

### Acceptance Criteria

- [x] `memlore.remember`, `get`, `verify`, `explain`, `search`
- [x] Tool errors: `validation_error: …`, `not_found: …`
- [x] Go contract tests mirror Python MCP contract suite
- [x] `go run ./cmd/memlore mcp` (stdio; logs on stderr)
- [x] Python `uv run memlore mcp` unchanged

### Implementation

- `internal/adapters/mcp/` (official Go MCP SDK)
- `internal/adapters/presenters/` (shared JSON with REST)
- `cmd/memlore mcp`

### Next Step

F106a governance hardening (DONE).

---

## F106a — Go Governance Hardening

**Status**: DONE  
**Branch**: `008-go-governance-hardening`  
**Spec**: `specs/008-go-governance-hardening/`

### Goal

Production-ready Go governance: embedded migrations, CI integration, installable
binary for cross-project MCP, Python deprecation notices.

### Acceptance Criteria

- [x] `memlore migrate` (embedded goose, cwd-independent)
- [x] Alembic integration test on empty DB
- [x] CI `go-integration` job with Postgres
- [x] Python `serve`/`mcp` deprecation stderr notices
- [x] `scripts/install-memlore.sh` + setup docs

### Next Step

F108 — graph retrieval orchestration.

---

## F106 — Graph Knowledge Service

**Status**: DONE  
**Branch**: `009-graph-service`  
**Spec**: `specs/009-graph-service/`

### Goal

Thin Python `graph-service/` isolating Graphiti/Neo4j behind MemLore HTTP contracts.
Go `KnowledgeGraph` port + HTTP client without Graphiti imports.

### Acceptance Criteria

- [x] `docker compose up` brings graph-service + Neo4j; `GET /health` returns ok
- [x] `POST /episodes` via Graphiti; integration test (skip without Neo4j/OpenAI)
- [x] `POST /search` MemLore-shaped results
- [x] Go `KnowledgeGraph` port + HTTP client + contract tests
- [x] CI graph-service job
- [x] Lore create wired via F107 outbox (not synchronous dual-write)

### Implementation

- `graph-service/` FastAPI, Graphiti adapter
- `internal/application/ports/knowledge_graph.go`
- `internal/infrastructure/graphclient/`
- `docs/api/graph-service.md`

---

## F107 — Transactional Outbox + Graph Sync

**Status**: DONE  
**Branch**: `010-transactional-outbox`  
**Spec**: `specs/010-transactional-outbox/`  
**Implements**: Product F004

### Goal

Lore create writes a pending outbox event in the same Postgres transaction.
`memlore worker` polls outbox and publishes episodes to graph-service.

### Acceptance Criteria

- [x] `outbox_events` migration (`00002`)
- [x] `CreateLore` adds `episode.ingest` outbox row atomically with lore + audit
- [x] `memlore worker` calls `KnowledgeGraph.IngestEpisode` with lore id as episode id
- [x] Failed events retry with attempts + last_error
- [x] Unit + Postgres integration tests

### Implementation

- `migrations/00002_outbox_events.sql`
- `internal/application/commands/process_outbox.go`
- `cmd/memlore` `worker` subcommand

---

## Migration Feature Map

| Migration ID | Replaces | Depends on |
|--------------|----------|------------|
| F101 | — | Discovery complete |
| F102 | Python domain types | F101 |
| F103 | SQLAlchemy repos | F101, F102 |
| F104 | F001 REST (Python) | F103 |
| F105 | F002 MCP (Python) | F104 |
| F106a | Ops hardening | F105 |
| F106 | (greenfield) | F004 planning |
| F107 | F004 outbox | F106 |

---

## Development Ledger Notes

### 2026-08-26 — F107 transactional outbox + graph sync

- Outbox migration, CreateLore atomic outbox write, `memlore worker`
- Implements product F004

### 2026-08-25 — F106 graph-service extraction

- `graph-service/` FastAPI + Graphiti adapter (`POST /episodes`, `POST /search`)
- Go `KnowledgeGraph` port + HTTP client + contract tests
- Docker Compose graph-service container; CI graph-service job

### 2026-08-25 — F106a Go governance hardening

- `memlore migrate`, CI Postgres integration tests, Python deprecation notices
- `scripts/install-memlore.sh` for cross-project MCP binary

### 2026-08-25 — F105 Go MCP lore tools

- Branch `007-go-mcp-lore-tools`: five tools via official Go MCP SDK
- `go run ./cmd/memlore mcp` on stdio (Python MCP unchanged)

### 2026-08-25 — F104 Go REST lore CRUD

- Branch `006-go-rest-lore-crud`: chi handlers, application layer, contract tests
- `go run ./cmd/memlore serve` on `:8080` (Python REST unchanged on `:8000`)

### 2026-08-25 — F002 merged; F101 specified

- Merged `002-mcp-lore-tools` to `main` (F002 DONE)
- ADR-0005 accepted; Spec Kit `003-go-core-skeleton` complete (F101 READY)
- Added `docs/architecture/target-architecture.md`

### Immediate recommended tasks

1. `/speckit.specify` F108 (graph retrieval orchestration)
2. Dogfood lore on insurance-core via `./bin/memlore mcp`

---

## Related Documents

- [MIGRATION_DISCOVERY.md](MIGRATION_DISCOVERY.md)
- [migration-inventory.md](migration-inventory.md)
- [../adr/README.md](../adr/README.md)
- [.specify/memory/constitution.md](../../.specify/memory/constitution.md)
