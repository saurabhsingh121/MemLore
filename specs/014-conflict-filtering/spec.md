# Feature Specification: Temporal filtering + conflict detection (F112)

**Feature Branch**: `014-conflict-filtering`  
**Created**: 2026-08-31  
**Status**: Ready  
**Depends on**: F110 (invalidate + supersede), F109 (context compiler), F108 (knowledge search)  
**Implements**: Product F009 (conflict/temporal filtering); completes deferred F109 compiler stages  
**Input**: User description: "Filter superseded and invalidated lore out of retrieval, and detect/surface conflicts among remaining current lore — without deleting history."

## Goal

Ensure agents and operators retrieve **currently trustworthy** lore by default,
while preserving full history for get/explain. When multiple current entries in
the same scope disagree, surface that disagreement explicitly so agents never
silently act on only one side.

## Clarifications

### Session 2026-08-31

Decisions encoded from the F112 implementation prompt. No remaining product
questions block planning.

- Q: What is “current” lore? → A: An entry is current iff `superseded_by_id` is
  unset **and** verification status is not `invalidated`.
- Q: Default list/search behavior? → A: Exclude superseded and invalidated from
  scope list, knowledge search (governance hits), and compiled context items.
- Q: Operator access to stale lore in lists? → A: Optional `include_stale` flag
  on search / knowledge_search / list-by-scope; default `false`. `get_for_task`
  never packs stale items into `items` even if a parallel search flag exists.
- Q: Conflict definition (v1)? → A: Structural only — two or more **current**
  governance entries in the **same scope** that appear in the same retrieval set
  and have **different** normalized statements (lowercase + trim). Identical
  duplicate statements are **not** conflicts. No LLM/NLI.
- Q: What happens to conflicting sides? → A: Neither is dropped. Both remain
  candidates for ranking and budget packing. Conflict metadata lists all entry
  ids (and statements) even if budget excludes one side from `items`.
- Q: Persist conflicts? → A: No — ephemeral metadata on the response only.
- Q: Graph-plane staleness? → A: Out of scope (graph-service has no
  invalidate/supersede yet). Do not invent graph-side filtering.
- Q: Filter placement? → A: Application/query layer for retrieval defaults;
  persistence list-all-in-scope remains available for `include_stale` / admin.
  GET-by-id and explain stay unfiltered.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Stale lore stays out of agent retrieval (Priority: P1)

An agent asks for lore in a scope where some entries were superseded or
invalidated. The agent receives only current entries. Historical entries remain
fetchable by id and via explain so humans can audit why knowledge changed.

**Why this priority**: Without temporal filtering, F110 lifecycle is invisible
to agents — they still act on wrong or obsolete rules.

**Independent Test**: Create current, superseded, and invalidated entries in one
scope; search / knowledge_search / get_for_task omit the stale ones; get and
explain still return them.

**Acceptance Scenarios**:

1. **Given** a scope with current, superseded, and invalidated lore, **When** an
   agent searches that scope (default), **Then** only current entries appear.
2. **Given** the same mix, **When** knowledge search runs (default), **Then**
   the governance results omit superseded and invalidated entries.
3. **Given** the same mix, **When** context is compiled for a task, **Then**
   packet `items` contain only current governance entries (plus any graph hits).
4. **Given** a superseded or invalidated entry id, **When** get or explain is
   called, **Then** the full historical entry (and explain payload) is returned.
5. **Given** `include_stale=true` on list/search/knowledge_search, **When** the
   query runs, **Then** superseded and invalidated entries are included for
   operator inspection.

---

### User Story 2 — Conflicting current lore is surfaced (Priority: P1)

Two current rules in the same scope disagree. When an agent compiles context
for a task that retrieves both, the packet includes both candidates (subject to
budget) and an explicit conflict group naming both sides. Nothing is silently
discarded.

**Why this priority**: Constitution requires conflicts preserved and surfaced;
silent single-winner selection would violate authority/temporal correctness.

**Independent Test**: Remember two current entries with different statements in
one scope; compile context; confirm a conflict group lists both ids/statements
and both remain eligible in ranking (neither auto-dropped).

**Acceptance Scenarios**:

1. **Given** two current governance entries in the same scope with different
   normalized statements present in one retrieval set, **When** context is
   compiled, **Then** the packet includes a conflict group with both entry ids
   and statements, and neither side is removed solely because of the conflict.
2. **Given** two current entries in the same scope with identical normalized
   statements, **When** context is compiled, **Then** no conflict group is
   created for that pair (duplicates are not contradictions).
3. **Given** a conflict group where token budget packs only one side into
   `items`, **When** the packet is returned, **Then** the conflict group still
   lists both entry ids so the agent knows a rival exists.
4. **Given** no disagreeing current statements in the retrieval set, **When**
   context is compiled, **Then** `conflicts` is an empty list.

---

### User Story 3 — Compile pipeline order is correct (Priority: P2)

Compiled context follows retrieve → temporal filter → conflict detect →
rank/dedup → budget. Invalidated or superseded lore never outranks current
unverified lore on the default path.

**Why this priority**: Matches target architecture and prevents ranking bugs
from reintroducing stale knowledge.

**Independent Test**: Unit tests on filter + conflict helpers and compile
handler with mixed fixtures prove order and ranking safety.

**Acceptance Scenarios**:

1. **Given** a retrieval set mixing current, superseded, and invalidated lore,
   **When** context is compiled (default), **Then** ranking input contains only
   current entries.
2. **Given** a regression fixture that would otherwise expose invalidated lore
   to ranking, **When** default compile runs, **Then** invalidated lore does
   not appear in items and does not outrank unverified current lore.
3. **Given** graph-service warnings already produced by retrieval, **When**
   compile completes, **Then** those warnings remain and conflicts are added
   separately (empty array when none).

---

### User Story 4 — REST and MCP share semantics (Priority: P2)

Operators and agents see the same filtering and conflict behavior whether they
use REST or MCP. No new MCP tools are added.

**Why this priority**: Parity is required by the MCP domain interface ADR and
existing F108/F109 contracts.

**Independent Test**: Contract tests for list-by-scope, knowledge_search,
compile / get_for_task, search, get, and explain.

**Acceptance Scenarios**:

1. **Given** identical inputs, **When** REST or MCP list/search/compile is
   invoked, **Then** current-only filtering and conflict metadata match the
   shared contracts.
2. **Given** the nine existing MCP lore tools, **When** tools are listed,
   **Then** the tool count remains nine (no new tool for this feature).

### Edge Cases

- Superseded predecessor and current successor in the same scope: only the
  successor appears in default retrieval; predecessor remains via get/explain.
- Invalidated entry that was never superseded: omitted from default retrieval;
  still gettable.
- Multiple conflict groups across different scopes in one multi-scope retrieval:
  each scope with ≥2 disagreeing current statements yields its own group
  (v1 retrieval is typically single-scope; behavior must still be correct).
- More than two disagreeing statements in one scope: one conflict group listing
  all distinct-statement entry ids in that scope within the retrieval set.
- Graph facts that textually resemble superseded governance statements: no
  automatic graph filtering in v1 (optional warning is nice-to-have, not
  required).
- Empty retrieval set: empty items, empty conflicts, existing warnings only.
- `include_stale` on get_for_task: not supported for packing stale into items;
  compile always uses current-only items.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST treat a lore entry as **current** only when it is not
  superseded and not invalidated.
- **FR-002**: Default scope list, search, and knowledge-search governance
  results MUST omit superseded and invalidated entries.
- **FR-003**: Default compiled context `items` MUST omit superseded and
  invalidated governance entries.
- **FR-004**: System MUST support an `include_stale` option (default false) on
  list-by-scope / search / knowledge_search so operators can include historical
  entries; get_for_task MUST NOT pack stale entries into `items`.
- **FR-005**: Get-by-id and explain MUST continue to return superseded and
  invalidated entries unchanged.
- **FR-006**: System MUST detect conflict groups among **current** governance
  entries in the same scope within a retrieval set when normalized statements
  differ.
- **FR-007**: System MUST NOT treat identical normalized statements as a
  conflict.
- **FR-008**: System MUST NOT drop either side of a conflict solely because a
  conflict was detected; ranking and token budget still apply afterward.
- **FR-009**: Compiled context responses MUST include a `conflicts` array
  (empty when none). Each group MUST identify scope, entry ids, and statements.
- **FR-010**: When budget excludes a conflicting side from `items`, the
  conflict group MUST still list that entry id.
- **FR-011**: Compile pipeline MUST apply temporal filter before conflict
  detection, and conflict detection before rank/dedup and token budget.
- **FR-012**: REST and MCP MUST share filtering and conflict response semantics
  for the affected operations; no new MCP tool in v1.
- **FR-013**: Graph-plane results MUST NOT be filtered for supersession or
  invalidation in v1 (graph lifecycle not available).
- **FR-014**: Invalidate and supersede transition rules from F110 MUST remain
  unchanged; this feature is retrieval/compile behavior only.

### Key Entities

- **Current lore entry**: Governance lore that is neither superseded nor
  invalidated; eligible for default retrieval and conflict consideration.
- **Stale lore entry**: Superseded and/or invalidated lore; history preserved;
  excluded from default retrieval; still fetchable by id / explain.
- **Conflict group**: Ephemeral grouping of two or more current governance
  entries in one scope with disagreeing normalized statements within one
  retrieval/compile response. Not a persisted record in v1.
- **Context packet**: Compiled agent response including items, meta, warnings,
  and conflicts.

## Out of Scope

- OIDC / RBAC (F111)
- Graph-service fact invalidate / supersede
- New outbox event types
- Persisted conflict entities or conflict-resolution workflow
- LLM / NLI contradiction detection
- Full authority factor model (source_reliability, evidence_strength, etc.)
- Changing F110 invalidate/supersede transition semantics
- Auto-resolving conflicts or picking a winner

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a fixture with current + superseded + invalidated lore, default
  search, knowledge search, and get_for_task omit 100% of the stale entries from
  result lists / packet items.
- **SC-002**: Get and explain still return stale entries for the same fixture
  (100% fetch success for known ids).
- **SC-003**: When two current disagreeing statements exist in one scope,
  compile responses include a conflict group naming both sides in ≥95% of
  automated contract cases (target: all cases).
- **SC-004**: Neither conflicting side is removed solely due to conflict
  detection; both remain ranking candidates.
- **SC-005**: Operators can opt into stale list results via `include_stale`
  without changing get_for_task packing behavior.
- **SC-006**: Agents and operators observe identical filtering/conflict
  semantics on REST and MCP for the covered operations.
- **SC-007**: No new MCP tool is required; existing preferred agent path
  (`get_for_task`) surfaces conflicts.

## Assumptions

- F110 fields (`superseded_by_id`, `verification_status=invalidated`,
  invalidate actor/time) are already persisted and correct.
- Statement normalization for conflict detection reuses the same lowercase+trim
  rule already used for compile-time dedup.
- v1 conflict detection is structural text disagreement within a scope, not
  semantic contradiction.
- Knowledge-search may include conflict metadata when inexpensive; compiled
  context MUST include it.
- Persistence “list all in scope” can remain unfiltered at the repository
  boundary; application/query layer owns the default current-only filter.
- Optional cross-plane warning (graph statement matching stale governance) is
  nice-to-have and not required for acceptance.
- Product tracking: F007 moves closer to complete (temporal/conflict stages no
  longer deferred); F009 is implemented by this feature.
