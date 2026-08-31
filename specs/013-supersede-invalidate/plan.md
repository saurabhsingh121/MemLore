# Implementation Plan: Governance lifecycle — invalidate + supersede (F110)

**Branch**: `013-supersede-invalidate` | **Date**: 2026-08-28 | **Spec**: [spec.md](./spec.md)

## Summary

Add domain transitions `ApplyInvalidation` and `ApplySupersession`, command
handlers with UoW (mirror `VerifyLoreHandler`), goose `00003` columns, sqlc
query updates, REST `POST .../invalidate` and `POST .../supersede`, and MCP
tools `memlore.invalidate` + `memlore.supersede`. Tool count becomes 9.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi, pgx/v5, sqlc, goose, MCP Go SDK  
**Storage**: PostgreSQL `lore_entries` + `audit_records` (additive goose 00003)  
**Testing**: `go test ./...`; domain unit; command + memory UoW; HTTP/MCP contract; postgres integration if DB up  
**Target Platform**: MemLore core CLI (`memlore mcp`, `memlore serve`)  
**Project Type**: hexagonal Go service (domain / application / adapters / postgres)  
**Performance Goals**: Same as existing lore mutate path (single UoW transaction)  
**Constraints**: TDD RED→GREEN→REFACTOR; no new outbox event types; no graph-service lifecycle; no F112 filtering  
**Scale/Scope**: Two commands, two REST routes, two MCP tools, one migration, presenter field additions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: behavioral work planned as RED → GREEN → REFACTOR; no retroactive test-only compliance
- [x] Spec-driven: measurable acceptance criteria exist; schema/idempotency clarified in spec
- [x] Architecture integrity: domain functions stay framework-free; Postgres via ports; no Neo4j/Graphiti coupling
- [x] Documentation: `docs/api/mcp.md`, REST contract, FEATURE_DEVELOPMENT in same unit of work
- [x] Authority & provenance: invalidate/supersede preserve evidence and audits; origin stays human_authored on successor
- [x] Temporal correctness: predecessor retained; superseded_by_id link; no overwrite of statement
- [x] Secure by default: explicit actor_id; no env default; F111 auth still out of scope
- [x] Observability: existing slog/error mapping; no new speculative telemetry
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: mirror VerifyLoreHandler; no speculative supersession graph API

## Project Structure

### Documentation (this feature)

```text
specs/013-supersede-invalidate/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── lifecycle-lore.md
└── tasks.md
```

### Source Code

```text
migrations/00003_supersede_invalidate.sql
db/queries/lore.sql
internal/domain/enums.go
internal/domain/lore.go
internal/domain/verification.go          # ApplyVerification rejects invalidated/superseded
internal/domain/lifecycle.go             # ApplyInvalidation, ApplySupersession
internal/application/commands/invalidate_lore.go
internal/application/commands/supersede_lore.go
internal/adapters/presenters/lore.go
internal/adapters/http/handlers.go
internal/adapters/mcp/server.go
internal/adapters/mcp/tools.go
internal/infrastructure/postgres/mapping.go
```

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/lifecycle-lore.md](./contracts/lifecycle-lore.md)
- [quickstart.md](./quickstart.md)

## Phase 2 — Tasks

See [tasks.md](./tasks.md) (`/speckit-tasks`).

## Constitution re-check (post-design)

Gates still pass: domain stays pure; persistence additive; REST/MCP adapters
only; history preserved; no graph-plane writes.
