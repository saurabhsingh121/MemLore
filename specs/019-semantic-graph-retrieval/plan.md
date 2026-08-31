# Implementation Plan: Fuller semantic search + graph retrieval (F006 remainder)

**Branch**: `019-semantic-graph-retrieval` | **Date**: 2026-09-01 | **Spec**: [spec.md](./spec.md)

## Summary

Upgrade `SearchKnowledgeHandler` so governance hits are **query-relevant**
(not a full scope dump), support **scope-less** search across
membership-allowed lore, and **prefer governance** when a graph fact’s
provenance points at a lore entry (attach graph as a receipt; omit standalone
duplicate). Keep F108 envelope additive; finish docs; mark F006 DONE.

## Technical Context

**Language/Version**: Go 1.25+  
**Primary Dependencies**: existing chi/pgx/sqlc/goose/MCP stack (no new libs)  
**Storage**: PostgreSQL — additive sqlc query for statement relevance; no new migration unless a simple `pg_trgm` index is justified in research (default: no migration)  
**Testing**: `go test ./...`; unit tests on relevance/merge; REST/MCP contract updates; membership isolation regression  
**Target Platform**: `memlore serve`, `memlore mcp`  
**Project Type**: CLI + REST + MCP governance service  
**Performance Goals**: Default limit 10; relevance filter before/with limit; avoid loading entire tenant lore when scoped search can use SQL `ILIKE`  
**Constraints**: TDD; domain pure; graph-service untouched unless provenance already present (it is); F114/F112 behavior preserved; MCP tool count unchanged  
**Scale/Scope**: Orchestration quality for F006 close-out — not Postgres embeddings product

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] TDD: RED→GREEN for relevance filter, scope-less authz, receipt dedupe, graph-down path
- [x] Spec-driven: clarifications Q1=B, Q2=B encoded; SC-001–SC-006
- [x] Architecture: Go orchestration + ports; no Graphiti types in adapters; no distributed TX
- [x] Documentation: rest.md, mcp.md, FEATURE_DEVELOPMENT F006 → DONE; optional contract pointer
- [x] Authority & provenance: graph receipt preserves score/provenance; lore remains SoT
- [x] Temporal correctness: `include_stale` + current-only defaults unchanged
- [x] Secure by default: membership filter on scope-less governance; inaccessible provenance → no leak
- [x] Observability: keep `graph_service_unavailable` warning; slog on graph errors
- [x] Dependency policy: no new third-party libraries
- [x] Simplicity: ILIKE/token relevance + provenance merge; no embedding store

**Post-design re-check**: Pass — single handler enhancement + repo search method +
additive presenter field; graph-service unchanged.

## Project Structure

### Documentation (this feature)

```text
specs/019-semantic-graph-retrieval/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/knowledge-search-v2.md
└── tasks.md
```

### Source Code (repository root)

```text
db/queries/lore.sql                              # SearchLoreByStatement (+ optional scope)
internal/application/ports/repositories.go       # SearchRelevant / SearchByQuery
internal/infrastructure/{postgres,memory}/       # impl
internal/application/queries/search_knowledge.go # relevance + merge + scope-less
internal/application/queries/search_knowledge_*_test.go
internal/adapters/presenters/knowledge_search.go # graph_receipt additive
internal/adapters/http|mcp                       # wire if signature changes
docs/api/{rest,mcp}.md
docs/development/FEATURE_DEVELOPMENT.md
```

**Structure Decision**: Go core only. Do not modify `graph-service/` (ingest
already sets `provenance_refs: [lore_id]`).

## Complexity Tracking

> No constitution violations requiring justification.

## Phase 0 — Research

See [research.md](./research.md).

## Phase 1 — Design

- [data-model.md](./data-model.md)
- [contracts/knowledge-search-v2.md](./contracts/knowledge-search-v2.md)
- [quickstart.md](./quickstart.md)

## Phase 2 — Tasks

See [tasks.md](./tasks.md) (produced by `/speckit-tasks`).

## Implementation approach (planning notes)

### Governance relevance

1. Parallel: graph `Search` (unchanged) + governance candidate fetch.
2. Scoped: SQL/memory search of lore in that scope whose `statement` matches
   query tokens (case-insensitive contains / all significant tokens).
3. Scope-less: same match without scope predicate; then filter by
   `CanAccessScope` when membership enforced.
4. Apply temporal filter (`include_stale`) then `limit`.
5. Tie-break: verified / higher authority score when equal relevance.

### Cross-plane merge (Q2=B)

1. For each graph fact, if any `provenance_refs` id resolves to an accessible
   lore entry L:
   - Ensure L is in governance set (hydrate if missing but accessible +
     temporal-allowed).
   - Attach `graph_receipt` (id, score, statement, provenance_refs) on that
     governance DTO.
   - **Remove** that fact from `graph.items`.
2. If provenance lore missing/inaccessible: drop the link; if fact has no other
   usable link, keep in `graph.items` only when the fact itself is not leaking
   inaccessible lore content beyond what graph already returns — prefer
   omitting facts whose only provenance is inaccessible when membership on
   (fail closed for tenant).
3. Graph-only facts (empty/non-lore provenance) remain in `graph.items`.

### Authz

Reuse F114 gates already wrapping knowledge_search; scope-less path must still
filter governance by membership. Explicit inaccessible single scope → existing
forbidden behavior unchanged.
