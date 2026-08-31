# Tasks: Temporal filtering + conflict detection (F112)

**Input**: Design documents from `/specs/014-conflict-filtering/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Constitution TDD — RED → GREEN → REFACTOR for every behavioral task.

## Phase 1: Setup

- [x] T001 Confirm branch `014-conflict-filtering` and artifacts under `specs/014-conflict-filtering/`

---

## Phase 2: Foundational — current helpers + conflict detect

- [x] T002 [P] RED: `IsCurrent` / filter tests in `internal/domain/lore_test.go` and/or `internal/application/context/current_test.go`
- [x] T003 GREEN: `IsCurrent` on `internal/domain/lore.go`; `FilterCurrent` in `internal/application/context/current.go`
- [x] T004 [P] RED: conflict detection tests in `internal/application/context/conflicts_test.go` (same statement ≠ conflict; different = conflict; multi-scope)
- [x] T005 GREEN: `ConflictGroup` + `DetectConflicts` in `internal/application/context/conflicts.go`
- [x] T006 [P] RED: ranking safety — invalidated must not outrank unverified if it reaches scoring (`internal/application/context/ranking_test.go`)
- [x] T007 GREEN: floor invalidated score in `governanceScore` (defense); keep verified > unverified > invalidated

---

## Phase 3: User Story 1 — Stale omitted from retrieval (P1)

- [x] T008 [P] [US1] RED: `ListLoreByScopeHandler` tests with mix of current/superseded/invalidated + `IncludeStale`
- [x] T009 [US1] GREEN: `ListLoreByScopeQuery` / `IncludeStale` on `list_lore_by_scope.go`; default filter current
- [x] T010 [P] [US1] RED: `SearchKnowledgeHandler` tests omit stale by default; include when flagged
- [x] T011 [US1] GREEN: wire `IncludeStale` through `search_knowledge.go`
- [x] T012 [P] [US1] RED: compile tests — stale never in `items` (`compile_context_test.go`)
- [x] T013 [US1] GREEN: `CompileContextHandler` filter before rank; never pack stale

---

## Phase 4: User Story 2 — Conflicts surfaced (P1)

- [x] T014 [P] [US2] RED: compile with two disagreeing current statements → `conflicts` populated; neither dropped from ranking candidates
- [x] T015 [P] [US2] RED: budget excludes one side but conflict group still lists both ids
- [x] T016 [US2] GREEN: attach `Conflicts` on `CompileContextResult`; detect after filter, before rank
- [x] T017 [US2] GREEN: present `conflicts` on `ContextPacket` in `context_packet.go`

---

## Phase 5: User Story 3 — Pipeline + ranking order (P2)

- [x] T018 [US3] RED/GREEN: compile order regression — filter → detect → rank → budget covered by handler tests
- [x] T019 [US3] Confirm graph warnings preserved alongside empty/non-empty conflicts

---

## Phase 6: User Story 4 — REST + MCP parity (P2)

- [x] T020 [P] [US4] RED: HTTP contract — list default omits stale; `include_stale=true` includes; compile returns `conflicts`
- [x] T021 [P] [US4] RED: MCP contract — search/knowledge_search/get_for_task omit stale; get/explain still return stale
- [x] T022 [US4] GREEN: HTTP DTOs/handlers for `include_stale` + conflicts
- [x] T023 [US4] GREEN: MCP tool args for `include_stale`; get_for_task packet shape
- [x] T024 [US4] Update prior contracts: `specs/012-.../context-compile.md`, `specs/011-.../knowledge-search.md`, `specs/001-.../rest-lore-entries.md`

---

## Phase 7: Polish & docs

- [x] T025 [P] Update `docs/api/mcp.md`, `docs/api/rest.md`
- [x] T026 [P] Update `docs/development/FEATURE_DEVELOPMENT.md`; mark F007 closer / F009 via F112; specify-rules next=F111
- [x] T027 Run `go test ./...`
- [x] T028 Optional dogfood: two contradictory current rules + superseded predecessor → get_for_task

## Dependency graph

```text
T001 → T002/T004/T006 → T003/T005/T007
     → T008/T010/T012 → T009/T011/T013
     → T014/T015 → T016/T017
     → T018/T019
     → T020/T021 → T022/T023/T024
     → T025/T026/T027/T028
```

## Parallel examples

- T002, T004, T006 in parallel (RED helpers)
- T008, T010, T012 in parallel (RED query handlers)
- T020, T021 in parallel (RED contracts)
- T025, T026 in parallel (docs)
