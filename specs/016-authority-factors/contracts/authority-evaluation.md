# Authority Evaluation Contract (F003)

## Compile / `memlore.get_for_task` item shape

Pipeline: retrieve → temporal filter → conflict detect → **evaluate + rank** →
budget.

Each `items[]` element:

```json
{
  "id": "uuid",
  "statement": "Use outbox for payments.",
  "source": "governance",
  "authority_score": 0.92,
  "trust_band": "canonical",
  "authority_factors": {
    "verification_status": "verified",
    "origin": "human_authored",
    "supersession_status": "current",
    "recency_boost": 0.08,
    "evidence_strength": 1.0,
    "source_type": "adr",
    "scope_match": 1.0
  },
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "evidence": [{ "type": "adr", "value": "0001-dual-plane-architecture" }],
  "provenance_refs": []
}
```

Graph item differences: `source` is `graph`; `trust_band` is `low`;
`authority_factors` includes `graph_score` and `source_type: "graph"`;
governance-only keys may be omitted.

`trust_band` is required. New factor keys are omitempty.

F112 fields (`conflicts`, current-only `items`) unchanged.

## `memlore.explain`

**Arguments**: `{ "id": "uuid-string" }` (unchanged)

**Success**: existing LoreEntry fields and chronological `audits`, plus:

```json
{
  "trust_band": "high",
  "authority_score": 0.82,
  "authority_factors": {
    "verification_status": "verified",
    "origin": "human_authored",
    "supersession_status": "current",
    "recency_boost": 0.05,
    "evidence_strength": 0.0,
    "source_type": "human_statement",
    "scope_match": 1.0
  },
  "factor_breakdown": [
    "verification_status=verified",
    "origin=human_authored",
    "supersession_status=current",
    "source_type=human_statement",
    "evidence_strength=0.00",
    "scope_match=1.00",
    "recency_boost=0.05",
    "trust_band=high"
  ]
}
```

No `summary` field. Unknown id → `not_found`. Stale entries are still
explained (`untrusted` if invalidated; superseded cannot be `canonical`).

## REST — `GET /v1/lore-entries/{id}/explain`

Same success JSON as MCP explain. `404` unknown id. Reader operation
(OIDC-on); local mode open like GET-by-id.

## Ranking order (contract tests)

In one compile set with recency-neutral fixtures:

1. verified + ADR evidence (canonical)
2. verified, no evidence (high)
3. unverified human (medium)
4. graph-only (low)
5. unverified agent_inference (low)
6. invalidated must not outrank (5) if it reaches the scorer

## MCP tools

Still exactly 9 tools. No new tool name.
