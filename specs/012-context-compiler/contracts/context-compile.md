# Context Compile Contract (F109 + F112)

Compiled context packet for agents. No Graphiti-specific keys.

Pipeline: retrieve → temporal filter (current only) → conflict detect →
authority evaluate + rank/dedup → token budget.

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
      "trust_band": "high",
      "authority_factors": {
        "verification_status": "verified",
        "origin": "human_authored",
        "supersession_status": "current",
        "recency_boost": 0.08,
        "evidence_strength": 0.0,
        "source_type": "human_statement",
        "scope_match": 1.0
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
  "warnings": [],
  "conflicts": []
}
```

- `items` contain **current** governance only (never superseded/invalidated).
- Each item includes `trust_band` and explainable `authority_factors` (F003).
- `conflicts` is always present (`[]` when none). Each group:
  `{ scope, entry_ids, statements }` for disagreeing current statements in one
  scope within the retrieval set.
- Conflict sides are not dropped; budget may exclude an id from `items` while
  still listing it in `conflicts`.

See also: [`specs/014-conflict-filtering/contracts/conflict-detection.md`](../../014-conflict-filtering/contracts/conflict-detection.md)
and [`specs/016-authority-factors/contracts/authority-evaluation.md`](../../016-authority-factors/contracts/authority-evaluation.md).

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

## Related

- `memlore.knowledge_search` — dual-plane search; optional `include_stale` (F108/F112)
- `memlore.search` — exact scope list; optional `include_stale` (F112)
