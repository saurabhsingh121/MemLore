# Research: F003 authority factor model + evaluation

## R1 — Persist vs ephemeral

**Decision**: Evaluate at compile/explain time from existing lore (and graph)
fields. No `authority_evaluations` table.

**Rationale**: Constitution forbids score-without-factors persistence; v1 has
no audit need for historical scores. Factors are fully determined by current
row fields + request scope + now. Caching would require invalidation on
verify/invalidate/supersede for little gain at retrieval limit 20.

**Rejected**: Snapshot table “for explain history” — out of scope unless a
later feature needs “what did the agent see then.”

## R2 — Pipeline order

**Decision**: retrieve → FilterCurrent → DetectConflicts → Evaluate+RankAndDedup
→ ApplyTokenBudget.

**Rationale**: Spec prefers filter-then-evaluate so stale never gets fancy
scores on the default compile path (cheaper + safer). Conflicts do not need
scores. Target-architecture.md currently lists authority *before* temporal
filter; update that doc to match this decision (F112 already filtered first).

**Safety**: Keep invalidated score cap (≤ 0.20) if a stale item reaches the
scorer (ranking unit test).

**Explain**: Evaluate the fetched entry even when stale (get/explain remain
unfiltered per F112).

## R3 — Scoring weights (deterministic)

**Decision**: Weighted sum, clamped to [0, 1].

Governance:

| Term | Rule |
|------|------|
| verification base | verified `0.72`; unverified `0.48`; invalidated `0.12` then **cap 0.20** |
| origin adj | `architecture_decision`/`human_verified` `+0.08`; `human_authored` `+0.05`; `imported_source` `+0.02`; `repository_observation` `0`; `agent_observation` `-0.16`; `agent_inference` `-0.22` |
| evidence adj | `0.10 * evidence_strength` (ADR `1.0`, other `0.6`, none `0.0`) |
| scope adj | `0.05 * scope_match` (exact `1.0`, kind-only `0.5`, none `0.0`) |
| recency | existing `0.10 * (1 - min(ageDays/365, 1))` |
| superseded | `-0.25` if superseded |

Then `score = clamp(sum, 0, 1)`; if invalidated, `score = min(score, 0.20)`.

Graph: `min(fact.Score * 0.45, 0.45)`.

**Ordering intent** (recency 0, exact scope):

| Case | Approx score | Band |
|------|--------------|------|
| verified human + ADR | 0.92 | canonical |
| verified human, no evidence | 0.82 | high |
| unverified human | 0.58 | medium |
| graph (score 0.99) | 0.445 | low |
| unverified agent_inference | 0.31 | low |
| invalidated | ≤ 0.20 | untrusted |

**Rejected**: Keep F109 bases `0.85`/`0.55` with only extra boosts — graph
could still outrank unverified human (cap 0.80). Spec requires unverified
human ≫ graph ≫ agent inference.

## R4 — Trust bands from factors, not score cut-points

**Decision**: First-match factor rules in spec FR-008. Score is derived for
ranking; band is independently assigned so a future weight tweak cannot
silently promote agent inference into `canonical`.

**Hard cap**: `agent_inference` / `agent_observation` never `canonical`;
`high` only when `verified`.

## R5 — source_type vs origin

**Decision**: Keep both. Origin is who produced the knowledge; source_type is
the evidence class (`adr`, `human_statement`, agent/repo/import/graph).

## R6 — source_reliability

**Decision**: Omit from v1 JSON. No stub constant that looks like a real
reputation score.

## R7 — Explain surface

**Decision**: Enrich `memlore.explain` in place (still 9 tools). Add REST
`GET /v1/lore-entries/{id}/explain` with the same payload for parity. No NL
`summary`. `factor_breakdown` is `[]string` of `key=value`.

**Rejected**: New MCP tool `memlore.authority` — violates tool-count constraint.

## R8 — Application vs domain split

**Decision**: `internal/domain/authority.go` owns types + `EvaluateAuthority(FactorInputs)`.
`internal/application/authority` maps `LoreEntry` / `GraphFact` + requested
scope + now. Ranking and explain call the application mapper.

**Rejected**: Keep scoring in `ranking.go` — that is the opaque v1 being
replaced. Rejected a large strategy framework.

## R9 — Schema / outbox / MCP tools / auth

**Decision**: No migration, no outbox events, no new MCP tools, no F111 changes.
Local-mode explain/compile remain open reads (OIDC-on: reader).
