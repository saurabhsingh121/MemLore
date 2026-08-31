# Tasks: Suggested Lore Review Queue (F035)

**Input**: Design documents from `/specs/035-suggested-lore-review/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/review-queue.md

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

- [x] T001 Confirm feature artifacts exist in specs/035-suggested-lore-review/ (spec.md, plan.md, research.md, data-model.md, contracts/review-queue.md, quickstart.md)
- [x] T002 [P] Verify .specify/feature.json points at specs/035-suggested-lore-review

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: human_verified constructor, extract identity, review decision types, ports/schema — required before any story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Add failing tests for NewHumanVerifiedLoreEntry (requires evidence, origin human_verified, verified + VerifiedBy; NewLoreEntry still rejects non-human; observational still commit/pr) in internal/domain/lore_test.go
- [x] T004 [P] Add failing tests for extract identity (commit preferred, else pr, statement checksum), eligibility, AcceptSuggestedLore as-stated vs edit, reject overlay rules in internal/domain/review_test.go
- [x] T005 Implement NewHumanVerifiedLoreEntry and verified human_authored helper in internal/domain/lore.go (GREEN T003)
- [x] T006 Implement ReviewDecision, ExtractIdentity, AcceptSuggestedLore, same-statement rule in internal/domain/review.go (GREEN T004)
- [x] T007 Add ReviewDecisionRepository port in internal/application/ports/review.go; extend UnitOfWork in internal/application/ports/repositories.go
- [x] T008 Add goose migration migrations/00008_lore_review.sql and sqlc queries db/queries/review.sql; regenerate sqlc under internal/infrastructure/postgres/sqlc/
- [x] T009 Wire ReviewDecision repos on postgres UnitOfWork in internal/infrastructure/postgres/unit_of_work.go and memory UnitOfWork in internal/infrastructure/memory/repositories.go

**Checkpoint**: Domain can represent Accept successors and reject identities; UoW can persist review decisions in tests

---

## Phase 3: User Story 1 - Reviewers see pending suggested lore (Priority: P1) 🎯 MVP

**Goal**: List pending git/PR observational items with statement, evidence, source type; omit confidence/reason; exclude F032 ADRs and rejected/accepted extracts; membership-scoped

**Independent Test**: Seed git + PR + ADR + rejected → list returns exactly two pending items

### Tests for User Story 1

- [x] T010 [P] [US1] Failing ListReviewQueueHandler tests in internal/application/queries/list_review_queue_test.go (git+pr pending; ADR excluded; rejected excluded; no confidence/reason)

### Implementation for User Story 1

- [x] T011 [US1] Implement ListReviewQueueHandler in internal/application/queries/list_review_queue.go (GREEN T010)

**Checkpoint**: Application query projects the pending queue from lore + decisions

---

## Phase 4: User Story 2 - Accept-as-stated with human verification provenance (Priority: P1)

**Goal**: Accept creates human_verified verified successor with copied evidence; predecessor superseded; outbox in one UoW; verify-in-place unchanged

**Independent Test**: Accept git observation → new current human_verified row; predecessor superseded observational; verify still does not change origin

### Tests for User Story 2

- [x] T012 [US2] Failing AcceptReviewHandler tests in internal/application/commands/accept_review_test.go (as-stated origin/evidence/verified, supersede, outbox, idempotent re-accept, reject ADR/human, blank actor)

### Implementation for User Story 2

- [x] T013 [US2] Implement AcceptReviewHandler in internal/application/commands/accept_review.go writing successor + supersede + audits + outbox + decision (GREEN T012)

**Checkpoint**: Accept-as-stated is the human_verified writer; verify is not Accept

---

## Phase 5: User Story 3 - Edit then Accept (Priority: P1)

**Goal**: Different statement → human_authored verified successor; evidence copied; predecessor statement unchanged

**Independent Test**: Accept with statement B; current is B human_authored; predecessor still A observational superseded

### Tests for User Story 3

- [x] T014 [US3] Failing edit-then-accept and same-statement-is-as-stated tests in internal/application/commands/accept_review_test.go

### Implementation for User Story 3

- [x] T015 [US3] Handle optional replacement statement in AcceptReviewHandler (GREEN T014)

**Checkpoint**: Edit provenance is human_authored; as-stated stays human_verified

---

## Phase 6: User Story 4 - Reject records a durable negative decision (Priority: P1)

**Goal**: Reject removes from pending; observation not deleted; re-ingest does not resurrect; idempotent re-reject; accept-after-reject fails

**Independent Test**: Reject then list pending absent; observational row remains; second reject no-ops; accept fails

### Tests for User Story 4

- [x] T016 [US4] Failing RejectReviewHandler tests in internal/application/commands/reject_review_test.go (pending gone, lore intact, idempotent reject, accept-after-reject, reject-after-accept)

### Implementation for User Story 4

- [x] T017 [US4] Implement RejectReviewHandler in internal/application/commands/reject_review.go (GREEN T016)
- [x] T018 [US4] Ensure list query hides rejected identities (GREEN T010 + T016 together)

**Checkpoint**: Negative decisions stick; observational history remains

---

## Phase 7: User Story 5 - CLI + REST operators; MCP unchanged (Priority: P2)

**Goal**: CLI review list/accept/reject; REST list/get/accept/reject; membership deny; ingest routes still work; no new MCP tool

**Independent Test**: CLI parse/format; REST contract; membership 403; MCP still 10; git/PR/ADR contracts still pass

### Tests for User Story 5

- [x] T019 [P] [US5] Failing CLI parse/format tests in internal/adapters/cli/review_test.go
- [x] T020 [P] [US5] Failing REST contract tests in internal/adapters/http/review_contract_test.go
- [x] T021 [P] [US5] Failing membership 403 contract in internal/adapters/http/review_membership_contract_test.go

### Implementation for User Story 5

- [x] T022 [US5] Implement CLI review list/accept/reject in internal/adapters/cli/review.go (GREEN T019)
- [x] T023 [US5] Implement REST GET /v1/review-queue, GET /v1/review-queue/{id}, POST accept/reject in internal/adapters/http/review.go and handlers.go (GREEN T020)
- [x] T024 [US5] Membership deny for review routes (GREEN T021)
- [x] T025 [US5] Wire memlore review in cmd/memlore/main.go and cmd/memlore/main_test.go
- [x] T026 [US5] Confirm MCP tool count remains 10 (existing MCP tests)

**Checkpoint**: Operators can list/accept/reject; ingest unchanged; MCP still 10 tools

---

## Phase 8: Trust boundary characterization

**Goal**: Accepted review-queue lore outranks leftover unverified observations; F032 ADR ranking unchanged; formulas frozen

- [x] T027 Add TestCompileContextAcceptedReviewOutranksUnverifiedObservation in internal/application/queries/compile_context_test.go
- [x] T028 Confirm TestCompileContextIngestedADROutranksGitAndPRObservation still passes

**Checkpoint**: Promotion raises authority without beating ADR evidence strength accidentally via formula edits

---

## Phase 9: Polish & Cross-Cutting Concerns

- [x] T029 [P] Update docs/api/rest.md with review-queue routes
- [x] T030 [P] Mark F035 DONE in docs/development/FEATURE_DEVELOPMENT.md; next F040 or F022; update Immediate recommended tasks
- [x] T031 [P] Update .cursor/rules/specify-rules.mdc, docs/development/contributing.md, README.md as needed
- [x] T032 Run go test ./... and go vet ./...; fix failures
- [x] T033 Confirm F030 git, F031 PR, and F032 ADR contract tests still pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Story 1 (P1)**: Depends on Foundational (list needs decisions + eligibility)
- **User Story 2 (P1)**: Depends on Foundational; list used to prove pending is gone after accept
- **User Story 3 (P1)**: Depends on US2 handler skeleton
- **User Story 4 (P1)**: Depends on Foundational; list must hide rejects
- **User Story 5 (P2)**: Depends on US1–US4 commands/queries
- **Characterization**: After human_verified origin exists (after T005/T013)
- **Polish**: After desired stories

### Parallel Opportunities

- T003–T004 tests in parallel
- T019–T021 contract tests in parallel
- T029–T031 docs in parallel

### MVP

Phase 1 + 2 + User Story 1 (pending list) + User Story 2 (Accept-as-stated)

---

## Notes

- Do not implement F033, F034, F040, F022, F120
- Do not add an 11th MCP tool
- Do not loosen NewLoreEntry
- Do not overload git/PR/ADR ingest tables
- Do not change compile ranking formulas
- Do not invent confidence or reason
- Do not treat verify as Accept
- Do not undo F032 trusted-source auto-verify
