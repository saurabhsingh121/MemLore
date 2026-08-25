# Implementation Plan: MCP Lore Tools

**Branch**: `002-mcp-lore-tools` | **Date**: 2026-08-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-mcp-lore-tools/spec.md`

## Summary

Expose the existing governance-plane lore application services to coding
agents through a **domain MCP adapter**: tools `memlore.remember`,
`memlore.get`, `memlore.verify`, `memlore.explain`, and `memlore.search`
(exact scope list), plus a local **stdio** server started with `memlore mcp`.
No new domain rules, no Graphiti/Neo4j, no Streamable HTTP, no OIDC. Mutating
tools require an explicit `actor_id` argument; `remember` always stores
`human_authored` origin (REST parity).

## Technical Context

**Language/Version**: Python 3.12  
**Primary Dependencies**: existing stack (FastAPI REST, Pydantic v2,
SQLAlchemy 2.x, Alembic, psycopg) **plus** official MCP SDK `mcp>=2.1,<3`
(`MCPServer`, stdio, in-memory `Client`)  
**Storage**: PostgreSQL 16 governance plane (reuse 001 schema; **no** new
migrations)  
**Testing**: pytest + pytest-asyncio; MCP contract tests via in-memory
`Client(server)`; sparse stdio e2e covering remember → get → verify →
explain → search; Docker Compose Postgres when e2e/integration need DB  
**Target Platform**: Local developer machine; coding-agent stdio attach  
**Project Type**: Backend service + CLI (hexagonal `src/memlore/`)  
**Performance Goals**: Agent remember→get→verify→explain→search under 2
minutes locally (SC-001); typical tool call well under 1s locally  
**Constraints**: Domain independent of MCP SDK/FastAPI/SQLAlchemy; stdout
reserved for MCP protocol (logs on stderr); no env-default actor; no
Graphiti tools on the list; HTTP MCP deferred  
**Scale/Scope**: Five tools; single stdio process; fixture-scale data

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: behavioral work planned as RED → GREEN → REFACTOR; no
  retroactive test-only compliance
- [x] Spec-driven: measurable acceptance criteria exist; ambiguous
  behavior/architecture clarified before irreversible choices
- [x] Architecture integrity: domain independent of FastAPI/Postgres/Neo4j/
  Graphiti/Redis; governance vs knowledge plane boundaries preserved;
  no distributed transactions across Postgres and Neo4j
- [x] Documentation: docs/ADR updates included in the same unit of work
- [x] Authority & provenance: human vs agent origins, evidence, verification,
  and explainable authority factors preserved
- [x] Temporal correctness: history not overwritten; conflicts surfaced
  *(N/A conflicts; verify/audit semantics unchanged from 001)*
- [x] Secure by default: authz, tenant isolation, secret handling, untrusted
  agent context considered *(required `actor_id` tool arg; no env default;
  tool input validated; RBAC/OIDC still deferred)*
- [x] Observability: meaningful logs/metrics/traces/health for critical paths
  *(structured logs per MCP tool on stderr; OTel still optional)*
- [x] Dependency policy: new third-party libraries justified
- [x] Simplicity: no speculative abstractions beyond requirements

**Post-design re-check**: Pass. MCP lives in `adapters/mcp/` and calls existing
application handlers through `AppContainer`. Explain composes get + audit list
in one UoW (no new domain). Official `mcp` SDK is justified. Docs (`docs/api/mcp.md`,
README, setup) update in the same feature. No Postgres/Neo4j dual-write.

## Project Structure

### Documentation (this feature)

```text
specs/002-mcp-lore-tools/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── mcp-lore-tools.md
└── tasks.md              # /speckit-tasks (not created by this command)
```

### Source Code (repository root)

```text
src/memlore/
├── domain/                 # unchanged (no new rules)
├── application/            # unchanged handlers; MCP calls these
├── infrastructure/         # unchanged Postgres; logging stderr for MCP CLI
├── adapters/
│   ├── rest/               # unchanged behavior; schemas reused for MCP JSON
│   ├── mcp/
│   │   ├── server.py       # create_mcp_server(container) → MCPServer
│   │   ├── tools.py        # remember/get/verify/explain/search wiring
│   │   └── errors.py       # ValidationError/NotFoundError → ToolError
│   └── cli/main.py         # add `mcp` subcommand (stdio)
└── bootstrap/              # reuse AppContainer / Settings / DSN

tests/
├── unit/adapters/mcp/      # mapping + actor_id validation with fakes
├── contract/               # existing REST tests remain
│   └── mcp/                # in-memory Client: tools + error contract
├── integration/            # no new repo tests unless CLI/DSN wiring needs them
└── e2e/
    └── test_mcp_stdio.py   # spawn `memlore mcp`; five-tool path (skip if no PG)
```

**Structure Decision**: Default hexagonal layout. This feature fills the MCP
adapter and CLI entrypoint only. Application/domain/persistence stay as in
001 unless a tiny shared helper is required for DTO mapping.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
