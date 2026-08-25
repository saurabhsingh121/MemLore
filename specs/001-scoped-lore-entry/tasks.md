# Tasks: Scoped Human-Authored Lore Entry

**Input**: Design documents from `/specs/001-scoped-lore-entry/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per the MemLore constitution, behavioral production work MUST use
TDD (RED → GREEN → REFACTOR). Explicit test tasks precede implementation in
each user-story phase.

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **MemLore default**: `src/memlore/`, `tests/{unit,integration,contract,e2e}/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Dependencies and config needed for governance-plane persistence

- [ ] T001 Add SQLAlchemy, Alembic, and psycopg dependencies in `pyproject.toml` and refresh `uv.lock`
- [ ] T002 [P] Add application settings for Postgres DSN in `src/memlore/bootstrap/settings.py`
- [ ] T003 [P] Add pytest integration markers, Postgres skip/bring-up guidance, and shared fixtures skeleton in `tests/conftest.py` (require `docker compose up -d postgres` or `MEMLORE_TEST_DATABASE_URL`)
- [ ] T004 Document local Postgres + migrate workflow updates in `docs/development/setup.md` and `docs/operations/migrations.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Domain types, ports, Postgres/Alembic baseline, REST wiring shared by all stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 [P] RED: Add failing unit tests for scope/evidence/origin/verification enums in `tests/unit/domain/test_lore_enums.py`
- [ ] T006 [P] GREEN: Implement domain enums in `src/memlore/domain/models/enums.py`
- [ ] T007 [P] RED: Add failing unit tests for `Scope` and `EvidenceReference` value objects in `tests/unit/domain/test_scope_evidence.py`
- [ ] T008 GREEN: Implement `Scope` and `EvidenceReference` in `src/memlore/domain/models/scope.py` and `src/memlore/domain/models/evidence.py`
- [ ] T009 [P] RED: Add failing unit tests for `LoreEntry` create defaults in `tests/unit/domain/test_lore_entry.py`
- [ ] T010 GREEN: Implement `LoreEntry` and `AuditRecord` domain models in `src/memlore/domain/models/lore_entry.py` and `src/memlore/domain/models/audit_record.py`
- [ ] T011 Define repository/UoW/clock ports in `src/memlore/application/ports/lore.py`, `src/memlore/application/ports/audit.py`, `src/memlore/application/ports/unit_of_work.py`, and `src/memlore/application/ports/clock.py`
- [ ] T012 Add domain exceptions in `src/memlore/domain/exceptions.py`
- [ ] T013 Configure Alembic env and initial empty migration scaffolding in `alembic.ini` and `alembic/env.py`
- [ ] T014 Implement SQLAlchemy engine/session factory in `src/memlore/infrastructure/postgres/session.py`
- [ ] T015 Implement SQLAlchemy table mappings for lore entries and audits in `src/memlore/infrastructure/postgres/models.py`
- [ ] T016 Generate and verify Alembic migration for lore/audit tables in `alembic/versions/`
- [ ] T017 [P] Implement in-memory fakes for lore/audit/clock ports in `tests/support/fakes.py` for application unit tests
- [ ] T018 Add REST error envelope helpers and exception handlers in `src/memlore/adapters/rest/errors.py`
- [ ] T019 Add `X-Memlore-Actor` dependency helper in `src/memlore/adapters/rest/deps.py`
- [ ] T020 Wire create_app composition root (DB session, routers, `SystemClock` in `src/memlore/infrastructure/clock.py`) in `src/memlore/bootstrap/container.py` and `src/memlore/adapters/rest/app.py`
- [ ] T021 [P] Add structured logging helper with operation/actor/lore_entry_id fields in `src/memlore/infrastructure/telemetry/logging.py`

**Checkpoint**: Foundation ready — user story implementation can begin

**DONE WHEN**: Domain enums/models + ports exist; Alembic can create lore/audit schema; app composition root wires session/clock/routers; unit fakes available; `ruff`/`mypy` clean for new modules.

---

## Phase 3: User Story 1 — Remember scoped lore (Priority: P1) 🎯 MVP

**Goal**: Create human-authored lore with scope, evidence, provenance, and create audit

**Independent Test**: POST create returns 201 with id/provenance; create audit exists; validation failures store nothing; duplicates allowed

### Tests for User Story 1

> Write these tests FIRST; confirm RED before implementation

- [ ] T022 [P] [US1] RED: Unit tests for create-lore application service (validation incl. statement >8000 reject, duplicates allowed, create audit) in `tests/unit/application/test_create_lore.py`
- [ ] T023 [P] [US1] RED: Contract tests for `POST /v1/lore-entries` per `contracts/rest-lore-entries.md` (incl. missing fields, statement length 8001 → 400, duplicate statement/scope → 201 new id) in `tests/contract/test_create_lore_entry.py`
- [ ] T024 [US1] RED: Integration test for Postgres persist of lore + create audit atomically in `tests/integration/test_lore_repository_create.py`

### Implementation for User Story 1

- [ ] T025 [US1] GREEN: Implement `CreateLore` command/service in `src/memlore/application/commands/create_lore.py` and `src/memlore/application/services/lore_service.py`
- [ ] T026 [US1] GREEN: Implement Postgres `LoreRepository` + `AuditRepository` create paths in `src/memlore/infrastructure/postgres/lore_repository.py` and `src/memlore/infrastructure/postgres/audit_repository.py`
- [ ] T027 [US1] GREEN: Implement UnitOfWork transaction boundary in `src/memlore/infrastructure/postgres/unit_of_work.py`
- [ ] T028 [US1] GREEN: Add Pydantic request/response schemas in `src/memlore/adapters/rest/schemas.py`
- [ ] T029 [US1] GREEN: Implement `POST /v1/lore-entries` route in `src/memlore/adapters/rest/routes_lore.py` and register router in `src/memlore/adapters/rest/app.py`
- [ ] T030 [US1] REFACTOR: Clean naming/duplication while keeping US1 tests green; add create operation logging

**Checkpoint**: US1 independently demonstrable via contract + integration tests

**DONE WHEN**: Focused US1 unit/contract/integration tests pass; create persists lore+create audit atomically; invalid creates leave no rows; docs for this behavior not yet required until polish.

---

## Phase 4: User Story 2 — Retrieve lore with provenance (Priority: P1)

**Goal**: Get lore by id with full provenance fields; unknown id → 404

**Independent Test**: Create then GET by id returns complete fields; unknown id returns not_found envelope

### Tests for User Story 2

- [ ] T031 [P] [US2] RED: Unit tests for get-lore query service in `tests/unit/application/test_get_lore.py`
- [ ] T032 [P] [US2] RED: Contract tests for `GET /v1/lore-entries/{id}` in `tests/contract/test_get_lore_entry.py`
- [ ] T033 [US2] RED: Integration test for repository get/not-found in `tests/integration/test_lore_repository_get.py`

### Implementation for User Story 2

- [ ] T034 [US2] GREEN: Implement get-lore query in `src/memlore/application/queries/get_lore.py` (wire through lore service)
- [ ] T035 [US2] GREEN: Implement repository get-by-id mapping in `src/memlore/infrastructure/postgres/lore_repository.py`
- [ ] T036 [US2] GREEN: Implement `GET /v1/lore-entries/{id}` in `src/memlore/adapters/rest/routes_lore.py`
- [ ] T037 [US2] REFACTOR: Shared response mapping helpers if needed in `src/memlore/adapters/rest/schemas.py`; keep tests green

**Checkpoint**: US2 independently testable

**DONE WHEN**: US2 unit/contract/integration tests pass; get returns full provenance; unknown id → not_found envelope without internal details.

---

## Phase 5: User Story 3 — Verify human-authored lore (Priority: P1)

**Goal**: Verify unverified lore (self-verify allowed); idempotent re-verify; preserve statement/origin

**Independent Test**: Verify sets status/verifier/time; second verify no-op without second verify audit; missing actor rejected

### Tests for User Story 3

- [ ] T038 [P] [US3] RED: Unit tests for verify rules (self-verify, idempotent, preserves origin) in `tests/unit/application/test_verify_lore.py`
- [ ] T039 [P] [US3] RED: Contract tests for `POST /v1/lore-entries/{id}/verify` in `tests/contract/test_verify_lore_entry.py`
- [ ] T040 [US3] RED: Integration test for verify + single verify audit in `tests/integration/test_lore_verify.py`

### Implementation for User Story 3

- [ ] T041 [US3] GREEN: Implement verify domain/application behavior in `src/memlore/domain/services/verification.py` and `src/memlore/application/commands/verify_lore.py`
- [ ] T042 [US3] GREEN: Persist verification fields and conditional verify audit in `src/memlore/infrastructure/postgres/lore_repository.py`, `src/memlore/infrastructure/postgres/audit_repository.py`, and `src/memlore/infrastructure/postgres/unit_of_work.py`
- [ ] T043 [US3] GREEN: Implement verify route in `src/memlore/adapters/rest/routes_lore.py`
- [ ] T044 [US3] REFACTOR: Extract shared actor validation; keep US3 tests green; log verify operations

**Checkpoint**: US3 independently testable

**DONE WHEN**: US3 unit/contract/integration tests pass; first verify writes one verify audit; re-verify is no-op without a second verify audit; statement/origin unchanged.

---

## Phase 6: User Story 5 — Inspect audit trail (Priority: P1)

**Goal**: List audits for a lore entry id chronologically; unknown entry → 404

**Independent Test**: After create+verify, audits list shows create then verify with actors/timestamps

### Tests for User Story 5

- [ ] T045 [P] [US5] RED: Unit tests for list-audits query in `tests/unit/application/test_list_audits.py`
- [ ] T046 [P] [US5] RED: Contract tests for `GET /v1/lore-entries/{id}/audits` in `tests/contract/test_list_audits.py`
- [ ] T047 [US5] RED: Integration test for audit ordering and not-found in `tests/integration/test_audit_repository_list.py`

### Implementation for User Story 5

- [ ] T048 [US5] GREEN: Implement list-audits query in `src/memlore/application/queries/list_audits.py`
- [ ] T049 [US5] GREEN: Implement audit list-by-target in `src/memlore/infrastructure/postgres/audit_repository.py`
- [ ] T050 [US5] GREEN: Implement audits route in `src/memlore/adapters/rest/routes_lore.py`
- [ ] T051 [US5] REFACTOR: Align audit response schema; keep tests green

**Checkpoint**: US5 independently testable (assumes US1 create audit writer already present)

**DONE WHEN**: US5 unit/contract/integration tests pass; create+verify yields chronological create then verify audits; unknown lore id → 404 (not empty list).

---

## Phase 7: User Story 4 — List lore within a scope (Priority: P2)

**Goal**: List entries by exact `scope_kind` + `scope_key`; empty list when none

**Independent Test**: Entries in A and B; list A returns only A; kind/key mismatch excluded

### Tests for User Story 4

- [ ] T052 [P] [US4] RED: Unit tests for list-by-scope query in `tests/unit/application/test_list_lore_by_scope.py`
- [ ] T053 [P] [US4] RED: Contract tests for `GET /v1/lore-entries?scope_kind=&scope_key=` in `tests/contract/test_list_lore_entries.py`
- [ ] T054 [US4] RED: Integration test for scope filter precision in `tests/integration/test_lore_repository_list_by_scope.py`

### Implementation for User Story 4

- [ ] T055 [US4] GREEN: Implement list-by-scope query in `src/memlore/application/queries/list_lore_by_scope.py`
- [ ] T056 [US4] GREEN: Implement repository list filter + index usage path in `src/memlore/infrastructure/postgres/lore_repository.py`
- [ ] T057 [US4] GREEN: Implement list route/query params in `src/memlore/adapters/rest/routes_lore.py`
- [ ] T058 [US4] REFACTOR: Keep list/get response mapping consistent; tests green

**Checkpoint**: US4 independently testable

**DONE WHEN**: US4 unit/contract/integration tests pass; list filter matches exact kind+key only; empty scope returns empty items array.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Docs, quickstart verification, full suite quality gates

- [ ] T059 [P] Update `docs/api/rest.md` to mark lore endpoints as implemented with examples
- [ ] T060 [P] Update `docs/concepts/lore.md` and `docs/concepts/provenance.md` for create/verify/audit behaviors
- [ ] T061 [P] Verify and adjust `specs/001-scoped-lore-entry/quickstart.md` against running API
- [ ] T062 Run full relevant suite (`uv run ruff check`, `uv run ruff format --check`, `uv run mypy`, `uv run pytest`) and fix failures
- [ ] T063 Add brief ADR or architecture note only if implementation drifted from ADR 0001/0002; otherwise update `docs/architecture/overview.md` status of first slice

**DONE WHEN**: Docs match implemented REST behavior; `ruff` / `mypy` / `pytest` (unit+contract+integration) pass on the feature branch.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)** → **Phase 2 (Foundational)** → User stories
- **US1 (Create)** first MVP; enables data for later stories
- **US2 (Get)** can follow US1 (needs persisted entries)
- **US3 (Verify)** depends on US1 (and typically US2 for assertion convenience)
- **US5 (Audits)** depends on US1 create-audit writer; verify-audit needs US3
- **US4 (List)** depends on US1; independent of verify/audits
- **Polish** after desired stories complete

### User Story Dependencies

```text
US1 (Create) ──┬──► US2 (Get)
               ├──► US3 (Verify) ──► US5 (Audits read)
               └──► US4 (List)
```

### Within Each User Story

1. RED tests (unit/contract/integration as listed)
2. Confirm failures
3. GREEN implementation (domain/app → infra → REST)
4. REFACTOR with suite green

### Parallel Opportunities

- [P] tasks in Setup/Foundational with different files
- Within a story: unit + contract RED tests in parallel before GREEN
- After US1: US2 and US4 can proceed in parallel; US3 then US5 sequentially preferred

---

## Parallel Example: User Story 1

```bash
# After Foundational checkpoint, in parallel:
# - tests/unit/application/test_create_lore.py
# - tests/contract/test_create_lore_entry.py
# Then sequential GREEN: service → repos/UoW → schemas → route
```

## Parallel Example: After US1 MVP

```bash
# Parallel tracks:
# Track A: US2 get tests + implementation
# Track B: US4 list tests + implementation
# Then US3 verify → US5 audits
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1–2
2. Complete **US1 only** (create + create audit) → demo/store value
3. Add US2 get → readable memory
4. Add US3 verify → authority distinction
5. Add US5 audit read → inspectable provenance
6. Add US4 list → discovery
7. Polish/docs/CI green

### Incremental Delivery

Each story checkpoint should leave `pytest` green for completed phases without requiring unfinished stories.

---

## Task Summary

| Phase | Story | Task IDs | Count |
|-------|-------|----------|-------|
| Setup | — | T001–T004 | 4 |
| Foundational | — | T005–T021 | 17 |
| US1 Remember | US1 | T022–T030 | 9 |
| US2 Retrieve | US2 | T031–T037 | 7 |
| US3 Verify | US3 | T038–T044 | 7 |
| US5 Audits | US5 | T045–T051 | 7 |
| US4 List | US4 | T052–T058 | 7 |
| Polish | — | T059–T063 | 5 |
| **Total** | | T001–T063 | **63** |

**Suggested MVP scope**: Phases 1–3 (through US1).

**Format validation**: All tasks use `- [ ]`, sequential `T###`, optional `[P]`, story labels on US phases only, and explicit file paths.
