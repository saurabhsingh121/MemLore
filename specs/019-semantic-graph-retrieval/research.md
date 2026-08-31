# Research: Fuller semantic search + graph retrieval (F006)

## R1 — Governance relevance without embeddings

**Decision**: Case-insensitive match on `lore_entries.statement` using
significant query tokens (split on whitespace; ignore tokens shorter than 2
chars). An entry matches if the full query substring appears **or** at least
**one** significant token appears (OR). Rank by token hit count / full-phrase
bonus, then verified status, then `created_at` desc. Cap with `limit`.

**Rationale**: No new infra; works offline when graph is down; meets SC-001
for clear on-topic vs off-topic fixtures. OR matching keeps `get_for_task`
usable when the task text only partially overlaps lore statements. Graph
plane remains the semantic engine.

**Alternatives considered**:
- Postgres `tsvector` / `pg_trgm` — better recall; deferred (migration + ops).
- Embedding column in Postgres — out of scope per spec.
- Graph-only hydration (no text match) — fails when graph is down (SC-004).

## R2 — Scope-less governance (Q1=B)

**Decision**: When `scope` is omitted, run statement search with no scope
predicate, then filter entries through existing membership `CanAccessScope`
when enforcement is on. Local mode: no membership filter.

**Rationale**: Spec clarification Option B.

**Alternatives considered**: Keep governance empty without scope (F108) —
rejected by clarification. Require scope — breaking.

## R3 — Cross-plane dedupe (Q2=B)

**Decision**: Prefer governance. When `GraphFact.ProvenanceRefs` contains a
lore id that is accessible and temporal-allowed, attach additive
`graph_receipt` on that lore’s API object and omit the fact from
`graph.items`. Hydrate lore into governance if text-match missed it but
provenance found it.

**Rationale**: Postgres is SoT; agents get one primary hit with receipt.

**Alternatives considered**: Dual listing with link marker (Q2=A) — noisier.
Score-winner collapse (Q2=C) — less predictable for governance audits.

## R4 — Inaccessible provenance

**Decision**: If membership on and provenance lore is inaccessible, do not
hydrate that lore; omit the graph fact from `graph.items` when **all** its
provenance refs are inaccessible lore ids (avoid cross-tenant graph echo).
If provenance is empty or non-UUID/non-lore, keep the fact (graph-only).

**Rationale**: Align with F114 fail-closed / no tenant leak.

## R5 — No graph-service code changes

**Decision**: Leave `graph-service/` unchanged. Outbox already sets
`provenance_refs: [entry.ID]`.

**Rationale**: Boundary + simplicity; contract already MemLore-shaped.

## R6 — Presenter shape

**Decision**: Add optional `graph_receipt` on each governance `LoreEntry` in
knowledge-search responses only (or a knowledge-search-specific wrapper).
Prefer extending knowledge-search governance items via a thin DTO field
without breaking lore CRUD presenters — e.g. `KnowledgeGovernanceItem`
embeds lore fields + `graph_receipt`.

**Rationale**: Additive; CRUD list unchanged.

## R7 — sqlc query

**Decision**: Add `SearchLoreEntriesByStatement` with params: query pattern(s),
optional scope_kind/key (nullable), limit. Memory repo mirrors filter logic.

For multi-token AND, either multiple `ILIKE` clauses in SQL or fetch candidates
with first token and filter in Go. Prefer Go-side multi-token filter after a
broad `ILIKE '%' || $1 || '%'` on the longest token (or full query string as
single substring for v1 simplicity).

**v1 simplification**: Match if `statement` contains the full trimmed query as
case-insensitive substring **OR** (when query has multiple tokens) contains
every token. Implement in Go over candidates from:
- scoped `ListByScope` then filter (fine for typical scope sizes), OR
- new SQL `WHERE statement ILIKE '%' || $1 || '%'` with optional scope +
  `LIMIT` high-water then re-filter/rank in Go.

**Decision for impl**: Scoped path may list-by-scope then filter in Go (parity
with memory). Scope-less path needs SQL/memory search without loading all
rows unbounded — use `ILIKE` + `LIMIT` (e.g. limit*5) then rank/filter.

## R8 — Docs / tracker

**Decision**: Update `docs/api/rest.md`, `docs/api/mcp.md`, point to new
contract; mark F006 DONE in FEATURE_DEVELOPMENT; note F108 as superseded
partial.
