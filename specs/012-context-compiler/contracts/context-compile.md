# Context Compile Contract (F109)

Compiled context packet for agents. No Graphiti-specific keys.

## REST — `POST /v1/context/compile`

### Request

```json
{
  "task": "Implement payment outbox handler",
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "token_budget": 4096
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| task | string | yes | Agent task description |
| query | string | no | Search text; defaults to `task` |
| scope | object | yes | `{ kind, key }` |
| token_budget | integer | no | Default 4096 |

### Response `200`

```json
{
  "task": "Implement payment outbox handler",
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "items": [
    {
      "id": "uuid",
      "statement": "Use outbox for payments.",
      "source": "governance",
      "authority_score": 0.92,
      "authority_factors": {
        "verification_status": "verified",
        "origin": "human_authored",
        "recency_boost": 0.08
      },
      "scope": { "kind": "repository", "key": "github.com/acme/payments" },
      "evidence": [],
      "provenance_refs": []
    }
  ],
  "meta": {
    "token_budget": 4096,
    "estimated_tokens": 120,
    "items_included": 1,
    "items_total_ranked": 3
  },
  "warnings": []
}
```

## MCP — `memlore.get_for_task`

### Arguments

| Field | Type | Required |
|-------|------|----------|
| task | string | yes |
| query | string | no |
| scope | object | yes |
| token_budget | integer | no |
| actor_id | string | yes (validated; audit future) |

Success payload identical to REST `200`.

## Unchanged

- `memlore.knowledge_search` — raw dual-plane search (F108)
- `memlore.search` — exact scope list
