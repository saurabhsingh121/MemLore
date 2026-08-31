# Tasks: Git Commit Ingestion (F030)

**Input**: Design documents from `/specs/030-git-commit-ingest/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/git-commit-ingest.md

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

- [x] T001 Confirm feature artifacts exist in specs/030-git-commit-ingest/ (spec.md, plan.md, research.md, data-model.md, contracts/git-commit-ingest.md, quickstart.md)
- [x] T002 [P] Verify .specify/feature.json points at specs/030-git-commit-ingest

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Evidence type, observational lore constructor, ingest ports/schema, git reader port — required before any story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Add failing tests for EvidenceType `commit` in internal/domain/enums_test.go
- [x] T004 [P] Add failing tests for NewObservationalLoreEntry (origin, unverified, require commit evidence; NewLoreEntry still rejects non-human) in internal/domain/lore_test.go
- [x] T005 [P] Add ConflictError (409) in internal/domain/errors.go with tests in internal/domain/errors_test.go if needed
- [x] T006 Implement EvidenceTypeCommit in internal/domain/enums.go (GREEN T003)
- [x] T007 Implement NewObservationalLoreEntry in internal/domain/lore.go (GREEN T004)
- [x] T008 Add GitCommitSnapshot + ingest run/cursor/processed-SHA domain types in internal/domain/ingest.go with tests in internal/domain/ingest_test.go
- [x] T009 Add GitReader port in internal/application/ports/git.go and ingest repository ports in internal/application/ports/ingest.go; extend UnitOfWork in internal/application/ports/repositories.go
- [x] T010 Add goose migration migrations/00005_git_ingest.sql and sqlc queries db/queries/ingest.sql; regenerate sqlc under internal/infrastructure/postgres/sqlc/
- [x] T011 Wire ingest repos on postgres UnitOfWork in internal/infrastructure/postgres/unit_of_work.go and memory UnitOfWork in internal/infrastructure/memory/repositories.go (plus ingest memory store)

**Checkpoint**: Domain can represent observational lore + commit evidence; UoW can persist ingest metadata in tests

---

## Phase 3: User Story 1 - Conservative commit capture as observational lore (Priority: P1) 🎯 MVP

**Goal**: Extract at most one observational unverified candidate per rationale SHA; skip noisy commits; evidence includes commit SHA; outbox like other creates

**Independent Test**: Fixture with rationale + chore + merge → one observational unverified candidate with commit evidence; origin not human_authored

### Tests for User Story 1

- [x] T012 [P] [US1] Failing extractor tests (skip merge/chore/empty; keep because/migration; no invent) in internal/application/ingest/extract_test.go
- [x] T013 [US1] Failing IngestGitCommitsHandler tests with fake GitReader in internal/application/commands/ingest_git_test.go (one candidate, skip noisy, origin/status/evidence, outbox event)

### Implementation for User Story 1

- [x] T014 [US1] Implement extractor in internal/application/ingest/extract.go (GREEN T012)
- [x] T015 [US1] Implement IngestGitCommitsHandler in internal/application/commands/ingest_git.go writing observational lore + audit + outbox + processed SHA per candidate (GREEN T013)
- [x] T016 [US1] Implement git CLI adapter with temp-repo test in internal/infrastructure/gitcli/reader.go and reader_test.go

**Checkpoint**: Application ingest produces observational lore from a fake or real git log; noisy SHAs skipped

---

## Phase 4: User Story 2 - Idempotent re-ingest and safe retry (Priority: P1)

**Goal**: Same SHA twice → one candidate; skipped SHAs stay skipped; retry after partial failure does not duplicate; cursor incremental

**Independent Test**: Ingest twice; count unchanged. Fail-after-one then retry → still one row for that SHA

### Tests for User Story 2

- [x] T017 [US2] Failing tests for re-ingest idempotency, skipped-SHA sticky, cursor increment, concurrent running conflict in internal/application/commands/ingest_git_test.go

### Implementation for User Story 2

- [x] T018 [US2] Persist git_ingest_shas unique key and skip already-processed SHAs in internal/application/commands/ingest_git.go and memory/postgres ingest repos
- [x] T019 [US2] Persist/update git_ingest_cursors; reject second running run (ConflictError) in internal/application/commands/ingest_git.go
- [x] T020 [US2] Per-SHA unit of work so a later failure does not roll back earlier candidates (GREEN T017)

**Checkpoint**: Re-run and retry are safe; one active run per repository

---

## Phase 5: User Story 3 - Operators trigger ingest and inspect status (Priority: P2)

**Goal**: CLI ingest git/status; REST trigger/list runs/get run/list candidates; membership deny; no new MCP tool

**Independent Test**: CLI+REST on fixture; 403 cross-tenant; MCP tool count 10

### Tests for User Story 3

- [x] T021 [P] [US3] Failing CLI parse/format tests in internal/adapters/cli/ingest_test.go
- [x] T022 [P] [US3] Failing REST contract tests (POST /v1/ingest/git, GET runs, GET candidates, 409, validation) in internal/adapters/http/ingest_contract_test.go
- [x] T023 [P] [US3] Failing membership deny contract test in internal/adapters/http/ingest_membership_contract_test.go
- [x] T024 [P] [US3] Failing query tests for list runs/candidates in internal/application/queries/ingest_status_test.go

### Implementation for User Story 3

- [x] T025 [US3] Implement ListIngestRuns / GetIngestRun / ListIngestCandidates in internal/application/queries/ingest_status.go (GREEN T024)
- [x] T026 [US3] CLI parse + format in internal/adapters/cli/ingest.go (GREEN T021)
- [x] T027 [US3] Wire memlore ingest git/status in cmd/memlore/main.go and cmd/memlore/main_test.go
- [x] T028 [US3] REST handlers + DTOs + ConflictError→409 in internal/adapters/http/handlers.go, dto.go, and new ingest handler file (GREEN T022)
- [x] T029 [US3] Membership gate on ingest routes (GREEN T023); confirm MCP tool count unchanged in existing mcp_contract_test.go

**Checkpoint**: Operators can trigger and inspect ingest; agents cannot promote via MCP

---

## Phase 6: User Story 4 - Trust boundary: compile still prefers verified architecture (Priority: P2)

**Goal**: Characterization — verified architecture outranks unverified git observation; ingest does not auto-verify

**Independent Test**: Seed both; compile; architecture first; git origin remains repository_observation / unverified

### Tests for User Story 4

- [x] T030 [US4] Failing compile ranking test with observational git lore vs verified architecture in internal/application/queries/compile_context_test.go (or ingest-specific test file)
- [x] T031 [US4] Assert ingest handler never sets verified / human_authored / human_verified in internal/application/commands/ingest_git_test.go

### Implementation for User Story 4

- [x] T032 [US4] GREEN T030 using existing EvaluateAuthority (no ranking formula change). Fix only if ingest origin/status is wrong
- [x] T033 [US4] Keep POST /v1/lore-entries human-authored (existing contract test must still pass)

**Checkpoint**: Capture does not look canonical in get_for_task ranking

---

## Phase 7: Polish & Cross-Cutting Concerns

- [x] T034 [P] Document REST ingest routes in docs/api/rest.md
- [x] T035 [P] Mark F030 DONE in docs/development/FEATURE_DEVELOPMENT.md; next F031 or F035; update Immediate recommended tasks
- [x] T036 [P] Update docs/development/contributing.md (next is no longer F021)
- [x] T037 [P] Update README.md ingest CLI line if usage is listed
- [x] T038 Structured slog on ingest start/complete/fail in cmd/memlore/main.go and/or ingest handler
- [x] T039 Run go test ./... and go vet ./...; fix regressions
- [x] T040 Mark spec.md Status Implemented

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Immediate
- **Foundational (Phase 2)**: Blocks all user stories
- **US1**: After foundational — MVP
- **US2**: After US1 (extends the same handler)
- **US3**: After US1 (needs handler); status queries can start after T011
- **US4**: After US1 (needs observational lore in compile)
- **Polish**: After stories

### User Story Dependencies

- **User Story 1 (P1)**: After Phase 2
- **User Story 2 (P1)**: After US1 (same command)
- **User Story 3 (P2)**: After US1
- **User Story 4 (P2)**: After US1; independent of US3

### Parallel Opportunities

- T003–T005 tests in parallel
- T021–T024 tests in parallel once handler exists
- T034–T037 docs in parallel

### Parallel Example: User Story 1

```bash
Task: "Failing extractor tests in internal/application/ingest/extract_test.go"
Task: "Failing IngestGitCommitsHandler tests in internal/application/commands/ingest_git_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1–2 foundation
2. Extractor + ingest handler
3. Validate: one candidate from mixed fixture

### Incremental Delivery

1. US1 capture → US2 idempotency → US3 CLI/REST → US4 compile characterization → docs/vet

### Suggested MVP scope

US1 + US2 (capture that is safe to re-run). US3 is required for the product surface before DONE.
