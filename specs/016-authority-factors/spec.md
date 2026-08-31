# Feature Specification: Authority factor model + evaluation (F003)

**Feature Branch**: `016-authority-factors`  
**Created**: 2026-08-31  
**Status**: Ready  
**Depends on**: F109 (context compiler), F110 (lifecycle), F112 (temporal filter + conflicts), F111 (authz — keep separate)  
**Implements**: Product F003 (explainable authority factors, trust bands, ranking)  
**Input**: User description: "Make authority explainable and first-class — evaluate explicit factors, expose them on compile/explain, and use them for ranking/trust bands — without opaque-only scores."

## Goal

Replace ad-hoc verified/unverified ranking with an **Authority Evaluation** that
computes explainable factors for each candidate, derives a numeric score **and**
a discrete trust band from those factors, and surfaces both on compiled context
and on explain — so agents and humans can see *why* a fact is trusted, not only
a scalar.

F111 answers **who may act**. F003 answers **how trustworthy a fact is**. They
MUST stay separate.

## Clarifications

### Session 2026-08-31

Decisions encoded from the F003 implementation prompt. No remaining product
questions block planning.

- Q: Persist evaluations? → A: **Ephemeral at read/compile time** from existing
  lore (and graph) fields. No `authority_evaluations` table in v1. Never persist
  a score without its factors (constitution); v1 persists neither.
- Q: Trust-band definitions? → A: Five discrete bands — `canonical`, `high`,
  `medium`, `low`, `untrusted` — assigned from factors by deterministic rules
  (not from score cut-points). Exact rules are in Functional Requirements.
- Q: Pipeline order vs F112? → A: retrieve → temporal filter → conflict detect
  → authority evaluate + rank/dedup → token budget. Stale lore is not scored on
  the default compile path. Invalidated score cap remains as defense if a stale
  item reaches the scorer. Explain evaluates the fetched entry even when stale.
- Q: `source_reliability`? → A: **Omit from v1 JSON** (no historical reputation
  store). Do not invent a per-source history. Document as deferred.
- Q: `source_type` vs origin? → A: **Include** as a derived factor (ADR vs
  human statement vs agent vs graph vs import vs repo observation). Not
  redundant: origin is who produced it; source_type is the evidence class.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Explainable factors and trust band on compiled context (Priority: P1)

An agent compiles context for a task. Each packed item includes the factor
breakdown that produced its authority score and a discrete trust band. Ranking
follows those evaluations, not a hardcoded verified/unverified pair of bases.

**Why this priority**: Without this, agents still see an opaque-ish score and
cannot distinguish a verified ADR from a verified one-liner or an unverified
human note from graph text.

**Independent Test**: Compile a mix of verified-with-ADR, verified-without-
evidence, unverified human, graph-only, and (defense) invalidated fixtures;
assert factor keys, bands, and sort order.

**Acceptance Scenarios**:

1. **Given** a current verified entry with ADR evidence in the compile scope,
   **When** context is compiled, **Then** that item’s trust band is `canonical`
   (or `high` if it does not meet the canonical evidence rule), and
   `authority_factors` include origin, verification, supersession, recency,
   evidence_strength, source_type, and scope_match.
2. **Given** a current verified human-authored entry with no evidence, **When**
   context is compiled, **Then** its trust band is `high` and it ranks below a
   canonical ADR in the same set.
3. **Given** a current unverified human-authored entry, **When** context is
   compiled, **Then** its trust band is `medium`.
4. **Given** graph-only hits (no matching governance statement), **When**
   context is compiled, **Then** those items have trust band `low` and include
   `graph_score` among factors.
5. **Given** mixed candidates, **When** items are ranked, **Then** order is
   verified+strong evidence ≫ verified weaker ≫ unverified human ≫ graph ≫
   unverified agent inference (subject to token budget).
6. **Given** an invalidated entry that somehow reaches ranking, **When** scores
   are computed, **Then** it does not outrank an unverified current human entry
   and its band is `untrusted`.

---

### User Story 2 — Agent inference cannot silently gain human authority (Priority: P1)

An unverified agent inference (or agent observation) never appears in
`canonical` or `high` trust bands. Human verification can raise it to `high`
but never to `canonical`.

**Why this priority**: Constitution Principle V — agent inference MUST NOT
silently gain human authority.

**Independent Test**: Evaluate fixtures for unverified and verified
`agent_inference` / `agent_observation` and assert band caps.

**Acceptance Scenarios**:

1. **Given** a current unverified `agent_inference` entry, **When** it is
   evaluated (compile or explain), **Then** trust band is `low`.
2. **Given** a current unverified `agent_observation` entry, **When** it is
   evaluated, **Then** trust band is `low`.
3. **Given** a current **verified** `agent_inference` entry (even with ADR
   evidence), **When** it is evaluated, **Then** trust band is `high`, never
   `canonical`.
4. **Given** the same verified agent inference and a verified human ADR in one
   compile set, **When** ranked, **Then** the human ADR ranks above the agent
   inference.

---

### User Story 3 — Explain surfaces the same evaluation (Priority: P1)

A human or agent asks to explain a lore entry. The payload includes the
evaluated factors, trust band, numeric score, and a short structured factor
breakdown (labels, not a generated essay). Stale entries remain explainable.

**Why this priority**: Compile shows ranking receipts; explain is the audit
path for a single fact, including history.

**Independent Test**: Create, optionally verify/invalidate/supersede, then
explain; assert authority fields present and no natural-language summary
field.

**Acceptance Scenarios**:

1. **Given** an existing lore entry, **When** `memlore.explain` is called,
   **Then** the payload includes `trust_band`, `authority_score`,
   `authority_factors`, `factor_breakdown`, plus existing entry fields and
   chronological audits, and MUST NOT include a generated `summary` essay.
2. **Given** a superseded or invalidated entry, **When** explain is called,
   **Then** the entry is still returned; band reflects stale state
   (`untrusted` if invalidated; superseded cannot be `canonical`).
3. **Given** an unknown id, **When** explain is called, **Then** `not_found`
   (unchanged).
4. **Given** REST `GET /v1/lore-entries/{id}/explain`, **When** the entry
   exists, **Then** the payload matches MCP explain (parity).

---

### User Story 4 — Compile pipeline and F112 remain intact (Priority: P2)

Default compile still omits superseded and invalidated items, still surfaces
conflict groups, still dedups graph against governance, and still applies
token budget. Authority evaluation runs **after** the temporal filter on the
compile path.

**Why this priority**: F003 must not re-litigate F112.

**Independent Test**: Existing F112 compile fixtures still omit stale items
and list conflicts; new factor/band fields appear only on remaining items.

**Acceptance Scenarios**:

1. **Given** current, superseded, and invalidated lore in one scope, **When**
   context is compiled, **Then** `items` contain only current governance (plus
   graph hits); superseded/invalidated are absent from `items`.
2. **Given** two disagreeing current statements in one scope, **When** context
   is compiled, **Then** a conflict group still lists both sides.
3. **Given** a graph fact whose normalized statement matches governance,
   **When** ranked, **Then** the graph duplicate is still dropped.

---

### Edge Cases

- Graph facts have no origin/verification: band `low`; factors expose
  `graph_score`; other governance-only keys omitted (`omitempty`).
- Compile with exact scope match vs an entry in a different key: `scope_match`
  is lower than exact; ranking may still include it if retrieval returned it.
- Explain without a compile scope: `scope_match` is treated as exact against
  the entry’s own scope (the fact is “in its own scope”).
- Empty evidence vs ADR vs URL/path-only: evidence_strength distinguishes
  strong (ADR present) from moderate (any other evidence) from none.
- Recency: brand-new entries get the maximum recency contribution; entries
  older than one year get none.
- Tie scores: stable sort by statement (existing ranking behavior).
- `source_reliability` absent from payloads (deferred).
- Local-mode auth (F111) unchanged; explain/compile remain reader operations.
- No new MCP tools; tool count stays at 9.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST evaluate an Authority Evaluation for each compile
  candidate (governance entry or graph fact) from explicit factors — never
  from an opaque model or stored scalar alone.
- **FR-002**: v1 factors MUST include: `origin`, `verification_status`,
  `supersession_status` (`current` | `superseded`), `recency` (as
  `recency_boost`), `evidence_strength`, `source_type`, `scope_match`, and
  `graph_score` when the item is from the graph plane.
- **FR-003**: `source_reliability` MUST NOT appear in v1 payloads (deferred
  historical reputation).
- **FR-004**: `source_type` MUST be derived as: `graph` for graph-plane items;
  `adr` if origin is `architecture_decision` or any evidence type is `adr`;
  otherwise map origin to `human_statement` (`human_authored`,
  `human_verified`), `agent_observation`, `agent_inference`,
  `repo_observation` (`repository_observation`), or `import`
  (`imported_source`).
- **FR-005**: `evidence_strength` MUST be `1.0` if any evidence type is `adr`,
  `0.6` if any other evidence is present, else `0.0`.
- **FR-006**: `scope_match` MUST be `1.0` when the entry scope equals the
  requested compile scope; `0.5` when kind matches but key differs, or when
  evaluating explain against the entry’s own scope (N/A mismatch); `0.0` when
  the item has no usable scope (typical graph-only).
- **FR-007**: Recency contribution MUST remain a 0.00–0.10 boost that decays
  linearly to zero at 365 days (existing recency behavior).
- **FR-008**: Trust bands MUST be assigned from factors using this first-match
  order:
  1. `untrusted` — verification is `invalidated`
  2. `low` — graph-plane item
  3. `low` — origin is `agent_inference` or `agent_observation` AND
     verification is not `verified`
  4. `high` — origin is `agent_inference` or `agent_observation` AND
     verification is `verified` (never `canonical`)
  5. `medium` — entry is superseded (not invalidated); cannot be `canonical`
     or `high`
  6. `canonical` — current, verified, and (`source_type` is `adr` OR
     `evidence_strength` is `1.0`), and origin is not agent_*
  7. `high` — current and verified
  8. `medium` — current, unverified, and origin is human-side
     (`human_authored`, `human_verified`, `architecture_decision`,
     `imported_source`)
  9. `low` — otherwise (e.g. unverified `repository_observation`)
- **FR-009**: Unverified `agent_inference` / `agent_observation` MUST NOT
  receive `canonical` or `high`. Verified agent origin MUST NOT receive
  `canonical`.
- **FR-010**: Ranking score MUST be derived from the same factors (weighted
  sum, clamped to [0, 1]) such that the order in US1 scenario 5 holds.
  Invalidated scores MUST be capped at `0.20` (plus no recency above that cap)
  so they cannot outrank unverified human lore.
- **FR-011**: Graph-plane scores MUST be capped at `0.45` so they cannot
  outrank unverified human-side current lore (typical ≥ `0.50`) and typically
  outrank unverified agent inference.
- **FR-012**: Compile pipeline MUST be: retrieve → temporal filter (current
  only) → conflict detect → evaluate + rank/dedup → token budget. Do not pack
  stale items into `items`.
- **FR-013**: Context packet items MUST include `trust_band` and the factor
  object (new keys `omitempty`). Existing `authority_score` remains.
- **FR-014**: `memlore.explain` MUST include `trust_band`, `authority_score`,
  `authority_factors`, and `factor_breakdown` (array of short `key=value`
  strings). MUST NOT add a generated narrative `summary`.
- **FR-015**: REST MUST expose the same explain payload at
  `GET /v1/lore-entries/{id}/explain` (reader; local-mode auth unchanged).
- **FR-016**: Evaluations MUST be computed at read time. v1 MUST NOT add an
  authority persistence table or store score without factors.
- **FR-017**: MCP tool count MUST remain 9 (enrich `memlore.explain` and
  `memlore.get_for_task` payloads only).
- **FR-018**: F110 transition rules, F112 conflict definition, F112 temporal
  defaults, F111 auth, and graph-service fact lifecycle MUST remain unchanged.

### Key Entities

- **Authority Evaluation**: Ephemeral result for one candidate — score, trust
  band, factor set, short breakdown. Not stored.
- **Trust Band**: One of `canonical`, `high`, `medium`, `low`, `untrusted`.
- **Factor Set**: Named explainable inputs listed in FR-002.
- **Context Item**: Compiled packet row; gains `trust_band` and richer
  `authority_factors`.
- **Explain Result**: Existing entry + audits; gains evaluation fields.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a golden fixture of at least six origin/verification/evidence
  combinations, every combination produces the specified trust band with 100%
  agreement (deterministic, no LLM).
- **SC-002**: In a mixed compile of verified ADR, unverified human, graph-only,
  and unverified agent inference, ranking order matches US1 scenario 5 on
  every run.
- **SC-003**: 100% of compiled `items` in default compile include `trust_band`
  and at least origin or `graph_score` among factors (governance vs graph).
- **SC-004**: 100% of successful explain calls (MCP and REST) include
  `trust_band`, `authority_factors`, and `factor_breakdown` without a
  `summary` essay field.
- **SC-005**: Existing F112 behaviors still hold: stale omitted from default
  compile `items`; conflicts still listed; get/explain still return history.
- **SC-006**: An unverified agent inference never appears as `canonical` or
  `high` in any golden test.

## Assumptions

- Recency clock is “now” at evaluation time (compile handler clock; explain
  uses current time).
- Governance create path still only allows `human_authored` origin; other
  origins exist on the domain model for evaluation and future ingest.
- Graph plane has no verification/origin; treating it as `low` is correct for
  v1.
- No schema migration is required.
- Token budget and graph dedup stay as F109.
- Documentation updates (`authority-model`, API docs, feature tracker,
  specify-rules next feature) are part of this work.

## Out of Scope

- F010 team/project membership
- Full ML/LLM authority judgment
- Changing F110 invalidate/supersede transition rules
- Changing F112 conflict definition or temporal filter defaults
- Graph-service fact lifecycle / graph-side stale filter
- Opaque score-only storage
- Historical `source_reliability` reputation system
- New MCP tools
- Persisted evaluation snapshots / audit of past scores
