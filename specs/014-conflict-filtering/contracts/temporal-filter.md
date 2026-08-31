# Contract: Temporal filtering (F112)

Retrieval defaults to **current** lore only. History remains via get/explain.

## Current definition

Entry is current iff not superseded (`superseded_by_id` unset) **and**
`verification_status` ≠ `invalidated`.

## Affected operations (default)

| Operation | Stale omitted? |
|-----------|----------------|
| `GET /v1/lore-entries?scope_kind=&scope_key=` | Yes |
| `memlore.search` | Yes |
| `POST /v1/knowledge-search` (governance items) | Yes |
| `memlore.knowledge_search` (governance items) | Yes |
| `POST /v1/context/compile` items | Yes |
| `memlore.get_for_task` items | Yes |

## Unaffected operations

| Operation | Behavior |
|-----------|----------|
| `GET /v1/lore-entries/{id}` | Returns entry including stale |
| `memlore.get` | Returns entry including stale |
| `memlore.explain` | Returns entry including stale |

## `include_stale` (list / search / knowledge_search)

| Field | Type | Default |
|-------|------|---------|
| include_stale | boolean | false |

When `true`, superseded and invalidated governance entries are included.
Compile / `get_for_task` do **not** accept packing stale into `items`.

### REST examples

```http
GET /v1/lore-entries?scope_kind=repository&scope_key=r1
GET /v1/lore-entries?scope_kind=repository&scope_key=r1&include_stale=true
```

```json
POST /v1/knowledge-search
{
  "query": "deploy rules",
  "scope": { "kind": "repository", "key": "r1" },
  "include_stale": false
}
```

### MCP

`memlore.search` and `memlore.knowledge_search` accept optional
`include_stale` (boolean, default false).

## Graph plane

Graph items are not filtered for supersession/invalidation in v1.
