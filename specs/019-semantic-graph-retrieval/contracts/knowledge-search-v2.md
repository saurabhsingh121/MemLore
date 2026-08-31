# Knowledge Search Contract (F006 remainder / v2)

Extends F108 [`011-graph-retrieval-orchestration/contracts/knowledge-search.md`](../../011-graph-retrieval-orchestration/contracts/knowledge-search.md).
Additive fields only; required F108 fields remain.

## REST — `POST /v1/knowledge-search`

### Request

Unchanged shape:

```json
{
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "limit": 10,
  "include_stale": false
}
```

| Field | Notes |
|-------|-------|
| query | Required; used for **both** planes |
| scope | Optional. When **omitted**, governance searches membership-allowed lore (local: all). When **set**, governance limited to that scope (plus authz). |
| limit | Default 10; applies to each plane after relevance/merge |
| include_stale | Default false |

### Response `200`

```json
{
  "query": "payment outbox",
  "scope": null,
  "governance": {
    "items": [
      {
        "id": "lore-uuid",
        "statement": "Payment events use transactional outbox.",
        "scope": { "kind": "repository", "key": "github.com/acme/payments" },
        "origin": "human_authored",
        "verification_status": "verified",
        "evidence": [],
        "created_by": "alice",
        "created_at": "2026-08-01T00:00:00Z",
        "graph_receipt": {
          "id": "fact-uuid",
          "statement": "Payment events use transactional outbox.",
          "score": 0.92,
          "provenance_refs": ["lore-uuid"]
        }
      }
    ]
  },
  "graph": {
    "items": [
      {
        "id": "fact-other",
        "statement": "Related graph-only observation.",
        "score": 0.81,
        "scope": { "kind": "repository", "key": "github.com/acme/payments" },
        "provenance_refs": []
      }
    ]
  },
  "warnings": []
}
```

### Semantics

1. **Governance relevance**: Items must match the query (statement contains
   query / all significant tokens). Unrelated scope lore MUST NOT appear.
2. **Scope omitted**: Governance may be non-empty; membership filtered when
   enforcement on.
3. **`graph_receipt`**: Optional; present when a graph fact was collapsed onto
   this lore. That fact MUST NOT also appear in `graph.items`.
4. **Graph-only facts**: Remain in `graph.items` (empty or non-lore provenance).
5. **Warnings**: `graph_service_unavailable` unchanged.
6. **Authz / stale**: F114 + F112 unchanged.

### Errors

Unchanged envelopes (`validation_error`, `forbidden`, `unauthorized`, …).

## MCP — `memlore.knowledge_search`

Same success payload as REST. Arguments unchanged (`query`, optional `scope`,
`limit`, `include_stale`, `actor_id`).

## Unchanged — `memlore.search`

Exact scope list only. Not semantic.
