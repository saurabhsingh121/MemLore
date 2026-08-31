# Tasks: Fuller semantic search + graph retrieval (F006 remainder)

**Input**: Design documents from `/specs/019-semantic-graph-retrieval/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/  
**Tests**: TDD RED → GREEN → REFACTOR (constitution).

**Organization**: Phases follow US1–US4 from spec.md.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallelizable (different files, no incomplete deps)
- **[Story]**: US1…US4
- Paths are repo-relative under MemLore Go core

## Phase 1: Setup

- [x] T001 Confirm branch `019-semantic-graph-retrieval` and artifacts under `specs/019-semantic-graph-retrieval/`

---

## Phase 2: Foundational (relevance helpers + port)

**Purpose**: Shared relevance matching + repository search API  
**⚠️** No story wiring until this phase is green

- [x] T002 [P] RED: `internal/domain/relevance_test.go` — token AND / substring match; short tokens ignored; case-insensitive
- [x] T003 GREEN: `internal/domain/relevance.go` — `StatementMatchesQuery`, `RankByRelevance` helpers (pure)
- [x] T004 [P] Extend `internal/application/ports/repositories.go` with `SearchRelevant(ctx, SearchRelevantOpts) ([]domain.LoreEntry, error)`
- [x] T005 GREEN: memory `LoreRepository.SearchRelevant` in `internal/infrastructure/memory/repositories.go`
- [x] T006 [P] Add sqlc query in `db/queries/lore.sql` (`SearchLoreEntriesByStatement` ILIKE + optional scope); `sqlc generate`
- [x] T007 GREEN: Postgres `SearchRelevant` in `internal/infrastructure/postgres/lore_repository.go`
- [x] T008 DONE WHEN: `go test ./internal/domain/ ./internal/infrastructure/memory/ -count=1` green for relevance + memory search

**Checkpoint**: Can search lore by statement without knowledge handler changes

---

## Phase 3: User Story 1 — Query-relevant governance (P1) 🎯 MVP

**Goal**: Scoped knowledge_search returns only query-relevant governance  
**Independent Test**: Seed on-topic + off-topic in scope; only on-topic returned

### Tests first

- [x] T009 [P] [US1] RED: extend `internal/application/queries/search_knowledge_test.go` — scoped search excludes off-topic lore; empty when no match; include_stale respected

### Implementation

- [x] T010 [US1] GREEN: `SearchKnowledgeHandler` uses `SearchRelevant` (or list+filter) instead of raw `ListByScope` for governance path
- [x] T011 [US1] GREEN: apply limit after relevance; preserve graph path + `graph_service_unavailable`
- [x] T012 [US1] DONE WHEN: T009 green; existing graph degradation tests still pass

**Checkpoint**: MVP — no more full scope dumps on knowledge_search

---

## Phase 4: User Story 2 — Cross-plane receipt + dedupe (P1)

**Goal**: Prefer governance; attach `graph_receipt`; omit duplicate from graph.items  
**Independent Test**: Fact with provenance → lore has receipt; fact not in graph.items

### Tests first

- [x] T013 [P] [US2] RED: unit tests — provenance collapse; hydrate lore missed by text match; graph-only fact remains; inaccessible provenance omitted when membership filter supplied (handler-level or pure merge helper)

### Implementation

- [x] T014 [US2] GREEN: merge helper in `internal/application/queries/` (or search_knowledge) implementing Q2=B
- [x] T015 [US2] GREEN: presenters — `graph_receipt` on knowledge-search governance items (`internal/adapters/presenters/knowledge_search.go`); avoid breaking lore CRUD DTO if needed via knowledge-specific item type
- [x] T016 [US2] GREEN: wire REST/MCP presenters to new result shape
- [x] T017 [US2] DONE WHEN: T013 green; contract tests updated for additive `graph_receipt`

**Checkpoint**: Dual-plane coherent receipts

---

## Phase 5: User Story 3 — Contract stability + ranking (P2)

**Goal**: REST/MCP parity; relevance then authority tie-break  
**Independent Test**: Same payload shape both adapters; ordering unit test

- [x] T018 [P] [US3] RED/GREEN: update HTTP/MCP knowledge-search contract tests for relevance + optional `graph_receipt`
- [x] T019 [US3] GREEN: tie-break verified / authority when relevance equal (reuse F003 evaluate if cheap; else verified > unverified)
- [x] T020 [US3] DONE WHEN: REST + MCP contract suites green; `memlore.search` unchanged

---

## Phase 6: User Story 4 — Scope-less + authz (P2)

**Goal**: No scope → membership-allowed governance search  
**Independent Test**: Alpha member sees alpha hit, not beta, on scope-less search

- [x] T021 [P] [US4] RED: knowledge_search without scope returns relevant lore; membership contract asserts beta omitted
- [x] T022 [US4] GREEN: handler scope-less path + membership filter (reuse F114 hooks)
- [x] T023 [US4] DONE WHEN: T021 green; local-mode scope-less still works without membership seed

---

## Phase 7: Polish

- [x] T024 [P] Update `docs/api/rest.md` and `docs/api/mcp.md` for relevance, scope-less, `graph_receipt`
- [x] T025 [P] Mark F006 **DONE** in `docs/development/FEATURE_DEVELOPMENT.md`; refresh “Immediate recommended tasks”
- [x] T026 Run `go test ./...` (skip integration as usual); fix fallout
- [x] T027 Quickstart sanity against `specs/019-semantic-graph-retrieval/quickstart.md`

---

## Dependencies

```text
Phase 1 → Phase 2 → Phase 3 (US1 MVP) → Phase 4 (US2) → Phase 5 (US3) → Phase 6 (US4) → Phase 7
```

## Parallel opportunities

- T002/T004/T006 in Phase 2
- T009 || presenter sketch after T008
- T024/T025 docs after behavior green

## MVP

Ship Phase 3 (US1) alone for immediate agent value; US2–US4 complete F006 DONE.
