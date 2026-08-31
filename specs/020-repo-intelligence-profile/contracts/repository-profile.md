# Repository Profile Contract (F020)

Compiled on-read briefing. No Graphiti-specific keys.

Pipeline: retrieve (overview query, repository scope) → temporal filter →
conflict detect → authority rank/dedup → token budget → classify into
sections → omit empty sections.

## REST — `POST /v1/repository-profile`

### Request

```json
{
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "token_budget": 4096
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| scope | object | yes | `{ kind, key }`; kind MUST be `repository` |
| token_budget | integer | no | Default 4096 |

Authz: same as `POST /v1/context/compile` (read + membership).

### Response `200`

```json
{
  "repository": { "kind": "repository", "key": "github.com/acme/payments" },
  "sections": [
    {
      "id": "decisions",
      "items": [
        {
          "id": "uuid",
          "statement": "Use Kafka instead of RabbitMQ.",
          "source": "governance",
          "authority_score": 0.92,
          "trust_band": "canonical",
          "authority_factors": {},
          "scope": { "kind": "repository", "key": "github.com/acme/payments" },
          "evidence": [{ "type": "adr", "value": "ADR-017" }],
          "provenance_refs": []
        }
      ]
    }
  ],
  "meta": {
    "token_budget": 4096,
    "estimated_tokens": 120,
    "items_included": 1,
    "items_total_ranked": 3,
    "unclassified_count": 2
  },
  "warnings": [],
  "conflicts": []
}
```

- `sections` omitted keys: only present section objects; never empty `items`.
- `conflicts` always present (`[]` when none).
- Validation errors: missing scope, non-repository kind, empty key.

## MCP — `memlore.repo_profile`

Read-only. Arguments:

| Field | Required | Notes |
|-------|----------|-------|
| scope | yes | `{ kind, key }` |
| token_budget | no | Default 4096 |
| actor_id | local read | Same as `get_for_task` |
| access_token | OIDC | Bearer when OIDC on |

Structured content matches REST 200 body.

## CLI — `memlore profile`

```text
memlore profile --repository github.com/acme/payments
memlore profile --repository github.com/acme/payments --token-budget 2048
```

Prints a text briefing (repository header + section headings + statements).
Exit non-zero on validation or connectivity errors.
