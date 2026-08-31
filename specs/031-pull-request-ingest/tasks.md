# Tasks: Pull Request Ingestion (F031)

**Input**: Design documents from `/specs/031-pull-request-ingest/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/pull-request-ingest.md

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

- [x] T001 Confirm feature artifacts exist in specs/031-pull-request-ingest/ (spec.md, plan.md, research.md, data-model.md, contracts/pull-request-ingest.md, quickstart.md)
- [x] T002 [P] Verify .specify/feature.json points at specs/031-pull-request-ingest

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Evidence type `pr`, observational lore accepts PR evidence, PR ingest ports/schema, PullRequestReader port — required before any story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Add failing tests for EvidenceType `pr` in internal/domain/enums_test.go
- [x] T004 [P] Add failing tests for NewObservationalLoreEntry accepting `pr` evidence (commit still works; neither fails; NewLoreEntry still rejects non-human) in internal/domain/lore_test.go
- [x] T005 [P] Add GitHub scope-key parser tests and PR evidence value tests in internal/domain/ingest_test.go
- [x] T006 Implement EvidenceTypePR in internal/domain/enums.go (GREEN T003)
- [x] T007 Extend NewObservationalLoreEntry in internal/domain/lore.go to require commit or pr evidence (GREEN T004)
- [x] T008 Add PullRequestSnapshot, PRIngestRun, PRIngestCursor, ProcessedPR, skip reasons, GitHubRepoFromScopeKey in internal/domain/ingest.go (GREEN T005)
- [x] T009 Add PullRequestReader port in internal/application/ports/github.go and PRIngestRepository in internal/application/ports/ingest.go; extend UnitOfWork in internal/application/ports/repositories.go
- [x] T010 Add goose migration migrations/00006_pr_ingest.sql and sqlc queries db/queries/pr_ingest.sql; regenerate sqlc under internal/infrastructure/postgres/sqlc/
- [x] T011 Wire PR ingest repos on postgres UnitOfWork in internal/infrastructure/postgres/unit_of_work.go and memory UnitOfWork in internal/infrastructure/memory/repositories.go (plus memory PR ingest store)

**Checkpoint**: Domain can represent observational lore + PR evidence; UoW can persist PR ingest metadata in tests

---

## Phase 3: User Story 1 - Conservative PR capture as observational lore (Priority: P1) 🎯 MVP

**Goal**: Extract at most one observational unverified candidate per rationale merged PR; skip unmerged/bots/noisy; evidence includes `pr`; outbox like other creates

**Independent Test**: Fixture with rationale merged + dependabot + unmerged → one observational unverified candidate with PR evidence; origin not human_authored

### Tests for User Story 1

- [x] T012 [P] [US1] Failing PR extractor tests (skip unmerged/bot/chore; keep because/migration; include used review comment urls; no invent) in internal/application/ingest/extract_pr_test.go
- [x] T013 [US1] Failing IngestPullRequestsHandler tests with fake PullRequestReader in internal/application/commands/ingest_pr_test.go (one candidate, skip noisy/unmerged, origin/status/evidence, outbox event)

### Implementation for User Story 1

- [x] T014 [US1] Implement ExtractPRCandidate in internal/application/ingest/extract_pr.go (GREEN T012)
- [x] T015 [US1] Implement IngestPullRequestsHandler in internal/application/commands/ingest_pr.go writing observational lore + audit + outbox + processed PR per candidate (GREEN T013)
- [x] T016 [US1] Implement GitHub REST adapter with httptest fixtures in internal/infrastructure/githubhttp/reader.go and reader_test.go

**Checkpoint**: Application ingest produces observational lore from a fake or httptest GitHub; noisy/unmerged PRs skipped

---

## Phase 4: User Story 2 - Idempotent re-ingest and safe retry (Priority: P1)

**Goal**: Same PR twice → one candidate; skipped PRs stay skipped; retry after partial failure does not duplicate; cursor incremental; --pr single number

**Independent Test**: Ingest twice; count unchanged. Fail-after-one then retry → still one row for that PR

### Tests for User Story 2

- [x] T017 [US2] Failing tests for re-ingest idempotency, skipped-PR sticky, cursor increment, concurrent running conflict, single --pr in internal/application/commands/ingest_pr_test.go

### Implementation for User Story 2

- [x] T018 [US2] Implement processed-PR unique key, cursor watermark, one-active-run conflict, and single-PR path in internal/application/commands/ingest_pr.go and infrastructure stores (GREEN T017)

**Checkpoint**: Re-ingest and retry are safe; 409 when a PR ingest run is already running

---

## Phase 5: User Story 3 - Operators trigger PR ingest and inspect status (Priority: P2)

**Goal**: CLI ingest pr / status --kind pr; REST trigger/list PR runs/get run/list PR candidates; membership deny; git routes still work; no new MCP tool

**Independent Test**: CLI parse/format; REST contract; membership 403; git contract tests still pass

### Tests for User Story 3

- [x] T019 [P] [US3] Failing CLI parse/format tests in internal/adapters/cli/ingest_test.go
- [x] T020 [P] [US3] Failing REST contract tests in internal/adapters/http/ingest_pr_contract_test.go
- [x] T021 [P] [US3] Failing membership 403 contract in internal/adapters/http/ingest_pr_membership_contract_test.go
- [x] T022 [P] [US3] Failing candidate evidence_type=pr filter tests in internal/application/queries/ingest_status_test.go

### Implementation for User Story 3

- [x] T023 [US3] Implement CLI ingest pr + status --kind in internal/adapters/cli/ingest.go (GREEN T019)
- [x] T024 [US3] Implement REST POST /v1/ingest/pr, GET /v1/ingest/pr-runs, GET /v1/ingest/pr-runs/{id}, candidates evidence_type filter in internal/adapters/http/ and handlers.go (GREEN T020, T022)
- [x] T025 [US3] Membership deny for PR ingest in internal/adapters/http (GREEN T021)
- [x] T026 [US3] Wire memlore ingest pr/status --kind in cmd/memlore/main.go and cmd/memlore/main_test.go
- [x] T027 [US3] Confirm MCP tool count remains 10 (existing MCP tests)

**Checkpoint**: Operators can trigger and inspect PR ingest; git ingest unchanged; MCP still 10 tools

---

## Phase 6: User Story 4 - Trust boundary (Priority: P2)

**Goal**: Compile still ranks verified architecture above PR-derived observations; ingest does not auto-verify

**Independent Test**: Characterization test with verified architecture + PR observation

- [x] T028 [US4] Add TestCompileContextVerifiedArchitectureOutranksPRObservation in internal/application/queries/compile_context_test.go
- [x] T029 [US4] Assert ingested PR lore remains unverified in ingest handler tests (no extra verify call)

**Checkpoint**: Capture does not look like canonical architecture

---

## Phase 7: Polish & Cross-Cutting Concerns

- [x] T030 [P] Update docs/api/rest.md with PR ingest routes
- [x] T031 [P] Mark F031 DONE in docs/development/FEATURE_DEVELOPMENT.md; next F032 or F035; update Immediate recommended tasks
- [x] T032 [P] Update .cursor/rules/specify-rules.mdc, docs/development/contributing.md, README.md, docs/architecture/overview.md
- [x] T033 Run go test ./... and go vet ./...; fix failures
- [x] T034 Confirm F030 git contract tests still pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Story 1 (P1)**: Depends on Foundational
- **User Story 2 (P1)**: Depends on US1 handler skeleton
- **User Story 3 (P2)**: Depends on US1/US2 command
- **User Story 4 (P2)**: Independent characterization once domain origin exists (can run after T007)
- **Polish**: After desired stories

### Parallel Opportunities

- T003–T005 tests in parallel
- T012 extractor tests parallel with T013 once ports exist
- T019–T022 contract tests in parallel
- T030–T032 docs in parallel

### MVP

Phase 1 + 2 + User Story 1 (conservative capture + observational persist + outbox)

---

## Notes

- Do not implement F032, F035, F054, F074
- Do not add an 11th MCP tool
- Do not log GitHub tokens
- Do not overload git_ingest_shas
