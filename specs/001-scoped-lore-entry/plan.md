# Implementation Plan: Scoped Human-Authored Lore Entry

**Branch**: `001-scoped-lore-entry` | **Date**: 2026-08-25 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-scoped-lore-entry/spec.md`

## Summary

Deliver the first MemLore product vertical slice: store, retrieve, verify, list
(by scope), and audit human-authored lore entries in the **governance plane**
(PostgreSQL), exposed via REST. Preserve provenance, structured scope
(`kind`+`key`), structured evidence (`type`+`value`), and explainable
verification status. Graphiti/Neo4j, MCP, OIDC/RBAC, conflicts, and supersession
remain out of scope.

## Technical Context

**Language/Version**: Python 3.12  
**Primary Dependencies**: FastAPI, Pydantic v2, SQLAlchemy 2.x, Alembic, psycopg
(PostgreSQL driver), uvicorn  
**Storage**: PostgreSQL 16 (governance plane only for this feature)  
**Testing**: pytest, pytest-asyncio where needed, httpx TestClient for API
contracts; testcontainers or Docker Compose Postgres for integration tests  
**Target Platform**: Local developer machine / Linux container; REST over HTTP  
**Project Type**: Backend service (hexagonal layout under `src/memlore/`)  
**Performance Goals**: Local create+id return under 30s wall-clock (SC-001);
typical single-row CRUD well under 1s locally  
**Constraints**: Domain independent of FastAPI/SQLAlchemy; no Graphiti dual-write;
actor identity via explicit request header (soft auth); statement max 8,000 chars;
duplicates allowed  
**Scale/Scope**: Single-process API; fixture-scale data for acceptance; not
multi-tenant hardened yet

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
  *(N/A conflicts this slice; verify is additive metadata, no overwrite of
  statement/origin; audits append-only)*
- [x] Secure by default: authz, tenant isolation, secret handling, untrusted
  agent context considered *(soft-auth explicit; RBAC deferred per spec)*
- [x] Observability: meaningful logs/metrics/traces/health for critical paths
  *(structured logs on create/verify/get; OTel wiring optional follow-up)*
- [x] Dependency policy: new third-party libraries justified
- [x] Simplicity: no speculative abstractions beyond requirements

**Post-design re-check**: Pass. Ports for lore/audit repositories keep domain
pure; REST DTOs stay in adapters; Alembic owns schema; no Neo4j involvement.

## Project Structure

### Documentation (this feature)

```text
specs/001-scoped-lore-entry/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── rest-lore-entries.md
└── tasks.md              # /speckit-tasks (not created by this command)
```

### Source Code (repository root)

```text
src/memlore/
├── domain/
│   ├── models/           # LoreEntry, Scope, Evidence, AuditRecord, enums
│   ├── services/         # verify rules, validation helpers if pure
│   └── policies/
├── application/
│   ├── commands/         # CreateLore, VerifyLore
│   ├── queries/          # GetLore, ListLoreByScope, ListAudits
│   ├── services/         # application orchestration
│   └── ports/            # LoreRepository, AuditRepository, UnitOfWork, Clock
├── infrastructure/
│   └── postgres/         # SQLAlchemy models, repositories, session, Alembic
├── adapters/
│   └── rest/             # routes, request/response schemas, error mapping
└── bootstrap/            # app wiring / DI

tests/
├── unit/                 # domain + application (fake ports)
├── integration/          # Postgres repositories + migrations
└── contract/             # REST OpenAPI-aligned HTTP tests
```

**Structure Decision**: MemLore default hexagonal layout. This feature fills
governance-plane persistence and REST adapters only.

## Complexity Tracking

> No constitution violations requiring justification.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
