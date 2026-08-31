# Feature Specification: Fuller semantic search + graph retrieval (F006 remainder)

**Feature Branch**: `019-semantic-graph-retrieval`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Depends on**: F108 (`011-graph-retrieval-orchestration`), F112 temporal filter,
F114 membership authz, F003 authority (for ranking signals where applicable)  
**Implements**: Product F006 remainder (close PARTIAL → DONE)  
**Input**: User description: "F006 remainder — fuller semantic search + graph
retrieval: make knowledge search query-aware on the governance plane, link and
dedupe across governance and graph results, and finish docs so agents get
relevant dual-plane answers with receipts — not a full scope dump beside
unrelated graph hits."

## Clarifications

### Session 2026-09-01

- Q: Scope-less governance? → A: When no scope is provided, search governance
  across **all membership-allowed scopes** (authn/membership on); in local mode
  (authn off), search across accessible lore without membership filtering
  (Option B).
- Q: Cross-plane duplicates? → A: **Prefer the governance lore entry** as the
  primary hit; attach the graph fact as a **linked receipt** on that entry (or
  equivalent additive receipt field). Do **not** also list that fact as an
  unexplained standalone graph item (Option B). Graph-only facts (no usable
  governance provenance) still appear under `graph.items`.

## Goal

F108 shipped the dual-plane **read path** (`knowledge_search` / REST). Today,
when a scope is provided, governance returns **every** current lore entry in
that scope regardless of the query text, while the graph plane searches
semantically. Agents therefore see noisy governance lists and duplicated or
unlinked graph facts.

This feature makes dual-plane search **query-relevant and cross-plane coherent**:
governance candidates must relate to the query; graph facts that already have
governance receipts are linked (and not double-presented); membership and
temporal rules already in force continue to apply. Completing this marks
product **F006 DONE**.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Query-relevant governance results (Priority: P1)

An agent asks a knowledge search about “transactional outbox” within a
repository scope that has many lore entries. Only governance entries that
relate to that question appear (or rank above unrelated ones), up to the
requested limit — not the entire scope dump.

**Why this priority**: Query-blind scope listing is the largest remaining gap
between “search” and useful agent context.

**Independent Test**: Seed a scope with one on-topic and several off-topic
current lore entries; run knowledge search with the on-topic query; assert
off-topic entries are omitted (or ranked below / truncated past limit) and the
on-topic entry is present.

**Acceptance Scenarios**:

1. **Given** scope S with lore A matching query Q and lore B unrelated to Q,
   **When** knowledge search runs with Q and S, **Then** governance items
   include A and exclude B (or B is not among the returned limited set).
2. **Given** scope S with no lore related to Q, **When** knowledge search runs,
   **Then** `governance.items` is an empty array (not an error) while graph
   search may still return facts.
3. **Given** the same request with `include_stale=true`, **When** a related
   superseded entry matches Q, **Then** it may appear; with default
   `include_stale=false`, stale related entries remain omitted.

---

### User Story 2 — Cross-plane linking and deduplication (Priority: P1)

When a graph fact was derived from (or references) a governance lore entry,
the agent can see that link. The same statement is not presented twice as
unrelated hits from both planes.

**Why this priority**: Dual-plane value depends on receipts and non-redundant
answers; F108 explicitly deferred this.

**Independent Test**: Seed lore L; ingest/search so a graph fact carries
provenance to L; run knowledge search; assert linkage and that L is not
double-counted as an unexplained duplicate.

**Acceptance Scenarios**:

1. **Given** lore L and a graph fact F with provenance referencing L,
   **When** knowledge search returns both planes, **Then** F (and/or L)
   exposes a clear link between them (e.g. matched lore id / provenance).
2. **Given** F and L represent the same underlying knowledge for the query,
   **When** results are returned, **Then** L appears in governance with an
   attached graph receipt for F, and F is **not** also listed as a standalone
   unexplained item in `graph.items`.
3. **Given** a graph fact with no governance provenance, **When** search runs,
   **Then** it still appears under graph items (graph-only knowledge remains
   valid).

---

### User Story 3 — Stable dual-plane contract with better ranking (Priority: P2)

REST and MCP knowledge search keep the familiar dual-plane envelope so existing
clients do not break, while ordering within each plane (and any optional unified
view) prefers higher relevance and, where available, higher authority/trust.

**Why this priority**: Agents and docs already teach the F108 shape; fuller
search should deepen quality without a breaking rename.

**Independent Test**: Contract tests prove response envelope compatibility;
unit tests prove ordering rules for governance and graph sections.

**Acceptance Scenarios**:

1. **Given** identical inputs, **When** REST `POST /v1/knowledge-search` and
   MCP `memlore.knowledge_search` run, **Then** success payloads share the same
   dual-plane fields (`query`, `scope`, `governance`, `graph`, `warnings`) plus
   any additive enrichment fields defined by this feature.
2. **Given** multiple governance matches, **When** results are returned,
   **Then** they are ordered by query relevance first, with authority/trust as
   a documented tie-break when scores are equal.
3. **Given** graph-service is down and scope S has query-relevant lore,
   **When** search runs, **Then** governance still returns those relevant
   items and `warnings` includes `graph_service_unavailable` (F108 behavior
   preserved).

---

### User Story 4 — Scope-less and authz-aware search (Priority: P2)

An agent issues a knowledge search with only a natural-language query (no
scope). Governance still returns query-relevant lore from every scope the
principal may read (membership-allowed when enforcement is on; unrestricted
by membership in local mode). Disallowed scopes never leak.

**Why this priority**: Agents often know the question before the exact scope
key; scope-less search is part of closing F006.

**Independent Test**: With membership on, subject in team A only; seed
on-topic lore in team A and team B; scope-less search returns A’s hit and
omits B’s.

**Acceptance Scenarios**:

1. **Given** no `scope` on the request, **When** knowledge search runs,
   **Then** governance may include query-relevant lore from multiple allowed
   scopes (not forced empty), and graph search still runs.
2. **Given** membership enforcement on, **When** search would otherwise
   include lore from a disallowed scope, **Then** those items are omitted
   (same isolation guarantees as F114).
3. **Given** local mode (authn off), **When** scope-less search runs,
   **Then** governance relevance search is not membership-filtered.

---

### Edge Cases

- Empty or whitespace-only query → validation error (unchanged).
- Limit ≤ 0 → default limit applies (unchanged default of 10 unless docs say
  otherwise).
- Graph returns facts whose provenance lore was deleted or inaccessible →
  omit the link / omit the fact from the caller’s view per authz (no existence
  leak for inaccessible lore).
- Very large scopes → governance results still respect `limit` after
  relevance filtering (not “return all then truncate blindly” without
  relevance).
- Local mode (authn off) → membership filtering off; relevance rules still
  apply.
- `memlore.search` / exact scope list tools → unchanged (not semantic).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Knowledge search MUST treat the query string as meaningful for
  **both** planes when selecting governance candidates (not only for graph
  search).
- **FR-002**: When a scope is provided, governance items returned MUST be
  limited to entries that are relevant to the query (relevance rule documented
  in plan/research), capped by `limit`.
- **FR-003**: Knowledge search MUST preserve F108 graceful degradation: graph
  failure yields `graph_service_unavailable` without failing governance when
  governance succeeded.
- **FR-004**: Cross-plane links MUST be exposed when a graph fact’s provenance
  references a returned (or returnable) lore entry.
- **FR-005**: When a graph fact links to an accessible lore entry L, knowledge
  search MUST present L as the primary governance hit with the graph fact
  attached as a linked receipt, and MUST NOT also list that fact as a
  standalone unexplained `graph.items` entry.
- **FR-006**: REST and MCP knowledge-search success envelopes MUST remain
  backward-compatible for existing required fields; new fields MUST be additive.
- **FR-007**: `memlore.search` MUST remain exact scope listing (not semantic).
- **FR-008**: Temporal defaults (`include_stale`, current-only governance) and
  F114 membership filtering MUST continue to apply to knowledge search.
- **FR-009**: Product documentation for knowledge search MUST describe
  relevance, linking/dedupe, warnings, and scope-less behavior so F006 docs
  checklist is complete.
- **FR-010**: When `scope` is omitted, governance MUST search query-relevant
  lore across membership-allowed scopes (or all lore in local mode), capped by
  `limit`, while graph search continues as today.
- **FR-011**: Go core MUST NOT expose Graphiti-specific types or field names in
  agent-facing responses (existing boundary).
- **FR-012**: This feature MUST NOT require changing the MCP tool count or
  renaming `memlore.knowledge_search`.

### Key Entities

- **Knowledge search request**: Query text, optional scope, limit, include_stale.
- **Governance hit**: Lore entry selected as query-relevant under authz +
  temporal rules.
- **Graph hit**: Knowledge-plane fact with score, optional scope, provenance.
- **Cross-plane link**: Association between a graph hit and a lore entry via
  provenance (and any dedupe marker).
- **Search warning**: Non-fatal condition (e.g. graph unavailable).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a fixture scope with ≥5 off-topic and 1 on-topic current lore
  entries, knowledge search for the on-topic question returns the on-topic
  entry and **zero** off-topic governance entries in the default limited
  result set (automated test).
- **SC-002**: When a graph fact references lore L via provenance, automated
  tests observe an explicit link in the knowledge-search response in ≥95% of
  seeded linked fixtures (all fixtures in the suite).
- **SC-003**: Existing F108 contract fields remain present; additive fields do
  not break current REST/MCP contract tests updated for compatibility.
- **SC-004**: With graph-service unavailable, agents still receive
  query-relevant governance results plus the standard unavailable warning
  (automated unit test).
- **SC-005**: F006 is marked **DONE** in the feature tracker with docs ✓ after
  this ships.
- **SC-006**: Membership isolation tests for knowledge search remain green
  (no cross-tenant governance leakage).

## Assumptions

- Graph-service semantic search remains the knowledge-plane engine; this
  remainder focuses on **orchestration quality** in MemLore Core (relevance,
  linking, dedupe, docs), not replacing Graphiti.
- Relevance for governance may use text matching and/or graph-led provenance
  hydration; exact algorithm is a plan/research decision, not a stakeholder
  wording concern.
- Authority factor scores may be used as tie-breakers when already available
  on the compile/explain path; inventing a new persisted ranking store is out
  of scope.
- Episode ingest already can carry lore ids in provenance; workers need not be
  redesigned unless provenance is missing for new writes (fix-forward only
  unless a small ingest gap is found).
- No new MCP tools; no change to OIDC role matrix.

## Out of Scope

- Turning `memlore.search` into semantic search
- Redesigning `memlore.get_for_task` / context compiler token budgeting (F007)
- New Neo4j schemas or Graphiti fork features beyond what graph-service already
  exposes (unless a thin contract field is required for provenance)
- Full-text search product UI / browser console
- Embedding store inside PostgreSQL as a mandatory v1 dependency
- Changing Apache-2.0 licensing or brand assets
- Graph-service OIDC

## Dependencies

- F108 dual-plane orchestrator and contracts
- F107 outbox ingest (lore → graph provenance path)
- F112 temporal filtering on governance items
- F114 membership filtering on knowledge search
- F003 authority evaluation available for optional tie-breaks
