# Tasks: ADR Auto-Ingestion (F032)

**Input**: Design documents from `/specs/032-adr-ingest/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/adr-ingest.md

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

- [x] T001 Confirm feature artifacts exist in specs/032-adr-ingest/ (spec.md, plan.md, research.md, data-model.md, contracts/adr-ingest.md, quickstart.md)
- [x] T002 [P] Verify .specify/feature.json points at specs/032-adr-ingest

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Trusted-source constructor, ADR ingest domain types, ports/schema, ADRReader — required before any story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Add failing tests for NewArchitectureDecisionLoreEntry (requires adr evidence, origin architecture_decision, verified + VerifiedBy; NewLoreEntry still rejects non-human; observational still requires commit/pr not adr) in internal/domain/lore_test.go
- [x] T004 [P] Add failing tests for ADR snapshot, ADRIngestRun, ProcessedADR, skip reasons, default dirs in internal/domain/ingest_test.go
- [x] T005 [P] Add failing tests for ApplySupersessionWithSuccessor (pre-built architecture_decision successor; does not call NewLoreEntry; human predecessor still works via existing ApplySupersession) in internal/domain/lifecycle_test.go
- [x] T006 Implement NewArchitectureDecisionLoreEntry in internal/domain/lore.go (GREEN T003)
- [x] T007 Add ADRSnapshot, ADRIngestRun, ADRIngestCursor, ProcessedADR, skip reasons, DefaultADRDirs in internal/domain/ingest.go (GREEN T004)
- [x] T008 Implement ApplySupersessionWithSuccessor in internal/domain/lifecycle.go (GREEN T005)
- [x] T009 Add ADRReader port in internal/application/ports/adr.go and ADRIngestRepository in internal/application/ports/ingest.go; extend UnitOfWork in internal/application/ports/repositories.go
- [x] T010 Add goose migration migrations/00007_adr_ingest.sql and sqlc queries db/queries/adr_ingest.sql; regenerate sqlc under internal/infrastructure/postgres/sqlc/
- [x] T011 Wire ADR ingest repos on postgres UnitOfWork in internal/infrastructure/postgres/unit_of_work.go and memory UnitOfWork in internal/infrastructure/memory/repositories.go (plus memory ADR ingest store)

**Checkpoint**: Domain can represent trusted-source ADR lore; UoW can persist ADR ingest metadata in tests

---

## Phase 3: User Story 1 - Trusted-source capture of accepted ADRs (Priority: P1) 🎯 MVP

**Goal**: Extract at most one verified architecture_decision lore entry per accepted ADR; skip draft/README/template/uncertain; evidence type `adr`; outbox like other creates; git/PR lore not upgraded

**Independent Test**: Fixture with accepted + draft + README + template → one verified architecture_decision lore with adr evidence; origin not human_authored or repository_observation

### Tests for User Story 1

- [x] T012 [P] [US1] Failing ADR parser tests (accepted MADR/Nygard; skip draft/README/template/NNNN-title; no invent; front-matter status) in internal/application/ingest/extract_adr_test.go
- [x] T013 [US1] Failing IngestADRsHandler tests with fake ADRReader in internal/application/commands/ingest_adr_test.go (one verified lore, skip draft/template, origin/status/evidence, outbox event, git/PR untouched)

### Implementation for User Story 1

- [x] T014 [US1] Implement ExtractADR in internal/application/ingest/extract_adr.go (GREEN T012)
- [x] T015 [US1] Implement IngestADRsHandler in internal/application/commands/ingest_adr.go writing trusted-source lore + audit + outbox + processed ADR per accepted file (GREEN T013)
- [x] T016 [US1] Implement local filesystem ADRReader with t.TempDir fixtures in internal/infrastructure/fsadr/reader.go and reader_test.go

**Checkpoint**: Application ingest produces verified architecture-decision lore from a fake or temp-dir tree; drafts/templates skipped

---

## Phase 4: User Story 2 - Idempotent re-ingest, content change, and safe retry (Priority: P1)

**Goal**: Same ADR twice → one current lore; checksum change supersedes prior ingest lore; retry after partial failure does not duplicate; extra --adr-dir scanned; one-active-run conflict

**Independent Test**: Ingest twice; count unchanged. Change file; new current + predecessor superseded. Fail-after-one then retry → still one current row for that ADR

### Tests for User Story 2

- [x] T017 [US2] Failing tests for re-ingest idempotency, checksum-change supersession, skipped-draft sticky, extra dirs, concurrent running conflict, missing path failed run in internal/application/commands/ingest_adr_test.go

### Implementation for User Story 2

- [x] T018 [US2] Implement processed (path, checksum) unique key, checksum-change supersede, extra dirs, one-active-run conflict, and missing-path failed run in internal/application/commands/ingest_adr.go and infrastructure stores (GREEN T017)

**Checkpoint**: Re-ingest and retry are safe; 409 when an ADR ingest run is already running; history not overwritten

---

## Phase 5: User Story 3 - ADR supersession maps to lore supersession (Priority: P2)

**Goal**: “Supersedes ADR-0003” chains ingest-created lore; human-authored not auto-superseded; deprecated/superseded files stored then invalidated; conflicts without a link left alone

**Independent Test**: Ingest 0003 then 0007-supersedes-0003; 0007 current, 0003 ingest lore superseded. Human-authored entry remains current.

### Tests for User Story 3

- [x] T019 [US3] Failing tests for explicit supersedes mapping, skip human-authored predecessor, deprecated-file invalidation in internal/application/commands/ingest_adr_test.go and parser tests in extract_adr_test.go

### Implementation for User Story 3

- [x] T020 [US3] Parse Supersedes links and apply ingest-only supersession / historical invalidation in extract_adr.go and ingest_adr.go (GREEN T019)

**Checkpoint**: Constitution VI mapping works; human ADRs left for humans

---

## Phase 6: User Story 4 - Operators trigger ADR ingest and inspect status (Priority: P2)

**Goal**: CLI ingest adr / status --kind adr; REST trigger/list ADR runs/get run/list ADR candidates; membership deny; git/PR routes still work; no new MCP tool

**Independent Test**: CLI parse/format; REST contract; membership 403; git and PR contract tests still pass

### Tests for User Story 4

- [x] T021 [P] [US4] Failing CLI parse/format tests in internal/adapters/cli/ingest_test.go
- [x] T022 [P] [US4] Failing REST contract tests in internal/adapters/http/ingest_adr_contract_test.go
- [x] T023 [P] [US4] Failing membership 403 contract in internal/adapters/http/ingest_adr_membership_contract_test.go
- [x] T024 [P] [US4] Failing candidate evidence_type=adr filter tests in internal/application/queries/ingest_status_test.go

### Implementation for User Story 4

- [x] T025 [US4] Implement CLI ingest adr + status --kind adr in internal/adapters/cli/ingest.go (GREEN T021)
- [x] T026 [US4] Implement REST POST /v1/ingest/adr, GET /v1/ingest/adr-runs, GET /v1/ingest/adr-runs/{id}, candidates evidence_type=adr in internal/adapters/http/ and handlers.go (GREEN T022, T024)
- [x] T027 [US4] Membership deny for ADR ingest in internal/adapters/http (GREEN T023)
- [x] T028 [US4] Wire memlore ingest adr/status --kind in cmd/memlore/main.go and cmd/memlore/main_test.go
- [x] T029 [US4] Confirm MCP tool count remains 10 (existing MCP tests)

**Checkpoint**: Operators can trigger and inspect ADR ingest; git/PR ingest unchanged; MCP still 10 tools

---

## Phase 7: User Story 5 - Trust boundary (Priority: P2)

**Goal**: Compile ranks ingested accepted ADR lore above git/PR observations; F032 does not change ranking formulas; git/PR stay observational unverified

**Independent Test**: Characterization test with ingested verified architecture_decision + git or PR observation

- [x] T030 [US5] Add TestCompileContextIngestedADROutranksGitAndPRObservation in internal/application/queries/compile_context_test.go
- [x] T031 [US5] Assert git/PR lore remains unverified repository_observation in existing ingest tests (no upgrade)

**Checkpoint**: Accepted ADRs look like architecture; git/PR capture does not

---

## Phase 8: Polish & Cross-Cutting Concerns

- [x] T032 [P] Update docs/api/rest.md with ADR ingest routes
- [x] T033 [P] Mark F032 DONE in docs/development/FEATURE_DEVELOPMENT.md; next F035 or F040; update Immediate recommended tasks
- [x] T034 [P] Update .cursor/rules/specify-rules.mdc, docs/development/contributing.md, README.md, docs/architecture/overview.md
- [x] T035 Run go test ./... and go vet ./...; fix failures
- [x] T036 Confirm F030 git and F031 PR contract tests still pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Story 1 (P1)**: Depends on Foundational
- **User Story 2 (P1)**: Depends on US1 handler skeleton
- **User Story 3 (P2)**: Depends on US1/US2 command
- **User Story 4 (P2)**: Depends on US1/US2 command
- **User Story 5 (P2)**: Characterization once trusted-source origin exists (can run after T006)
- **Polish**: After desired stories

### Parallel Opportunities

- T003–T005 tests in parallel
- T012 parser tests parallel with T013 once ports exist
- T021–T024 contract tests in parallel
- T032–T034 docs in parallel

### MVP

Phase 1 + 2 + User Story 1 (trusted-source capture of accepted ADRs + persist + outbox)

---

## Notes

- Do not implement F033, F035, F040
- Do not add an 11th MCP tool
- Do not loosen NewLoreEntry
- Do not overload git_ingest_shas or pr_ingest_prs
- Do not change compile ranking formulas
