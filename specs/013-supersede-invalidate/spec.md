# Feature Specification: Governance lifecycle — invalidate + supersede (F110)

**Feature Branch**: `013-supersede-invalidate`  
**Created**: 2026-08-28  
**Status**: Ready  
**Depends on**: F106a (MCP lore tools), F104 (verify), F103 (Postgres persistence)  
**Implements**: Product F008 (partial — invalidate + supersede; conflict detection remains F112)

## Goal

Let humans and agents **invalidate** lore without erasing evidence or audit
history, and **supersede** lore with a successor entry that preserves the
predecessor as queryable history. Both operations are available on REST and
MCP with identical entry shape and error codes.

## Clarifications

### Session 2026-08-28

Decisions encoded from the F110 implementation prompt (schema and
idempotency). No remaining product questions block planning.

- Q: Invalidating an already-superseded entry? → A: `validation_error` (reject; do not silently no-op).
- Q: Re-invalidating an already-invalidated entry? → A: Idempotent no-op; no duplicate audit (same pattern as re-verify).
- Q: Superseding an already-superseded entry? → A: `validation_error` (not silent).
- Q: Superseding an invalidated entry? → A: `validation_error` (predecessor must not be invalidated).
- Q: How is supersession stored? → A: Predecessor holds `superseded_by_id` pointing at the successor. Current vs superseded is derived from whether that pointer is set. No separate `supersession_status` column in v1.
- Q: Where is invalidate actor/time stored? → A: `invalidated_by` / `invalidated_at` on the entry (parity with `verified_by` / `verified_at`) plus an `invalidate` audit.
- Q: Audit shape for supersede? → A: Dual audits in one transaction: `supersede` on the predecessor, `create` on the successor.
- Q: Successor evidence when omitted? → A: Empty list; predecessor evidence is not copied.
- Q: Outbox? → A: Successor emit uses the existing create/episode-ingest outbox event. No new outbox event types for invalidate or predecessor supersede.
- Q: Verify after invalidate or supersede? → A: `validation_error` (cannot revive). Invalidating a verified entry is allowed.
- Q: Filtering superseded/invalidated from search / get_for_task? → A: Deferred to F112.

## User Scenarios & Testing

### User Story 1 — Invalidate lore without deleting history (Priority: P1)

A maintainer discovers a stored rule is wrong. They mark it invalid. The
statement, scope, origin, evidence, and prior audits remain. Later readers
can still fetch and explain the entry; they see it is invalidated and who
did it.

**Why this priority**: Invalidation is the minimum temporal-correctness
operation: knowledge can become untrusted without being erased.

**Independent Test**: Create an entry, invalidate it, get/explain it, and
confirm status is invalidated, evidence is intact, and a single invalidate
audit exists. Re-invalidate and confirm no second audit.

**Acceptance Scenarios**:

1. **Given** an unverified or verified lore entry that is not superseded, **When** an actor invalidates it, **Then** verification status becomes `invalidated`, statement/scope/origin/evidence/prior audits are unchanged, and one `invalidate` audit is appended with that actor and timestamp.
2. **Given** an already-invalidated entry, **When** any actor invalidates it again, **Then** the entry is unchanged and no additional audit is written.
3. **Given** a missing actor, **When** invalidate is requested, **Then** the system returns a validation error and does not change the entry.
4. **Given** an unknown id, **When** invalidate is requested, **Then** the system returns not found.
5. **Given** a superseded entry, **When** invalidate is requested, **Then** the system returns a validation error and does not change the entry.

---

### User Story 2 — Supersede lore with a successor (Priority: P1)

A maintainer replaces a rule with a better statement. The old entry remains
fetchable and explainable as history; a new current entry exists in the same
scope. The old entry points at the successor.

**Why this priority**: Supersession is how MemLore changes truth without
overwriting history (constitution: temporal correctness).

**Independent Test**: Create an entry, supersede it with a new statement,
get both ids, and confirm predecessor is linked to successor, successor is
unverified human-authored lore in the same scope, and audits exist on both.

**Acceptance Scenarios**:

1. **Given** a current (not superseded, not invalidated) lore entry, **When** an actor supersedes it with a non-empty replacement statement, **Then** a new successor entry is created in the same scope with origin `human_authored`, the predecessor stores `superseded_by_id` of the successor, a `supersede` audit is appended on the predecessor, and a `create` audit is appended on the successor.
2. **Given** optional evidence on the supersede request, **When** the successor is created, **Then** that evidence is stored on the successor; if omitted, successor evidence is empty (predecessor evidence is not copied).
3. **Given** an already-superseded entry, **When** supersede is requested, **Then** the system returns a validation error and creates no successor.
4. **Given** an invalidated entry, **When** supersede is requested, **Then** the system returns a validation error and creates no successor.
5. **Given** a missing actor or empty statement, **When** supersede is requested, **Then** the system returns a validation error.
6. **Given** an unknown id, **When** supersede is requested, **Then** the system returns not found.

---

### User Story 3 — REST and MCP parity (Priority: P2)

The same lifecycle operations are available to HTTP clients and MCP agents.
Response fields match. Error codes match (`validation_error`, `not_found`).

**Why this priority**: ADR-0003 treats MCP as the domain contract; REST is
the parallel automation surface. Divergence is a product defect.

**Independent Test**: Contract tests for both transports covering success,
idempotent invalidate, rejected supersede of superseded, missing actor, and
unknown id.

**Acceptance Scenarios**:

1. **Given** the same invalidate inputs, **When** REST or MCP is used, **Then** the returned lore entry includes `verification_status`, `invalidated_by`, `invalidated_at`, and `superseded_by_id` with the same meaning.
2. **Given** a successful supersede, **When** REST or MCP is used, **Then** the response is the successor lore entry; GET of the predecessor shows `superseded_by_id` set.
3. **Given** MCP `tools/list`, **When** listed, **Then** exactly nine tools are advertised, including `memlore.invalidate` and `memlore.supersede`.
4. **Given** explain on predecessor after supersede, **When** audits are listed, **Then** `supersede` appears in chronological order after prior actions.

---

### Edge Cases

- Invalidating a verified entry succeeds; `verified_by` / `verified_at` remain; origin is unchanged.
- Verifying an invalidated or superseded entry is rejected with `validation_error` (cannot revive).
- Successor is unverified even if the predecessor was verified.
- Duplicate replacement statements in the same scope are allowed (new successor id), matching create semantics.
- Concurrent supersede of the same current entry: at most one successor wins; the other request is a validation error once the predecessor is already superseded.
- Search, knowledge_search, and get_for_task continue to return superseded and invalidated entries in v1 (filtering is F112).
- Graph-service fact invalidate/supersede is out of scope.

## Requirements

### Functional Requirements

- **FR-001**: System MUST allow an actor to invalidate a current (not superseded) lore entry, setting verification status to `invalidated` without deleting statement, scope, origin, evidence, or audits.
- **FR-002**: Invalidate MUST require a non-empty actor; missing actor is `validation_error`.
- **FR-003**: Invalidate of an unknown id MUST return `not_found`.
- **FR-004**: Re-invalidating an already-invalidated entry MUST be a no-op with no additional audit.
- **FR-005**: Invalidate of a superseded entry MUST return `validation_error`.
- **FR-006**: System MUST record `invalidated_by` and `invalidated_at` on first successful invalidate and append one `invalidate` audit.
- **FR-007**: System MUST allow an actor to supersede a current (not superseded, not invalidated) lore entry by creating a successor in the same scope with origin `human_authored` and a new id.
- **FR-008**: Predecessor MUST store `superseded_by_id` equal to the successor id after a successful supersede.
- **FR-009**: Supersede MUST append a `supersede` audit on the predecessor and a `create` audit on the successor in the same unit of work.
- **FR-010**: Supersede MUST require non-empty actor and non-empty statement; otherwise `validation_error`.
- **FR-011**: Supersede of unknown id MUST return `not_found`.
- **FR-012**: Supersede of an already-superseded or invalidated predecessor MUST return `validation_error` and MUST NOT create a successor.
- **FR-013**: Successor evidence is the request evidence if provided, otherwise empty; predecessor evidence is not copied.
- **FR-014**: REST MUST expose `POST /v1/lore-entries/{id}/invalidate` (actor header) and `POST /v1/lore-entries/{id}/supersede` (actor header; body statement + optional evidence).
- **FR-015**: MCP MUST expose `memlore.invalidate` and `memlore.supersede` with the same error format `{code}: {message}`.
- **FR-016**: Lore entry API shape MUST include `superseded_by_id` (null when current), `invalidated_by`, and `invalidated_at` on REST and MCP.
- **FR-017**: Invalidate REST response is the updated predecessor (`200`). Supersede REST response is the successor (`201`). MCP returns the same respective entries.
- **FR-018**: Verify MUST reject invalidated and superseded entries with `validation_error`.
- **FR-019**: Successor creation MUST emit the existing create/episode-ingest outbox event; invalidate and predecessor supersede MUST NOT introduce new outbox event types.
- **FR-020**: `memlore.explain` and REST audit list MUST include `invalidate` and `supersede` actions in chronological order.

### Key Entities

- **LoreEntry**: Engineering knowledge with statement, scope, origin,
  verification status (`unverified` | `verified` | `invalidated`), evidence,
  provenance timestamps, optional `superseded_by_id` (null = current),
  optional invalidate actor/time.
- **Successor LoreEntry**: New current entry created by supersede; same
  scope as predecessor; origin `human_authored`; unverified.
- **AuditRecord**: Append-only event with action `create` | `verify` |
  `invalidate` | `supersede`.

## Out of Scope

- Conflict detection and surfacing (F112)
- Filtering superseded/invalidated from `search`, `knowledge_search`, and `get_for_task` (F112)
- Graph-service fact invalidate/supersede
- New outbox event types for invalidate/supersede of the predecessor
- OIDC / RBAC (F111)
- Agent-authored origins on create/supersede
- Revive/re-verify of invalidated lore
- Deleting lore entries
- Changing successor scope independently of the predecessor

## Success Criteria

### Measurable Outcomes

- **SC-001**: An actor can invalidate a current entry and still retrieve the original statement, evidence, and full audit history including one invalidate event.
- **SC-002**: Re-invalidating the same entry does not add a second invalidate audit.
- **SC-003**: An actor can supersede a current entry and retrieve both predecessor (linked to successor) and successor (new statement, same scope).
- **SC-004**: Superseding an already-superseded or invalidated entry fails without creating another successor.
- **SC-005**: REST and MCP expose the same lore fields for the new lifecycle attributes and the same two error codes for these operations.
- **SC-006**: MCP advertises nine tools including invalidate and supersede.
- **SC-007**: Explain of a superseded predecessor lists create (and verify if any) then supersede in time order.

## Assumptions

- Mutating calls always pass an explicit actor; actor is never inferred from environment.
- `rejected` verification status from the authority-model table is not introduced in this slice (`invalidated` is the untrusted terminal verification state).
- Authority ranking and retrieval filtering of superseded/invalidated items is unchanged until F112.
- Self-verify remains allowed only for current, non-invalidated entries.
- Goose migration `00003` is additive and reversible.
- Python MCP contract tests that still advertise a five-tool Python server are out of this slice unless they are already running against Go.
