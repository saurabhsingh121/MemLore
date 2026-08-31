# Tasks: Membership-scoped authorization (F010 remainder)

**Input**: Design documents from `/specs/018-membership-authz/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/  
**Tests**: TDD RED → GREEN → REFACTOR (constitution). Every behavioral task
includes an explicit test-first step.

**Organization**: Phases follow user stories US1–US6 from spec.md.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Parallelizable (different files, no incomplete deps)
- **[Story]**: US1…US6
- Paths are repo-relative under MemLore Go core

## Phase 1: Setup

**Purpose**: Confirm feature branch and design artifacts

- [x] T001 Confirm branch `018-membership-authz` and artifacts under `specs/018-membership-authz/` (spec, plan, research, data-model, contracts, quickstart)

---

## Phase 2: Foundational (schema + ports + domain policy)

**Purpose**: Persistence and pure AuthorizeScope — blocks all stories  
**⚠️** No story work until this phase is green

### Schema + sqlc

- [x] T002 RED: migration smoke — add failing expectation in `migrations/migration_integration_helpers_test.go` or new membership migration test that `users`/`teams` tables exist after up
- [x] T003 GREEN: write `migrations/00004_membership.sql` (users, teams, projects, memberships, scope_bindings) with goose Up/Down; RESTRICT FKs
- [x] T004 [P] Add `db/queries/membership.sql` (ensure user, CRUD team/project, memberships, bindings, lookups by subject+scope)
- [x] T005 GREEN: run `sqlc generate`; commit generated code under `internal/infrastructure/postgres/sqlc/`

### Domain policy (pure)

- [x] T006 [P] RED: `internal/domain/membership_test.go` — team allow/deny; org≡team; project direct + parent-team inherit; unbound child deny; admin bypass; CanAccess helpers
- [x] T007 GREEN: `internal/domain/membership.go` — `AuthorizeScope` / access resolution with injectable membership view (no pgx)
- [x] T008 [P] DONE WHEN: `go test ./internal/domain/ -count=1` covers SC ACL rules from clarifications Q1–Q3

### Ports + Postgres directory

- [x] T009 [P] Add `internal/application/ports/membership.go` — `MembershipDirectory` (EnsureUser, CreateTeam/Project, Add/RemoveMember, Bind/Unbind, ResolveAccess inputs)
- [x] T010 GREEN: Postgres implementation wrapping sqlc in `internal/infrastructure/postgres/` (membership directory)
- [x] T011 RED/GREEN: integration test (tag `integration` if repo pattern) seeding membership and resolving access via directory
- [x] T012 Wire directory into bootstrap / UoW or standalone pool used by authz helper — `internal/bootstrap/` or existing wire in `cmd/memlore`

**Checkpoint**: Domain ACL + DB round-trip work; no HTTP lore gating yet

---

## Phase 3: User Story 1 — Tenant isolation on lore access (P1) 🎯 MVP

**Goal**: OIDC writer member of A can mutate/read A; cannot access B; get-by-id no leak  
**Independent Test**: HMAC OIDC fixtures + seeded memberships; create/get/list allow/deny

### Tests first

- [x] T013 [P] [US1] RED: HTTP contract tests in `internal/adapters/http/membership_contract_test.go` — writer in team alpha: create/get/list alpha OK; create/list beta → forbidden; get beta id → not_found; reader in alpha create → forbidden (verb)
- [x] T014 [P] [US1] RED: shared authz helper unit tests if extracted (`internal/application/auth` or `internal/application/authz`)

### Implementation

- [x] T015 [US1] GREEN: application authz helper — `MembershipEnforced()` from OIDC Enabled; load directory; call domain after role check
- [x] T016 [US1] GREEN: HTTP create + list-by-scope gate on request scope (`internal/adapters/http/handlers.go` / `auth.go`)
- [x] T017 [US1] GREEN: HTTP get/explain/audits/verify/invalidate/supersede — post-load AuthorizeScope; map no-access → not_found
- [x] T018 [US1] REFACTOR: single helper used by all lore routes; keep F111 error mapping
- [x] T019 [US1] DONE WHEN: T013 suite green; existing local-mode lore contracts still pass without membership seed

**Checkpoint**: US1 MVP — tenant isolation on primary lore paths

---

## Phase 4: User Story 2 — List/search/compile never leak (P1)

**Goal**: search / compile / knowledge_search filter to allowed scopes  
**Independent Test**: lore in alpha+beta; member of alpha only sees alpha

### Tests first

- [x] T020 [P] [US2] RED: extend `membership_contract_test.go` (or dedicated) — search/compile/knowledge_search omit beta; list beta → forbidden; no memberships → empty/forbidden per contract

### Implementation

- [x] T021 [US2] GREEN: filter search/knowledge_search results by CanAccessScope
- [x] T022 [US2] GREEN: compile / get_for_task — forbid inaccessible explicit scope; filter retrieved items
- [x] T023 [US2] DONE WHEN: T020 green; no foreign tenant lore in response bodies

**Checkpoint**: Read-path leak coverage complete

---

## Phase 5: User Story 3 — Admin membership bypass (P1)

**Goal**: JWT admin with zero memberships can access any lore scope  
**Independent Test**: admin token, no membership rows, get/list team B OK

### Tests first

- [x] T024 [P] [US3] RED: membership contract — admin no memberships can get/list beta; writer no memberships cannot

### Implementation

- [x] T025 [US3] GREEN: confirm domain admin short-circuit; fix if adapters skip incorrectly
- [x] T026 [US3] DONE WHEN: T024 green; writer deny still holds

**Checkpoint**: Platform operator path proven

---

## Phase 6: User Story 4 — Local mode unchanged (P1)

**Goal**: OIDC unset → membership off; F111 contracts green without seed  
**Independent Test**: existing `lore_contract_test.go` + `mcp_contract_test.go` + `auth_contract_test.go`

### Tests first

- [x] T027 [P] [US4] RED if needed: assert MembershipEnforced false when OIDC unset in `internal/application/auth` test
- [x] T028 [US4] Run existing HTTP/MCP lore + auth contract suites (no membership seed) — must stay green

### Implementation

- [x] T029 [US4] GREEN: ensure adapters skip membership when `!Config.Enabled()`
- [x] T030 [US4] DONE WHEN: `go test ./internal/adapters/http/ ./internal/adapters/mcp/ -count=1` green without membership fixtures for legacy suites

**Checkpoint**: Dogfood/CI local mode preserved

---

## Phase 7: User Story 5 — Admin REST tenancy APIs (P2)

**Goal**: Admin CRUD teams/projects/members/bindings; non-admin forbidden  
**Independent Test**: admin creates team+project+binding+member; writer cannot call admin APIs

### Tests first

- [x] T031 [P] [US5] RED: `internal/adapters/http/admin_membership_contract_test.go` — admin create team/project/member/binding; writer → forbidden; member gains child-scope access after bind

### Implementation

- [x] T032 [US5] GREEN: application commands in `internal/application/membership/` (create team/project, members, bindings, ensure user)
- [x] T033 [US5] GREEN: HTTP routes `/v1/admin/...` in `internal/adapters/http/` wired in router/`cmd/memlore`
- [x] T034 [US5] DONE WHEN: T031 green; MCP tool count still 9

**Checkpoint**: Control plane usable for seeding enforcement tests (may backfill helpers used by US1 fixtures)

---

## Phase 8: User Story 6 — REST + MCP shared policy (P2)

**Goal**: MCP remember/get/list/search match HTTP allow/deny  
**Independent Test**: same subject/role/membership fixtures on MCP tools

### Tests first

- [x] T035 [P] [US6] RED: `internal/adapters/mcp/membership_contract_test.go` — non-member remember/get denied; member remember OK; tool count == 9

### Implementation

- [x] T036 [US6] GREEN: MCP adapter calls same authz helper after role check
- [x] T037 [US6] DONE WHEN: T035 green; parity with HTTP for create/get scoped ops

**Checkpoint**: Adapter parity complete

---

## Phase 9: Polish & docs

- [x] T038 [P] Update `docs/api/rest.md` with admin routes + membership semantics
- [x] T039 [P] Update `docs/architecture/security.md` (scope-aware authz now implemented)
- [x] T040 [P] Update `docs/development/FEATURE_DEVELOPMENT.md` — F010 closer to DONE; add F114 row if matching tracker style
- [x] T041 [P] Update `.env.example` if any new vars (none expected if gate=OIDC only); note in F111 spec that FR-010 completed by 018
- [x] T042 Full `go test ./...` green; confirm `graph-service/` untouched (`git diff --stat` excludes it)
- [x] T043 REFACTOR pass: names, duplication, delete dead code; keep tests green

---

## Dependencies

```text
Phase 1 → Phase 2 (schema/domain/ports)
       → Phase 3 US1 (MVP lore isolation) ──┬→ Phase 4 US2 (filter reads)
                                           ├→ Phase 5 US3 (admin bypass)
                                           └→ Phase 6 US4 (local mode verify)
Phase 2 → Phase 7 US5 (admin REST) — ideally before/alongside US1 fixtures
Phase 3+ → Phase 8 US6 (MCP parity)
All stories → Phase 9 polish
```

Suggested order: T001–T012 → T031–T034 (admin APIs for seeding) → T013–T019 → T020–T026 → T027–T030 → T035–T037 → T038–T043.

## Parallel opportunities

- T004 ∥ T006 (queries vs domain tests)
- T009 ∥ T006
- T013 ∥ T014
- T020 ∥ T024 ∥ T027
- T038 ∥ T039 ∥ T040 ∥ T041

## MVP

Phases 1–3 (US1) + enough of US5 to seed memberships = shippable tenant isolation.

## Implementation strategy

1. RED domain AuthorizeScope → GREEN  
2. Migration + sqlc + directory  
3. Admin REST (seed path)  
4. HTTP lore gates (US1) then filters (US2) + admin bypass tests (US3)  
5. Confirm local mode (US4)  
6. MCP parity (US6)  
7. Docs + full suite
