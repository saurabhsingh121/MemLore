# Research: Agent Context Bootstrap (F021)

## Decision 1 — Extend compile, do not add a new tool or store

**Decision**: Enrich `CompileContextHandler`, REST `POST /v1/context/compile`,
and MCP `memlore.get_for_task`. Keep MCP tool count at 10. Do not persist
packets.

**Rationale**: Agents already call `get_for_task`. A second tool would split
the entry point. A stored packet would drift from lore (same reason F020 is
compile-on-read).

**Alternatives rejected**: New `memlore.bootstrap` tool; packet cache table;
breaking replacement of `items[]` with sections-only.

## Decision 2 — Reuse F020 classifier; add `task_context` mapping

**Decision**: Call existing `ClassifyItem` (same cues and first-match order).
Packet sections are a subset: `architecture`, `decisions`, `conventions`,
`gotchas`, plus `task_context` for leftover task-relevant items. Do not invent
a second classification system.

**Rationale**: Spec requires one section model. F020 already maps engineering
language to stable ids.

**Alternatives rejected**: LLM labeling; a generic `other` dump section;
reordering F020 `ClassificationOrder`.

## Decision 3 — Retrieval merge: task search ∪ briefing-class lore

**Decision**: Search query = query-or-task plus ticket and file path strings.
When scope kind is `repository` and a lore lister is wired, list current lore
for that scope and keep entries that `ClassifyItem` as architecture,
decisions, conventions, or gotchas. Union with search hits, then existing
`FilterCurrent` → `DetectConflicts` → `RankAndDedup` → `ApplyTokenBudget`.
Non-repository scopes stay search-only (v1).

**Rationale**: Narrow task queries would otherwise miss high-authority
architecture/decisions. Merging *all* repository lore would flood the budget
with unclassified noise. Filtering list hits to the four briefing classes
keeps the merge small and honest.

**Alternatives rejected**: (a) Search-only (misses briefing). (b) Full
list-by-scope unfiltered (F023-scale volume). (c) Changing authority scores
with a task-relevance boost (would mix query match into trust; F023 can add
a drop ladder later). Agent identity is never a rank key.

## Decision 4 — Keep `items[]`; additive `sections` and `sources`

**Decision**: `items` remains the budgeted ranked list (F007 contract).
`sections` is an additive classified view of those items. `sources` is the
unique evidence refs from included items, omitted when empty. Optional request
fields are echoed when non-empty (`omitempty`). `meta.unclassified_count` is
additive.

**Rationale**: Existing clients must keep working. Sections are the new agent
UX; items remain the machine-stable list.

## Decision 5 — Omit drift and stale sections in v1

**Decision**: No `observed_drift` or `stale` packet sections. Conflicts stay
on the existing `conflicts` array. No `include_stale` on compile.

**Rationale**: F050 does not exist; inventing drift from files would violate
constitution IX/VI. Default retrieval already omits stale; resurrecting it
into a briefing would fight F112.

## Decision 6 — CLI `memlore context`

**Decision**: Developer command `memlore context --task … --repository …`
with optional flags matching the additive compile fields. Same local DSN
wiring as `memlore profile`.

**Rationale**: Constitution requires CLI for DX. MCP/REST-only would leave
humans without a briefing surface.

## Decision 7 — Defer F022/F023

**Decision**: No `profile` query field and no new section drop-order ladder.
Document default behavior as F021 compile (authority rank + token budget),
not a coding workflow profile.

**Rationale**: User asked to prefer deferring F022/F023 if they expand scope.
