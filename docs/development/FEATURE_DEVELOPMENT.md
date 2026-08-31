# MemLore Feature Development Tracker

**Last Updated**: 2026-08-31  
**Current Milestone**: M17 — membership-scoped authorization  
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
| F001 | Scoped human-authored lore entry (REST) | DONE | ✓ | ✓ | ✓ | 0001 | Go REST (Python core removed) |
| F002 | MCP lore tools (remember/get/verify/explain/search) | DONE | ✓ | ✓ | ✓ | 0003 | Merged to main |
| F003 | Authority factor model + evaluation | DONE | ✓ | ✓ | ✓ | — | Ephemeral factors + trust bands on compile/explain |
| F113 | Retire legacy Python core | DONE | ✓ | ✓ | ✓ | 0005 | Removed `src/memlore/`; graph-service kept |
| F004 | Transactional outbox + graph sync | DONE | ✓ | ✓ | ✓ | — | Implemented as F107 |
| F005 | Graph knowledge service (Graphiti isolation) | DONE | ✓ | ✓ | ✓ | — | F106 graph-service |
| F006 | Semantic search + graph retrieval | PARTIAL | ✓ | ✓ | — | — | F108 read path; full F006 deferred |
| F007 | Context compiler + `get_for_task` | DONE | ✓ | ✓ | ✓ | — | F109 + F112 (temporal filter + conflicts) |
| F008 | Supersession + invalidation | DONE | ✓ | ✓ | ✓ | 0003 | Implemented as F110 |
| F009 | Conflict detection | DONE | ✓ | ✓ | ✓ | — | Implemented as F112 |
| F010 | Auth (OIDC) + team/project scopes | DONE | ✓ | ✓ | ✓ | — | F111 authn+RBAC; F114 membership ACL |
| F114 | Membership-scoped authorization | DONE | ✓ | ✓ | ✓ | — | `018-membership-authz`; goose 00004 |
| F101 | Go project skeleton + tooling | DONE | ✓ | ✓ | ✓ | 0005 | `go test ./...` green |
| F102 | Go domain primitives (lore/scope/evidence) | DONE | ✓ | ✓ | — | — | Characterization parity with Python |
| F103 | Go PostgreSQL persistence (sqlc/goose) | DONE | ✓ | ✓ | — | — | Repositories + UoW |
| F104 | Migrate lore CRUD/verify REST to Go | DONE | ✓ | ✓ | — | — | `memlore serve` :8080 |
| F105 | Migrate MCP lore tools to Go | DONE | ✓ | ✓ | — | 0003 | `memlore mcp` stdio |
| F106a | Go governance hardening + Python cutover | DONE | ✓ | ✓ | ✓ | — | `memlore migrate`, CI; Python core later removed |
| F106 | Extract graph-service + contracts | DONE | ✓ | ✓ | ✓ | — | `graph-service/`, Go KnowledgeGraph port |
| F107 | Transactional outbox + graph sync worker | DONE | ✓ | ✓ | ✓ | — | `memlore worker`, outbox migration |
| F110 | Invalidate + supersede lifecycle | DONE | ✓ | ✓ | ✓ | 0003 | REST + MCP |
| F112 | Temporal filter + conflict detection | DONE | ✓ | ✓ | ✓ | — | Filter stale from retrieval; surface conflicts |
| F111 | OIDC + RBAC | DONE | ✓ | ✓ | ✓ | — | Optional OIDC; reader/writer/admin; membership deferred |

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

Historical Python tree (removed in F113):

- `src/memlore/domain/models/`
- `src/memlore/application/commands/`, `queries/`
- `src/memlore/infrastructure/postgres/`
- `src/memlore/adapters/rest/`
- `alembic/versions/0001_lore_audit.py`

Canonical implementation is Go REST (F104).

### Documentation

- `docs/api/rest.md`
- `docs/concepts/lore.md`, `provenance.md`

### Open Questions

- When to extend `VerificationStatus` beyond unverified/verified?

### Next Step

Superseded by Go REST (F104). Python REST handlers removed (`017-retire-python-core`).

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

None — complete. Python MCP removed (`017-retire-python-core`); Go `memlore mcp` is canonical.

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

**Status**: DONE
**Branch**: `016-authority-factors`
**Spec**: `specs/016-authority-factors/`

### Goal

Evaluate explainable authority factors at compile/explain time, derive a
score **and** trust band, and rank compiled context from those factors —
without opaque-only scores or a persistence table.

### Specification

`specs/016-authority-factors/` (specify → clarify → plan → tasks).

### Acceptance Criteria

- [x] Verified ADR with evidence → `canonical` (or `high` if evidence rule not met)
- [x] Unverified agent inference → `low`; verified agent inference → `high` never `canonical`
- [x] Factors + `trust_band` on compile/`get_for_task` items
- [x] `memlore.explain` and `GET /v1/lore-entries/{id}/explain` include evaluation + `factor_breakdown` (no NL summary)
- [x] Ranking uses evaluator (not only verified/unverified bases); invalidated cannot outrank unverified
- [x] No authority table; ephemeral evaluation
- [x] F112 filter/conflicts and F111 local-mode auth unchanged
- [x] `go test ./...` green

### Implementation

- `internal/domain/authority.go` — pure `EvaluateAuthority`
- `internal/application/authority/` — LoreEntry / GraphFact adapters
- Ranking, compile, explain presenters, REST GET explain

### Next Step

F010 remainder — team/project membership-scoped authorization.

---

## F113 — Retire legacy Python governance core

**Status**: DONE  
**Branch**: `017-retire-python-core`  
**Spec**: `specs/017-retire-python-core/`

### Goal

Remove the stale Python FastAPI REST/MCP app (`src/memlore/`) now that Go is
the production path. Keep `graph-service/` as the Python Graphiti boundary.

### Acceptance Criteria

- [x] `src/memlore/`, root `tests/`, Alembic, root `pyproject.toml` removed
- [x] CI root Python `quality` job removed; graph-service job kept
- [x] README / setup / constitution point at Go core
- [x] `go test ./...` green

### Next Step

F010 remainder — team/project membership.

---

## F104 — Migrate lore CRUD/verify REST to Go

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
- [x] Python serve later removed (F113)

### Implementation

- `internal/application/commands/`, `queries/`
- `internal/adapters/http/`
- `internal/infrastructure/memory/` (contract tests)
- `cmd/memlore serve`

### Next Step

F106a then F113 — Python adapters deprecated and removed.

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
- [x] Python MCP later removed (F113)

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

F113 retired the Python core. Next product work: F010 membership.

---

## F111 — OIDC authentication + RBAC

**Status**: DONE  
**Branch**: `015-oidc-rbac`  
**Spec**: `specs/015-oidc-rbac/`  
**Implements**: Product F010 (partial — no team/project membership yet)

### Goal

Optional OIDC Bearer JWT auth with reader/writer/admin RBAC. Local mode
preserves `X-Memlore-Actor` / `actor_id` when OIDC is unset.

### Acceptance Criteria

- [x] OIDC optional until configured
- [x] Actor from token `sub` when OIDC on; header spoofing ignored
- [x] Role matrix enforced (reader/writer/admin)
- [x] REST + MCP parity; `/health` open
- [x] `go test ./...` green

### Next Step

F010 remainder — team/project membership-scoped authorization.

---

## F112 — Temporal filtering + conflict detection

**Status**: DONE  
**Branch**: `014-conflict-filtering`  
**Spec**: `specs/014-conflict-filtering/`  
**Implements**: Product F009; completes F007 temporal/conflict stages

### Goal

Omit superseded/invalidated lore from default retrieval; surface structural
conflict groups on compiled context. History remains via get/explain.

### Acceptance Criteria

- [x] Default list/search/knowledge_search/get_for_task omit stale
- [x] `include_stale` opt-in on list/search/knowledge_search
- [x] get/explain still return stale
- [x] Compile `conflicts` array; neither side silently dropped
- [x] REST + MCP parity; still 9 MCP tools
- [x] `go test ./...` green

### Implementation

- `internal/domain/lore.go` (`IsCurrent`)
- `internal/application/context/{current,conflicts}.go`
- `internal/application/queries/{list_lore_by_scope,search_knowledge,compile_context}.go`
- Presenters + HTTP/MCP adapters

### Next Step

F111 OIDC/RBAC (DONE).

---

## F110 — Invalidate + supersede (governance lifecycle)

**Status**: DONE  
**Branch**: `013-supersede-invalidate`  
**Spec**: `specs/013-supersede-invalidate/`  
**Implements**: Product F008 (lifecycle; retrieval filtering is F112)

### Goal

Mark lore invalidated without deleting history, and supersede lore with a
successor that preserves the predecessor. REST and MCP (`memlore.invalidate`,
`memlore.supersede`) share the same entry shape.

### Acceptance Criteria

- [x] `verification_status=invalidated`; idempotent re-invalidate
- [x] `superseded_by_id` on predecessor; successor in same scope
- [x] Dual audits (`supersede` + `create`); explain chronological
- [x] REST `POST .../invalidate` and `POST .../supersede`
- [x] MCP nine tools including invalidate + supersede
- [x] Goose `00003` + sqlc mapping

### Implementation

- `internal/domain/lifecycle.go`
- `internal/application/commands/invalidate_lore.go`
- `internal/application/commands/supersede_lore.go`
- `migrations/00003_supersede_invalidate.sql`

### Next Step

F111 OIDC/RBAC (DONE).

---

## F109 — Context Compiler + get_for_task

**Status**: DONE  
**Branch**: `012-context-compiler`  
**Spec**: `specs/012-context-compiler/`  
**Implements**: Product F007 (completed via F112 temporal/conflict stages)

### Goal

Compile token-budgeted ContextPacket for agents: retrieve via F108, authority
ranking, dedup, token budgeting. REST `POST /v1/context/compile` and MCP
`memlore.get_for_task`.

### Acceptance Criteria

- [x] `CompileContextHandler` with authority ranking and token budget
- [x] Verified governance outranks graph hits (unit test)
- [x] Cross-plane statement dedup (v1)
- [x] `POST /v1/context/compile` + contract tests
- [x] `memlore.get_for_task` MCP tool + contract tests

### Implementation

- `internal/application/context/ranking.go`
- `internal/application/queries/compile_context.go`
- `internal/adapters/presenters/context_packet.go`

### Next Step

F111 — OIDC/RBAC.

---

## F108 — Graph Retrieval Orchestration

**Status**: DONE  
**Branch**: `011-graph-retrieval-orchestration`  
**Spec**: `specs/011-graph-retrieval-orchestration/`  
**Implements**: Product F006 (partial)

### Goal

Go application orchestrator parallel-fetches governance scope list (when scope
provided) and knowledge graph search via `ports.KnowledgeGraph`. REST
`POST /v1/knowledge-search` and MCP `memlore.knowledge_search` return merged
MemLore-shaped response. `memlore.search` unchanged.

### Acceptance Criteria

- [x] `SearchKnowledgeHandler` with parallel governance + graph fetch
- [x] Graceful graph degradation with `graph_service_unavailable` warning
- [x] `POST /v1/knowledge-search` REST endpoint + contract tests
- [x] `memlore.knowledge_search` MCP tool + contract tests
- [x] Bootstrap wiring in `memlore serve` and `memlore mcp`
- [x] Unit + integration tests; `memlore.search` contract unchanged

### Implementation

- `internal/application/queries/search_knowledge.go`
- `internal/adapters/presenters/knowledge_search.go`
- `internal/adapters/http/handlers.go` — `POST /v1/knowledge-search`
- `internal/adapters/mcp/tools.go` — `memlore.knowledge_search`

### Next Step

Supersede / invalidate (F008) or conflict detection (F009).

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
| F108 | F006 read path | F107 |
| F109 | F007 compiler v1 | F108 |

---

## Development Ledger Notes

### 2026-08-27 — F109 context compiler + get_for_task

- `CompileContextHandler` with authority ranking, dedup, token budgeting
- `POST /v1/context/compile` and `memlore.get_for_task`
- Implements product F007 (partial)

### 2026-08-27 — F108 graph retrieval orchestration

- `SearchKnowledgeHandler` parallel governance + graph search
- `POST /v1/knowledge-search` and `memlore.knowledge_search`
- Implements product F006 (partial read path)

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

1. Team/project membership-scoped authorization (F010 remainder)
2. Dogfood OIDC-on with HMAC or IdP JWKS

---

## Related Documents

- [MIGRATION_DISCOVERY.md](MIGRATION_DISCOVERY.md)
- [migration-inventory.md](migration-inventory.md)
- [../adr/README.md](../adr/README.md)
- [.specify/memory/constitution.md](../../.specify/memory/constitution.md)
