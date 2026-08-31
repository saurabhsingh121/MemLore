# Tasks: Authority factor model + evaluation (F003)

**Input**: Design documents from `/specs/016-authority-factors/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Constitution TDD — RED → GREEN → REFACTOR for every behavioral task.

## Phase 1: Setup

- [x] T001 Confirm branch `016-authority-factors` and artifacts under `specs/016-authority-factors/`

---

## Phase 2: Foundational — domain evaluator

**Goal**: Pure EvaluateAuthority + golden band/score matrix. Blocks all stories.

- [x] T002 RED: Golden tests for band matrix and ordering in `internal/domain/authority_test.go` (canonical ADR; high verified; medium unverified human; low agent inference; low graph; untrusted invalidated; verified agent inference → high not canonical; superseded cap; invalidated cannot outrank unverified)
- [x] T003 GREEN: `TrustBand`, `SourceType`, `FactorInputs`, `EvaluateAuthority` in `internal/domain/authority.go`
- [x] T004 [P] RED: Application mapper tests in `internal/application/authority/evaluate_test.go` (LoreEntry + GraphFact + requested scope)
- [x] T005 GREEN: `EvaluateGovernance` / `EvaluateGraph` in `internal/application/authority/evaluate.go`

---

## Phase 3: User Story 1 — Compile factors, bands, ranking (P1) 🎯 MVP

**Goal**: RankAndDedup and compile items expose factors + trust_band; ranking uses evaluator.

**Independent Test**: Mixed fixtures rank verified+ADR > verified > unverified human > graph > agent inference; packet items include `trust_band`.

- [x] T006 [P] [US1] RED: Ranking tests in `internal/application/context/ranking_test.go` (order regressions; new factor keys; trust_band; keep invalidated floor)
- [x] T007 [US1] GREEN: Call evaluator from `RankAndDedup` in `internal/application/context/ranking.go`; add `TrustBand` on `RankedItem`; extend `AuthorityFactors`
- [x] T008 [P] [US1] RED: Compile handler tests include `trust_band` + new factors on items (`internal/application/queries/compile_context_test.go`)
- [x] T009 [US1] GREEN: Pass requested compile scope into ranking from `compile_context.go`
- [x] T010 [US1] GREEN: Present `trust_band` and new factor keys in `internal/adapters/presenters/context_packet.go`

---

## Phase 4: User Story 2 — Agent inference hard cap (P1)

**Goal**: Unverified agent origin never canonical/high; verified agent never canonical.

**Independent Test**: Domain golden tests already cover; ranking mixed set with verified human ADR vs verified agent inference.

- [x] T011 [P] [US2] RED: Ranking fixture — verified agent inference below verified human ADR (`ranking_test.go`)
- [x] T012 [US2] GREEN: Confirm origin penalties + band rules in domain evaluator (no extra framework)

---

## Phase 5: User Story 3 — Explain surfaces evaluation (P1)

**Goal**: MCP explain + REST GET explain return evaluation fields; no summary essay.

**Independent Test**: Explain existing id includes trust_band, factors, breakdown; unknown → not_found; stale still explained.

- [x] T013 [P] [US3] RED: `ExplainLoreHandler` tests in `internal/application/queries/explain_lore_test.go`
- [x] T014 [US3] GREEN: `ExplainLoreHandler` in `internal/application/queries/explain_lore.go`
- [x] T015 [US3] GREEN: Extend `ExplainResult` in `internal/adapters/presenters/lore.go`; wire MCP `explain` in `internal/adapters/mcp/tools.go`
- [x] T016 [P] [US3] RED: HTTP contract `GET /v1/lore-entries/{id}/explain` in `internal/adapters/http/lore_contract_test.go` (or dedicated test file)
- [x] T017 [US3] GREEN: REST route + handler in `internal/adapters/http/handlers.go`
- [x] T018 [P] [US3] RED: MCP explain contract asserts authority fields and no `summary` in `internal/adapters/mcp/mcp_contract_test.go`

---

## Phase 6: User Story 4 — F112 pipeline intact (P2)

**Goal**: Stale still omitted from compile items; conflicts still listed; graph dedup still works.

- [x] T019 [US4] Confirm existing compile/conflict tests still pass; add regression that stale items lack fancy compile scores because they are filtered first (`compile_context_test.go`)

---

## Phase 7: Contracts + docs

- [x] T020 [P] Update `specs/012-context-compiler/contracts/context-compile.md` and `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md`
- [x] T021 [P] HTTP compile contract asserts `trust_band` + factor keys (`context_compile_contract_test.go`); MCP `get_for_task` same (`mcp_contract_test.go`)
- [x] T022 [P] Update `docs/architecture/authority-model.md`, `docs/concepts/authority.md`, `docs/architecture/target-architecture.md` (pipeline: filter then evaluate)
- [x] T023 [P] Update `docs/api/mcp.md`, `docs/api/rest.md`, `docs/development/FEATURE_DEVELOPMENT.md`, `.cursor/rules/specify-rules.mdc`
- [x] T024 Run `go test ./...`
- [x] T025 Optional dogfood per `quickstart.md`

## Dependency graph

```text
T001 → T002 → T003 → T004 → T005
     → T006 → T007 → T008 → T009 → T010
     → T011 → T012
     → T013 → T014 → T015 → T016 → T017 → T018
     → T019
     → T020/T021/T022/T023 → T024 → T025
```

## Parallel examples

- T004 mapper tests after T003
- T006 ranking RED in parallel with T008 compile RED once evaluator exists
- T016 HTTP and T018 MCP explain contracts in parallel after presenter exists
- T020–T023 docs in parallel after behavior is green

## Implementation strategy

MVP: Phase 2 + User Story 1 (compile factors + ranking). Then agent-cap (mostly already in domain tests), then explain, then docs.
