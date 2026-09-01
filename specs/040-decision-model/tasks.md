# Tasks: First-Class Decision Model (F040)

**Input**: Design documents from `/specs/040-decision-model/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/decisions.md

**Tests**: TDD (RED → GREEN → REFACTOR) for all behavioral work.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **MemLore default**: `cmd/memlore/`, `internal/`, `migrations/`, `db/queries/`

## Phase 1: Setup

**Purpose**: Feature directory already exists; pin wiring points

- [x] T001 Confirm feature artifacts exist in specs/040-decision-model/ (spec.md, plan.md, research.md, data-model.md, contracts/decisions.md, quickstart.md)
- [x] T002 [P] Verify .specify/feature.json points at specs/040-decision-model

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Decision aggregate, lore constructor, ports/schema — required before any story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Add failing tests for NewDecisionLoreEntry (verified human_authored, optional evidence, NewLoreEntry still rejects non-human, architecture constructor still requires adr) in internal/domain/lore_test.go
- [x] T004 [P] Add failing tests for Decision create/validate (required question/choice/owner, alternatives labels, IsCurrent, ProjectADRDecision, human vs adr source) in internal/domain/decision_test.go
- [x] T005 Implement NewDecisionLoreEntry in internal/domain/lore.go (GREEN T003)
- [x] T006 Implement Decision, DecisionAlternative, NewHumanDecision, ProjectADRDecision, MarkSuperseded in internal/domain/decision.go (GREEN T004)
- [x] T007 Add DecisionRepository port in internal/application/ports/decision.go; extend UnitOfWork in internal/application/ports/repositories.go
- [x] T008 Add goose migration migrations/00009_decisions.sql and sqlc queries db/queries/decisions.sql; generate or hand-write sqlc under internal/infrastructure/postgres/sqlc/ (v1.28.0 style)
- [x] T009 Wire Decision repos on postgres UnitOfWork in internal/infrastructure/postgres/unit_of_work.go and memory UnitOfWork in internal/infrastructure/memory/repositories.go (plus internal/infrastructure/memory/decision.go and internal/infrastructure/postgres/decision_repository.go)

**Checkpoint**: Domain can represent human Decisions and ADR projections; UoW can persist decision rows in tests

---

## Phase 3: User Story 1 - Record a human engineering decision (Priority: P1) 🎯 MVP

**Goal**: Create a structured Decision with linked verified human_authored lore; retrieve by id; dual-write + audit + outbox in one UoW; remember/POST lore-entries unchanged

**Independent Test**: Create with question, choice, owner, one alternative; get returns those fields and source human

### Tests for User Story 1

- [x] T010 [P] [US1] Failing CreateDecisionHandler tests in internal/application/commands/create_decision_test.go (required fields, alternatives, same id as lore, verified human_authored, outbox, blank actor, non-repository scope)
- [x] T011 [P] [US1] Failing GetDecisionHandler tests in internal/application/queries/get_decision_test.go (human row by id; unknown id not found)

### Implementation for User Story 1

- [x] T012 [US1] Implement CreateDecisionHandler in internal/application/commands/create_decision.go (GREEN T010)
- [x] T013 [US1] Implement GetDecisionHandler in internal/application/queries/get_decision.go for human rows (GREEN T011; ADR projection comes in US2)

**Checkpoint**: Human create+get works against memory UoW

---

## Phase 4: User Story 2 - Trusted ADR lore is queryable as decisions (Priority: P1)

**Goal**: List-current unions human Decisions and ADR projections; git/PR/review items excluded; ADR appears once; get-by-id of lore id returns projected Decision

**Independent Test**: Seed ADR lore + human Decision + git observation + pending PR → list returns exactly two

### Tests for User Story 2

- [x] T014 [P] [US2] Failing ListDecisionsHandler tests in internal/application/queries/list_decisions_test.go (human+ADR once; git/PR excluded; superseded ADR excluded)
- [x] T015 [US2] Extend GetDecisionHandler tests for ADR projection and historical ADR in internal/application/queries/get_decision_test.go

### Implementation for User Story 2

- [x] T016 [US2] Implement ListDecisionsHandler in internal/application/queries/list_decisions.go (GREEN T014)
- [x] T017 [US2] Complete GetDecisionHandler ADR projection in internal/application/queries/get_decision.go (GREEN T015)

**Checkpoint**: ADR corpus is queryable as Decisions without a second current fact

---

## Phase 5: User Story 3 - Supersede without deleting history (Priority: P1)

**Goal**: New current successor; predecessor gettable; ADR may be superseded by a human successor; dual-write in one UoW

**Independent Test**: Supersede human A→B; list current is B; get A is historical. Supersede ADR→human; old ADR not a second current fact

### Tests for User Story 3

- [x] T018 [US3] Failing SupersedeDecisionHandler tests in internal/application/commands/supersede_decision_test.go (human chain, ADR predecessor, reject already superseded/invalidated, outbox, blank actor)

### Implementation for User Story 3

- [x] T019 [US3] Implement SupersedeDecisionHandler in internal/application/commands/supersede_decision.go (GREEN T018)

**Checkpoint**: History preserved; list-current is the successor only

---

## Phase 6: User Story 4 - Decisions feed compile / get_for_task (Priority: P1)

**Goal**: First-class Decisions appear in existing `decisions` section; ranking formulas unchanged; characterization vs leftover observations and F032 ADR authority

**Independent Test**: Human Decision + ADR + git observation → both decisions in section `decisions`, observations outranked, ADR vs observation ranking not weakened

### Tests for User Story 4

- [x] T020 [P] [US4] Failing ClassifyItem FirstClassDecision tests in internal/application/context/profile_test.go
- [x] T021 [US4] Failing compile characterization tests in internal/application/queries/compile_context_test.go (human Decision outranks unverified observation; ingested ADR still outranks git/PR; no duplicate id in decisions section)

### Implementation for User Story 4

- [x] T022 [US4] Add RankedItem.FirstClassDecision and ClassifyItem handling in internal/application/context/ranking.go and internal/application/context/profile.go (GREEN T020)
- [x] T023 [US4] Mark first-class Decision items during compile in internal/application/queries/compile_context.go (GREEN T021)

**Checkpoint**: Packet `decisions` is fed by the F040 model without reopening F007 formulas

---

## Phase 7: User Story 5 - CLI + REST; MCP stays at 10 (Priority: P2)

**Goal**: `memlore decision create|get|list|supersede`; REST `/v1/decisions`; membership 403/404; MCP still exactly 10 tools; ingest/review unchanged

**Independent Test**: CLI + REST contract + membership contract + MCP tool count

### Tests for User Story 5

- [x] T024 [P] [US5] Failing REST contract tests in internal/adapters/http/decision_contract_test.go (create 201, get, list, supersede, remember still lore-only)
- [x] T025 [P] [US5] Failing membership contract tests in internal/adapters/http/decision_membership_contract_test.go (403 list/create without membership; 404 cross-tenant get)
- [x] T026 [P] [US5] Failing CLI tests in internal/adapters/cli/decision_test.go and cmd/memlore/main_test.go
- [x] T027 [US5] Confirm TestListToolsIsExactlyTenLoreTools still passes in internal/adapters/mcp/mcp_contract_test.go (no new tool)

### Implementation for User Story 5

- [x] T028 [US5] Implement REST handlers and routes in internal/adapters/http/decision.go and internal/adapters/http/handlers.go (GREEN T024)
- [x] T029 [US5] Wire membership gates so T025 passes
- [x] T030 [US5] Implement CLI parse/format in internal/adapters/cli/decision.go and wire memlore decision in cmd/memlore/main.go (GREEN T026)

**Checkpoint**: Operator surfaces complete; agents do not mutate Decisions

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Docs, tracker, compile characterization leftover, green suite

- [x] T031 [P] Document REST routes in docs/api/rest.md
- [x] T032 [P] Mark F040 DONE in docs/development/FEATURE_DEVELOPMENT.md; update next-step pointers; update docs/architecture/target-architecture.md if the decision aggregate is listed; contributing.md / README.md as needed
- [x] T033 Update .cursor/rules/specify-rules.mdc Active feature plan to F040 DONE
- [x] T034 Assert F035 Accept does not create a Decision (test in internal/application/commands/accept_review_test.go or list_decisions_test.go)
- [x] T035 Run `go test ./...` and `go vet ./...` until green

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1**: After Foundational
- **US2**: After US1 get handler exists (extends get)
- **US3**: After US1 create + US2 get/list projection
- **US4**: After US1+US2 (needs Decision ids in compile)
- **US5**: After US1–US3 handlers exist (adapters wrap them); US4 can proceed in parallel with US5
- **Polish**: After US1–US5

### User Story Dependencies

- **User Story 1 (P1)**: Foundational only — MVP
- **User Story 2 (P1)**: Needs GetDecisionHandler from US1
- **User Story 3 (P1)**: Needs create + get + list
- **User Story 4 (P1)**: Needs current Decisions in store
- **User Story 5 (P2)**: Needs application handlers

### Parallel Opportunities

- T003/T004 tests in parallel
- T010/T011 tests in parallel
- T014/T015 tests in parallel
- T020/T021 tests in parallel
- T024/T025/T026 tests in parallel
- T031/T032 docs in parallel

### MVP

Phase 1 + 2 + User Story 1 (human create/get) is the smallest demo. Full F040 DONE requires US1–US5 plus polish.

## Implementation Strategy

1. Foundational domain + ports + memory repo so tests can run without Postgres
2. US1 create/get (MVP)
3. US2 ADR projection list/get
4. US3 supersede
5. US4 compile flag
6. US5 REST/CLI
7. Docs + `go test ./...` / `go vet ./...`

Do not commit unless asked.
