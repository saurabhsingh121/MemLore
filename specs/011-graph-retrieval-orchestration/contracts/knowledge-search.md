# Knowledge Search Contract (F108)

MemLore-shaped dual-plane search. Response MUST NOT contain Graphiti-specific keys
(`group_id`, `EntityEdge`, etc.).

## REST — `POST /v1/knowledge-search`

### Request

```json
{
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "limit": 10
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| query | string | yes | Non-empty search text |
| scope | object | no | `{ kind, key }`; triggers governance list when set |
| limit | integer | no | Default 10 |

### Response `200`

```json
{
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "governance": {
    "items": []
  },
  "graph": {
    "items": [
      {
        "id": "fact-uuid",
        "statement": "Payment events use transactional outbox.",
        "score": 0.92,
        "scope": { "kind": "repository", "key": "github.com/acme/payments" },
        "provenance_refs": []
      }
    ]
  },
  "warnings": []
}
```

- `scope` is `null` when omitted from request.
- `governance.items` uses `LoreEntry` shape (same as lore CRUD).
- `graph.items` may be empty; `governance.items` may be empty.
- `warnings` may include `graph_service_unavailable` when graph-service fails but
  governance succeeded.

### Errors

Same envelope as lore REST (`validation_error`, `internal_error`).

## MCP — `memlore.knowledge_search`

### Arguments

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| query | string | yes | Search text |
| scope | object | no | `{ kind, key }` |
| limit | integer | no | Default 10 |
| actor_id | string | yes | Validated non-empty (audit future) |

### Success payload

Identical JSON object to REST `200` response (structured content).

### Errors

`validation_error` for missing query or blank `actor_id`.

## Unchanged — `memlore.search`

Exact scope list only (`{ items: LoreEntry[] }`). Not semantic/graph search.
