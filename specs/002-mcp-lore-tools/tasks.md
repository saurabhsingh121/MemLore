# Tasks: MCP Lore Tools

**Input**: Design documents from `/specs/002-mcp-lore-tools/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Per the MemLore constitution, behavioral production work MUST use
TDD (RED → GREEN → REFACTOR). Explicit test tasks precede implementation in
each user-story phase.

**Organization**: Tasks are grouped by user story to enable independent
implementation and testing of each story. Application/domain/Postgres from
001 are reused; this feature is the MCP adapter + `memlore mcp` CLI.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **MemLore default**: `src/memlore/`, `tests/{unit,integration,contract,e2e}/`
- MCP tests live in `tests/unit/adapters/test_mcp_*.py` and
  `tests/contract/test_mcp_*.py` (not a `tests/**/mcp/` package — that name
  shadows the `mcp` SDK). Shared harness: `tests/support/mcp_client.py`.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add the official MCP SDK and adapter/test package layout

- [x] T001 Add runtime dependency `mcp>=2.1,<3` in `pyproject.toml` and refresh `uv.lock`
- [x] T002 [P] Create MCP adapter and test package stubs in `src/memlore/adapters/mcp/__init__.py`, `tests/unit/adapters/mcp/__init__.py`, `tests/contract/mcp/__init__.py`, and `tests/e2e/__init__.py`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: MCP server factory, error/payload mapping, in-memory contract harness shared by all tools

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] RED: Add failing unit tests for mapping `ValidationError` / `NotFoundError` to `ToolError` with `{code}: {message}` (`validation_error`, `not_found`; no internals) in `tests/unit/adapters/test_mcp_errors.py`
- [x] T004 GREEN: Implement error mapping in `src/memlore/adapters/mcp/errors.py`
- [x] T005 [P] RED: Add failing unit tests that lore/audit JSON fields match REST `LoreEntryResponse` / `AuditRecordResponse` in `tests/unit/adapters/test_mcp_payloads.py`
- [x] T006 GREEN: Implement payload mappers (reuse `src/memlore/adapters/rest/schemas.py` models) in `src/memlore/adapters/mcp/tools.py`
- [x] T007 RED: Add failing unit tests that `create_mcp_server(container)` returns an `MCPServer` named `memlore` with no Graphiti/Neo4j tools in `tests/unit/adapters/test_mcp_server.py`
- [x] T008 GREEN: Implement `create_mcp_server` plus stderr-only `configure_mcp_logging()` in `src/memlore/adapters/mcp/server.py`
- [x] T009 Add shared in-memory MCP `Client` fixture (`build_memory_container` + `create_mcp_server`) in `tests/support/mcp_client.py`

**Checkpoint**: Foundation ready — user story implementation can begin

**DONE WHEN**: `mcp` is importable; error/payload helpers exist; `create_mcp_server` builds a named server from `AppContainer`; contract tests can obtain an in-memory client; logs are designed for stderr; `ruff`/`mypy` clean for new modules.

---

## Phase 3: User Story 1 — Remember lore via MCP (Priority: P1) 🎯 MVP

**Goal**: Coding agents create human-authored scoped lore through `memlore.remember` with required `actor_id`

**Independent Test**: Call `memlore.remember` with statement, scope, actor, optional evidence; confirm id + provenance; invalid input stores nothing; duplicates get a new id

### Tests for User Story 1

> Write these tests FIRST; confirm RED before implementation

- [x] T010 [P] [US1] RED: Unit tests for remember wiring (required/non-empty `actor_id`, no env default, origin always `human_authored`, duplicates allowed, validation stores nothing) in `tests/unit/adapters/test_mcp_remember.py`
- [x] T011 [P] [US1] RED: Contract tests for `memlore.remember` per `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md` (success payload, missing/blank `actor_id` → `validation_error` + `is_error`, invalid statement/scope, duplicate → new id) in `tests/contract/test_mcp_remember.py`

### Implementation for User Story 1

- [x] T012 [US1] GREEN: Implement and register `memlore.remember` (call `CreateLoreHandler`; never accept origin from the client) in `src/memlore/adapters/mcp/tools.py` and `src/memlore/adapters/mcp/server.py`
- [x] T013 [US1] REFACTOR: Deduplicate argument parsing and add `log_operation` with `operation=mcp.remember` in `src/memlore/adapters/mcp/tools.py`; keep US1 tests green

**Checkpoint**: US1 independently demonstrable via unit + in-memory MCP contract tests

**DONE WHEN**: US1 unit/contract tests pass; remember persists via existing create service; origin is `human_authored`; missing/blank `actor_id` does not store a row.

---

## Phase 4: User Story 2 — Get and explain lore via MCP (Priority: P1)

**Goal**: Fetch a lore entry by id and return structured fields plus chronological audits (`memlore.get`, `memlore.explain`)

**Independent Test**: Seed an entry (MCP remember or `CreateLoreHandler`); get returns full fields; explain adds chronological `audits`; unknown id → `not_found` (not empty success)

### Tests for User Story 2

- [x] T014 [P] [US2] RED: Unit tests for get/explain composition (same UoW; unknown id; explain has no NL summary field) in `tests/unit/adapters/test_mcp_get_explain.py`
- [x] T015 [P] [US2] RED: Contract tests for `memlore.get` and `memlore.explain` per `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md` in `tests/contract/test_mcp_get_explain.py`

### Implementation for User Story 2

- [x] T016 [US2] GREEN: Implement and register `memlore.get` (`GetLoreHandler`) and `memlore.explain` (`GetLoreHandler` + `ListAuditsHandler` in one UoW) in `src/memlore/adapters/mcp/tools.py` and `src/memlore/adapters/mcp/server.py`
- [x] T017 [US2] REFACTOR: Share get-by-id error path and log `mcp.get` / `mcp.explain` in `src/memlore/adapters/mcp/tools.py`; keep US2 tests green

**Checkpoint**: US2 independently testable (seed via application handler if remember is unavailable)

**DONE WHEN**: US2 unit/contract tests pass; get returns REST-parity lore fields; explain adds chronological `audits`; unknown id → `not_found` without internals.

---

## Phase 5: User Story 3 — Verify lore via MCP (Priority: P1)

**Goal**: Verify via `memlore.verify` with required `actor_id`; self-verify allowed; idempotent re-verify; origin unchanged

**Independent Test**: Verify unverified entry; status/verifier/time set; second verify no extra `verify` audit; missing actor / unknown id fail clearly

### Tests for User Story 3

- [x] T018 [P] [US3] RED: Unit tests for verify wiring (blank `actor_id`, unknown id, idempotent re-verify with exactly one `verify` audit via `ListAuditsHandler`, origin preserved) in `tests/unit/adapters/test_mcp_verify.py`
- [x] T019 [P] [US3] RED: Contract tests for `memlore.verify` per `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md` (verified payload, origin unchanged, missing/blank `actor_id` → `validation_error`, unknown id → `not_found`; do **not** call `memlore.explain`—audit idempotence is T018) in `tests/contract/test_mcp_verify.py`

### Implementation for User Story 3

- [x] T020 [US3] GREEN: Implement and register `memlore.verify` (`VerifyLoreHandler`) in `src/memlore/adapters/mcp/tools.py` and `src/memlore/adapters/mcp/server.py`
- [x] T021 [US3] REFACTOR: Share mutating `actor_id` validation with remember and log `mcp.verify` in `src/memlore/adapters/mcp/tools.py`; keep US3 tests green

**Checkpoint**: US3 independently testable

**DONE WHEN**: US3 unit/contract tests pass; first verify matches REST semantics; re-verify is a no-op without a second verify audit; missing `actor_id` → `validation_error`; unknown id → `not_found`.

---

## Phase 6: User Story 4 — Search/list lore via MCP (Priority: P1)

**Goal**: List lore by exact scope through `memlore.search` (empty list when none; validation when scope incomplete)

**Independent Test**: Entries in scopes A and B; search A returns only A; empty scope → `items: []`; missing kind/key → `validation_error`

### Tests for User Story 4

- [x] T022 [P] [US4] RED: Unit tests for search wiring (exact kind+key, empty list, incomplete scope) in `tests/unit/adapters/test_mcp_search.py`
- [x] T023 [P] [US4] RED: Contract tests for `memlore.search` per `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md` in `tests/contract/test_mcp_search.py`

### Implementation for User Story 4

- [x] T024 [US4] GREEN: Implement and register `memlore.search` (`ListLoreByScopeHandler`; `{items: [...]}`) in `src/memlore/adapters/mcp/tools.py` and `src/memlore/adapters/mcp/server.py`
- [x] T025 [US4] REFACTOR: Align search `scope` args with remember and log `mcp.search` in `src/memlore/adapters/mcp/tools.py`; keep US4 tests green

**Checkpoint**: US4 independently testable

**DONE WHEN**: US4 unit/contract tests pass; search is exact scope list only (not semantic); empty match is `items: []`; incomplete scope → `validation_error`.

---

## Phase 7: User Story 5 — Run local MCP server (Priority: P1)

**Goal**: Developers start a stdio MCP server with `memlore mcp` so agents can attach and complete the five-tool path

**Independent Test**: `uv run memlore mcp` advertises the five lore tools and completes remember → get → verify → explain → search

### Tests for User Story 5

- [x] T026 [P] [US5] RED: Contract test that `tools/list` is exactly `memlore.remember`, `memlore.get`, `memlore.verify`, `memlore.explain`, `memlore.search` (no Graphiti/Neo4j/`get_for_task`/`supersede`/`invalidate`) in `tests/contract/test_mcp_list_tools.py`
- [x] T027 [P] [US5] RED: e2e stdio test spawning `memlore mcp`, listing tools, then remember → get → verify → explain → search (skip if Postgres unavailable; no CI wall-clock assert for SC-001) in `tests/e2e/test_mcp_stdio.py`

### Implementation for User Story 5

- [x] T028 [US5] GREEN: Add `mcp` subcommand that builds `AppContainer`, configures stderr logging, and runs stdio in `src/memlore/adapters/cli/main.py`
- [x] T029 [US5] REFACTOR: Keep stdout protocol-pure and document `memlore mcp` in argparse help in `src/memlore/adapters/cli/main.py`; make T026–T027 green

**Checkpoint**: US5 independently demonstrable with CLI + e2e (requires US1–US4 tools registered)

**DONE WHEN**: `uv run memlore mcp` speaks MCP on stdio; tool list matches the contract; remember → get → verify → explain → search works against the governance DB when Postgres is up.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Product docs, quickstart, quality gates for the MCP surface

- [x] T030 [P] Update `docs/api/mcp.md` from planned to implemented (five tools, actor_id, stdio CLI, out of scope)
- [x] T031 [P] Update agent-connect and CLI sections in `README.md` and `docs/development/setup.md`
- [x] T032 [P] Update first-slice/MCP status in `docs/architecture/overview.md`
- [x] T033 Verify `specs/002-mcp-lore-tools/quickstart.md` five-tool path (remember, get, verify, explain, search) against `uv run memlore mcp` and the in-memory contract suite; confirm it completes in local development in under 2 minutes (SC-001, manual)
- [x] T034 Run `uv run ruff check src tests`, `uv run ruff format --check src tests`, `uv run mypy`, and `uv run pytest tests/unit tests/contract tests/integration tests/e2e`; fix failures under `src/memlore/` and `tests/`

**DONE WHEN**: Docs match implemented MCP behavior; REST still green; `ruff` / `mypy` / pytest (unit+contract+integration+e2e) pass on the feature branch (integration/e2e skip if Postgres is down).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)** → **Phase 2 (Foundational)** → User stories
- **US1 (Remember)** first MVP increment (write path)
- **US2 (Get/Explain)** can seed via `CreateLoreHandler` (independent of MCP remember) but typically follows US1
- **US3 (Verify)** needs an existing entry (seed via create handler or US1)
- **US4 (Search)** needs entries in two scopes (seed via create handler or US1)
- **US5 (stdio CLI)** depends on US1–US4 so the advertised tool list and five-tool e2e path are complete
- **Polish** after desired stories complete

### User Story Dependencies

```text
Foundational
    ├──► US1 Remember ──┐
    ├──► US2 Get/Explain ┼──► US5 stdio CLI (full tool list + five-tool e2e)
    ├──► US3 Verify     ┤
    └──► US4 Search     ┘
```

US2–US4 can proceed in parallel after Foundational if each story seeds data through existing application services (not through MCP remember). US3 contract tests (T019) MUST NOT call `memlore.explain`. They all edit `src/memlore/adapters/mcp/tools.py` and `src/memlore/adapters/mcp/server.py`, so a **single implementer should run them sequentially** to avoid file conflicts.

### Within Each User Story

1. RED tests (unit + contract as listed; e2e for US5)
2. Confirm failures
3. GREEN implementation (adapter tools → server registration → CLI for US5)
4. REFACTOR with suite green

### Parallel Opportunities

- Setup T002 in parallel with T001 after T001’s lockfile refresh, or immediately after T001
- Foundational: T003 + T005 RED in parallel (different files)
- Within a story: unit + contract RED tests in parallel before GREEN
- Polish T030–T032 in parallel
- After Foundational, US2–US4 *tests* can be written in parallel; implementations should be sequential on shared adapter files

---

## Parallel Example: User Story 1

```bash
# After Foundational checkpoint, in parallel:
# - tests/unit/adapters/mcp/test_remember.py
# - tests/contract/mcp/test_remember.py
# Then sequential GREEN: tools.py remember + server.py register → refactor/logging
```

## Parallel Example: After US1 MVP

```bash
# Parallel RED (different test files):
# - US2 tests/unit/adapters/mcp/test_get_explain.py + tests/contract/mcp/test_get_explain.py
# - US3 tests/unit/adapters/mcp/test_verify.py + tests/contract/mcp/test_verify.py
# - US4 tests/unit/adapters/mcp/test_search.py + tests/contract/mcp/test_search.py
# Then sequential GREEN on tools.py / server.py: US2 → US3 → US4 → US5 CLI
```

---

## Implementation Strategy

### MVP First

1. Complete Phase 1–2
2. Complete **US1 only** (`memlore.remember`) → agents can write lore
3. Add US2 get/explain → readable + inspectable provenance
4. Add US3 verify → authority distinction over MCP
5. Add US4 search → scoped discovery
6. Add US5 `memlore mcp` stdio → agents can attach
7. Polish/docs/CI green

### Incremental Delivery

Each story checkpoint should leave `pytest` green for completed phases. Do not require unfinished tools to pass earlier stories (except US5, which asserts the full five-tool list).

---

## Task Summary

| Phase | Story | Task IDs | Count |
|-------|-------|----------|-------|
| Setup | — | T001–T002 | 2 |
| Foundational | — | T003–T009 | 7 |
| US1 Remember | US1 | T010–T013 | 4 |
| US2 Get/Explain | US2 | T014–T017 | 4 |
| US3 Verify | US3 | T018–T021 | 4 |
| US4 Search | US4 | T022–T025 | 4 |
| US5 stdio CLI | US5 | T026–T029 | 4 |
| Polish | — | T030–T034 | 5 |
| **Total** | | T001–T034 | **34** |

**Suggested first demo increment**: Phases 1–3 (US1 remember only). **Feature-complete for 002 is US1–US5** (`search` and `memlore mcp` are in-spec; do not stop after US1).

**Format validation**: All tasks use `- [ ]`, sequential `T###`, optional `[P]`, story labels on US phases only, and explicit file paths.
