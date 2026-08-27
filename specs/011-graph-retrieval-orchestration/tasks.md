# Tasks: Graph Retrieval Orchestration (F108)

**Input**: Design documents from `/specs/011-graph-retrieval-orchestration/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Phase 1: Setup

- [x] T001 Create feature spec directory `specs/011-graph-retrieval-orchestration/`
- [x] T002 [P] Document contract in `specs/011-graph-retrieval-orchestration/contracts/knowledge-search.md`

## Phase 2: Foundational

- [x] T003 Add presenter types in `internal/adapters/presenters/knowledge_search.go`
- [x] T004 Extend `internal/infrastructure/memory/knowledge_graph.go` Search stub for tests

## Phase 3: User Story 1 — Parallel governance + graph search (P1)

**Goal**: Orchestrator merges governance and graph results  
**Independent Test**: Unit test with fake graph + memory UoW

- [x] T005 [US1] RED: Unit tests in `internal/application/queries/search_knowledge_test.go`
- [x] T006 [US1] GREEN: Implement `internal/application/queries/search_knowledge.go`

## Phase 4: User Story 2 — Graceful graph degradation (P1)

**Goal**: Graph down returns governance + warning  
**Independent Test**: Unit test with graph error

- [x] T007 [US2] Test graph degradation in `internal/application/queries/search_knowledge_test.go`

## Phase 5: User Story 3 — REST and MCP parity (P2)

**Goal**: REST and MCP expose same JSON shape  
**Independent Test**: Contract tests

- [x] T008 [P] [US3] REST handler + contract test in `internal/adapters/http/`
- [x] T009 [P] [US3] MCP `memlore.knowledge_search` in `internal/adapters/mcp/`
- [x] T010 [US3] Update MCP contract test tool list in `internal/adapters/mcp/mcp_contract_test.go`

## Phase 6: Bootstrap & Integration

- [x] T011 Wire graphclient in `cmd/memlore/main.go` for serve and mcp
- [x] T012 [P] Integration test in `internal/application/queries/search_knowledge_integration_test.go`

## Phase 7: Polish

- [x] T013 [P] Update `docs/api/rest.md` and `docs/api/mcp.md`
- [x] T014 Update `docs/development/FEATURE_DEVELOPMENT.md` and `.cursor/rules/specify-rules.mdc`

## Dependencies

- US1 → US2 (degradation builds on orchestrator)
- US1 → US3 (adapters need handler)
- T011 depends on T006, T008, T009

## Parallel Opportunities

- T008 REST and T009 MCP can proceed in parallel after T006
- T012 integration and T013 docs can run in parallel after adapters
