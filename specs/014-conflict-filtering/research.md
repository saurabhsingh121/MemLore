# Research: F112 temporal filtering + conflict detection

## R1 — Where to filter (SQL vs application)

**Decision**: Filter at the application/query layer (`ListLoreByScopeHandler`
and compile). Keep Postgres `ListByScope` unfiltered.

**Rationale**: Spec requires `include_stale` for operators and unfiltered
get-by-id. One application default avoids forked SQL paths and keeps
memory-UoW tests simple. Volume of lore per scope is small in v1.

**Rejected**: SQL `WHERE superseded_by_id IS NULL AND verification_status <>
'invalidated'` as the only path — complicates include_stale and duplicates
logic across sqlc + memory repos.

## R2 — Single filter choke point

**Decision**: `ListLoreByScopeHandler` accepts `IncludeStale bool` (default
false via zero value). Search and HTTP/MCP list pass the flag through.
`CompileContextHandler` always requests current-only (ignores any stale flag)
and re-filters governance before ranking as defense in depth.

**Rationale**: Spec: “do not fork two inconsistent defaults.” List and search
share the handler. Compile never packs stale into `items`.

## R3 — Conflict definition

**Decision**: Structural conflict = ≥2 **current** governance entries in the
**same scope** within the retrieval set with **different**
`NormalizeStatement` values. Identical statements are not conflicts.

**Rationale**: Constitution requires surfacing conflicts without LLM. Matches
F001 duplicate-statement allowance. Cheap and explainable.

**Rejected**: NLI/LLM contradiction; persisted `conflict_records`; auto-resolve.

## R4 — Conflict attachment surface

**Decision**: Required on compile / `get_for_task` (`conflicts: []`). Optional
on knowledge_search if cheap — **include empty/omitted for v1 on
knowledge_search** to minimize contract churn; compile is the agent-preferred
path (ADR 0003).

**Rationale**: Agents should use `get_for_task`. Listing conflicts on raw
search is nice-to-have; compile is mandatory.

## R5 — Ranking interaction

**Decision**: Pipeline: retrieve → FilterCurrent → DetectConflicts →
RankAndDedup → ApplyTokenBudget. Conflict groups reference entry ids from the
post-filter set even if budget drops an item from `items`.

**Rationale**: Target architecture order. Budget must not erase conflict
awareness.

**Safety**: If invalidated somehow reaches `governanceScore`, treat base score
≤ unverified (regression test). Prefer filter so invalidated never ranks on
default path. Optional `supersession_status: "current"` factor only when
useful; omitempty OK — default packet only has current.

## R6 — Graph plane

**Decision**: No graph-side stale filter. No required cross-plane warning when
graph text matches stale governance.

**Rationale**: Graph invalidate/supersede deferred; inventing staleness would
lie. Nice-to-have warning left out of v1 acceptance.

## R7 — Schema / outbox / MCP tools

**Decision**: No migration, no new outbox events, no new MCP tools (remain 9).

**Rationale**: Filtering/compile-only feature; F110 already stores lifecycle
fields.
