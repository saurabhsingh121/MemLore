# Implementation Plan: Go Core Project Skeleton (F101)

**Branch**: `003-go-core-skeleton` | **Date**: 2026-08-25 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/003-go-core-skeleton/spec.md`  
**ADR**: [ADR-0005](../../docs/adr/0005-go-memlore-core.md)

## Summary

Add an **additive** Go module for MemLore Core: `go.mod`, minimal `cmd/memlore`
entrypoint, hexagonal `internal/` stubs, goose migration port of Alembic `0001`,
sqlc scaffold, and CI `go test`/`go vet` gates. **No** lore handlers, **no**
MCP server, **no** traffic switch — Python REST/MCP remain default.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi (declared, minimal use), pgx/v5, sqlc (codegen),
goose (migrations), slog (stdlib)  
**Storage**: PostgreSQL 16 — goose DDL only in F101; no runtime DB connection
required for default `go test`  
**Testing**: `go test ./...`, `go vet ./...`; optional build-tagged goose
integration test  
**Target Platform**: Linux/macOS dev; GitHub Actions CI  
**Project Type**: Monorepo — Go core + Python legacy + future `graph-service/`  
**Performance Goals**: `go test ./...` completes in under 30s locally (SC-001)  
**Constraints**: Domain packages must not import chi/pgx/MCP; Python tests must
stay green; no Alembic removal  
**Scale/Scope**: Skeleton only — one migration file, one sqlc smoke query,
minimal main

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: tasks ordered RED → GREEN → REFACTOR for behavioral checks (version
  command, migration parity tests)
- [x] Spec-driven: spec.md acceptance criteria and SC-001–SC-004 defined
- [x] Architecture integrity: `internal/domain` stub has no infra imports;
  strangler preserves Python
- [x] Documentation: ADR-0005, feature tracker, migration inventory updates
  in same unit of work
- [x] Authority & provenance: N/A (no lore behavior in F101)
- [x] Temporal correctness: N/A
- [x] Secure by default: no new attack surface; no public Go server required
- [x] Observability: slog in stub main; OTel deferred
- [x] Dependency policy: chi, pgx, sqlc, goose justified in research.md
- [x] Simplicity: no worker, no HTTP API, no domain lore types yet

**Post-design re-check**: Pass. F101 is tooling-only; F102+ fill domain.

## Project Structure

### Documentation (this feature)

```text
specs/003-go-core-skeleton/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── go-module-layout.md
├── checklists/
│   └── requirements.md
└── tasks.md
```

### Source Code (repository root — created by F101 implementation)

```text
go.mod
go.sum
sqlc.yaml
cmd/
└── memlore/
    └── main.go              # version/health stub; exit 0
internal/
├── domain/
│   └── doc.go               # package docs only (or version.go constant)
├── application/
│   └── doc.go
├── adapters/
│   └── doc.go
└── infrastructure/
    └── postgres/
        └── doc.go
db/
└── queries/
    └── smoke.sql            # SELECT 1 AS ok
migrations/
└── 00001_lore_audit.sql     # port of Alembic 0001
internal/infrastructure/postgres/sqlc/   # generated (gitignored or committed)
```

**Structure Decision**: Minimal files per research R10. Expand in F102–F104.

## Phase 0 — Research

Completed in [research.md](./research.md). No NEEDS CLARIFICATION remain.

## Phase 1 — Design artifacts

- [data-model.md](./data-model.md) — DDL port reference (no new tables)
- [contracts/go-module-layout.md](./contracts/go-module-layout.md) — layout contract
- [quickstart.md](./quickstart.md) — contributor commands

## Phase 2 — Implementation (via /speckit.implement)

Follow [tasks.md](./tasks.md). Stop after skeleton green; do not start F104.

## Risk Notes

| Risk | Mitigation |
|------|------------|
| goose vs Alembic drift | Characterization: compare `\d` output / information_schema |
| CI Go version missing | `setup-go` with `go-version: '1.25.x'` |
| sqlc gen not in CI | Commit generated code OR run sqlc in CI before test |

**Decision**: Commit sqlc-generated package in F101 so `go test` works without
codegen step in CI.

## Dependencies

- ADR-0005 accepted
- `main` includes F001 + F002 (merged)
- Migration discovery docs present
