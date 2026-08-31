# Implementation Plan: Temporal filtering + conflict detection (F112)

**Branch**: `014-conflict-filtering` | **Date**: 2026-08-31 | **Spec**: [spec.md](./spec.md)

## Summary

Filter superseded and invalidated lore out of default retrieval
(list/search/knowledge_search/compile), keep get/explain historical, and
surface structural conflict groups on compiled context packets — completing
the deferred F109 pipeline stages without persisting conflicts or changing
F110 transitions.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: chi, pgx/v5, sqlc, goose, MCP Go SDK (no new libs)  
**Storage**: PostgreSQL governance (no schema change required for v1)  
**Testing**: `go test ./...`; domain/context unit; query handlers + memory UoW; HTTP/MCP contract  
**Target Platform**: MemLore core (`memlore serve`, `memlore mcp`)  
**Project Type**: hexagonal Go service  
**Performance Goals**: Same as existing list/search/compile paths (in-memory filter + O(n) conflict scan)  
**Constraints**: TDD; no new MCP tools; no graph-plane lifecycle; no conflict persistence; no F110 semantic changes  
**Scale/Scope**: Pure helpers + wire into list/search/compile + presenters + docs/contracts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED → GREEN → REFACTOR for filter, conflict detect, compile, contracts
- [x] Spec-driven: clarifications encoded; acceptance scenarios measurable
- [x] Architecture integrity: pure helpers in domain/application; adapters only map I/O; no Neo4j/Graphiti coupling
- [x] Documentation: REST/MCP API docs, FEATURE_DEVELOPMENT, contracts, specify-rules next=F111
- [x] Authority & provenance: conflicts surfaced with entry ids/statements; get/explain preserve history
- [x] Temporal correctness: stale excluded from default retrieval; history not deleted; conflicts not silently discarded
- [x] Secure by default: no auth changes (F111); include_stale is retrieval opt-in only
- [x] Observability: reuse existing warnings; conflicts are explicit response fields
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: small pure functions; ephemeral conflict metadata; no new subsystem/table

## Project Structure

### Documentation (this feature)

```text
specs/014-conflict-filtering/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── temporal-filter.md
│   └── conflict-detection.md
└── tasks.md
```

### Source Code

```text
internal/domain/lore.go                    # IsCurrent (alongside IsSuperseded)
internal/application/context/current.go    # FilterCurrent
internal/application/context/conflicts.go  # ConflictGroup, DetectConflicts
internal/application/context/ranking.go    # optional supersession_status / invalidated floor
internal/application/queries/list_lore_by_scope.go  # IncludeStale option
internal/application/queries/search_knowledge.go    # IncludeStale + filter
internal/application/queries/compile_context.go     # filter → detect → rank → budget
internal/adapters/presenters/context_packet.go      # conflicts[]
internal/adapters/http/handlers.go
internal/adapters/mcp/tools.go
docs/api/mcp.md
docs/api/rest.md
docs/development/FEATURE_DEVELOPMENT.md
.cursor/rules/specify-rules.mdc
```

**Structure Decision**: Application-layer filter (not SQL). Repository
`ListByScope` stays unfiltered. `ListLoreByScopeHandler` owns the default
current-only filter so search/list share one default. Compile additionally
filters (defense) then detects conflicts before ranking.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/temporal-filter.md](./contracts/temporal-filter.md)
- [contracts/conflict-detection.md](./contracts/conflict-detection.md)
- [quickstart.md](./quickstart.md)

## Phase 2 — Tasks

See [tasks.md](./tasks.md) (`/speckit-tasks`).

## Constitution re-check (post-design)

Gates still pass: pure helpers; no schema; REST/MCP parity; history preserved;
conflicts surfaced; no speculative persistence or LLM contradiction.
