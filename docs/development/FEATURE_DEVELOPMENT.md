# MemLore Feature Development Tracker

**Last Updated**: 2026-08-25  
**Current Milestone**: M1 — Go core skeleton specified (F101); Python governance MVP on main  
**Current Release Target**: v0.1.0 governance MVP (Python, merged); v0.2.0 Go skeleton (F101)

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
| F004 | Transactional outbox + graph sync | PLANNED | — | — | partial | — | No code |
| F005 | Graph knowledge service (Graphiti isolation) | PLANNED | — | — | partial | — | Greenfield Python service |
| F006 | Semantic search + graph retrieval | PLANNED | — | — | — | — | Depends on F005 |
| F007 | Context compiler + `get_for_task` | PLANNED | — | — | partial | — | |
| F008 | Supersession + invalidation | PLANNED | — | — | — | — | |
| F009 | Conflict detection | PLANNED | — | — | — | — | |
| F010 | Auth (OIDC) + team/project scopes | PLANNED | — | — | partial | — | Actor header only today |
| F101 | Go project skeleton + tooling | READY | ✓ | — | ✓ | 0005 | Spec on `003-go-core-skeleton` |
| F102 | Go domain primitives (lore/scope/evidence) | PLANNED | — | — | — | — | Migration |
| F103 | Go PostgreSQL persistence (sqlc/goose) | PLANNED | — | — | — | pending | Migration |
| F104 | Migrate lore CRUD/verify REST to Go | PLANNED | — | — | — | — | First vertical slice |
| F105 | Migrate MCP lore tools to Go | PLANNED | — | — | — | 0003 | After F104 |
| F106 | Extract graph-service + contracts | PLANNED | — | — | — | pending | Migration |

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

Maintain until Go port (F104) reaches Verified; then deprecate Python REST handlers.

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

None — complete. Python MCP remains until F105.

---

## F101 — Go Project Skeleton + Tooling

**Status**: READY  
**Branch**: `003-go-core-skeleton`  
**ADR**: [ADR-0005](../adr/0005-go-memlore-core.md)

### Goal

Additive Go module: `go.mod`, goose/sqlc scaffold, CI gates, layout contract.
No lore handlers; Python REST/MCP unchanged.

### Specification

`specs/003-go-core-skeleton/`

### Acceptance Criteria (draft)

- [ ] `go test ./...` and `go vet ./...` pass
- [ ] CI Go job added
- [ ] goose `00001` ports Alembic schema
- [ ] Python pytest still green

### TDD Progress

- [ ] RED — tasks T004, T008, T010, T013, T015 in `tasks.md`
- [ ] GREEN — implementation via `/speckit.implement`
- [ ] REFACTOR — layout polish

### Next Step

Run `/speckit.implement` on `specs/003-go-core-skeleton/tasks.md`.

---

## F104 — Migrate Lore CRUD/Verify REST to Go

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

## F101 — Go Project Skeleton + Tooling

**Status**: PLANNED  
**Migration feature**

### Goal

Introduce `go.mod`, directory layout, CI jobs, goose + sqlc scaffold without
changing runtime behavior.

### Specification

Pending — propose `specs/003-go-core-skeleton/` via Spec Kit

### Acceptance Criteria (draft)

- `go test ./...` and `go vet ./...` pass
- goose migration reproduces `0001` schema
- sqlc generates typed queries (may be unused until F103)
- Python app remains default entrypoint

### ADRs

- Proposed: ADR-0005 Go for MemLore Core (supersedes ADR-0002 in part)

### Next Step

Draft ADR-0005 + Spec Kit plan; no production traffic switch.

---

## F104 — Migrate Lore CRUD/Verify REST to Go

**Status**: PLANNED  
**Migration feature** — **recommended first Go vertical slice**

### Goal

Parity with F001 REST behavior using Go domain, sqlc, chi.

### Specification

Extend or fork `specs/001-scoped-lore-entry/` with Go implementation plan

### Acceptance Criteria (draft)

- Ported contract tests pass against Go server
- Characterization fixtures match Python outputs
- Python REST can run side-by-side until cutover
- No Graphiti dependency

### TDD Progress

- [ ] RED — port `tests/contract/test_create_lore_entry.py` to Go HTTP harness
- [ ] GREEN — minimal handlers
- [ ] REFACTOR — extract domain packages

### Next Step

Complete F101/F102/F103; export characterization vectors from Python tests.

---

## Migration Feature Map

| Migration ID | Replaces | Depends on |
|--------------|----------|------------|
| F101 | — | Discovery complete |
| F102 | Python domain types | F101 |
| F103 | SQLAlchemy repos | F101, F102 |
| F104 | F001 REST (Python) | F103 |
| F105 | F002 MCP (Python) | F104 |
| F106 | (greenfield) | F004 planning |

---

## Development Ledger Notes

### 2026-08-25 — F002 merged; F101 specified

- Merged `002-mcp-lore-tools` to `main` (F002 DONE)
- ADR-0005 accepted; Spec Kit `003-go-core-skeleton` complete (F101 READY)
- Added `docs/architecture/target-architecture.md`

### Immediate recommended tasks

1. `/speckit.implement` F101 (Go skeleton)
2. Characterization fixtures for F104

---

## Related Documents

- [MIGRATION_DISCOVERY.md](MIGRATION_DISCOVERY.md)
- [migration-inventory.md](migration-inventory.md)
- [../adr/README.md](../adr/README.md)
- [.specify/memory/constitution.md](../../.specify/memory/constitution.md)
