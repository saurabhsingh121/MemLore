# Contract: Conflict detection on compile (F112)

## Definition

A **conflict group** exists when two or more **current** governance entries in
the **same scope**, present in the same compile retrieval set, have **different**
normalized statements (lowercase + trim). Identical statements are not a conflict.

Conflicts are **surfaced**, never used to drop a side. Ranking and token budget
still run; the conflict group retains all entry ids even if budget excludes one
from `items`.

## Response shape — compile / get_for_task

Add `conflicts` to the ContextPacket (REST + MCP identical):

```json
{
  "task": "...",
  "query": "...",
  "scope": { "kind": "repository", "key": "r1" },
  "items": [ /* ranked, budgeted; current only */ ],
  "meta": { /* unchanged */ },
  "warnings": [],
  "conflicts": [
    {
      "scope": { "kind": "repository", "key": "r1" },
      "entry_ids": ["uuid-a", "uuid-b"],
      "statements": ["Use blue-green deploys", "Use rolling deploys"]
    }
  ]
}
```

| Field | Type | Notes |
|-------|------|-------|
| conflicts | array | Always present; `[]` when none |
| conflicts[].scope | object | `{ kind, key }` |
| conflicts[].entry_ids | string[] | All ids in the disagreeing current set for that scope |
| conflicts[].statements | string[] | Distinct statements as stored |

## Pipeline order

1. Retrieve (F108)
2. Temporal filter (current only)
3. Detect conflicts
4. RankAndDedup
5. Token budget

## Unchanged

- No new MCP tools (still 9)
- knowledge_search v1 need not include `conflicts` (compile is preferred agent path)
- F110 invalidate/supersede rules unchanged
