# Context Packet Contract (F021, extends F007)

Compiled task packet for agents. Additive fields only. No Graphiti-specific
keys.

Pipeline: retrieve (task/query + ticket + files; repository briefing-class
list merge) → temporal filter (current only) → conflict detect → authority
rank/dedup → token budget → classify packet sections → omit empty sections.

F007 contract: [`specs/012-context-compiler/contracts/context-compile.md`](../../012-context-compiler/contracts/context-compile.md).

## REST — `POST /v1/context/compile`

### Request

```json
{
  "task": "Implement payment outbox handler",
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "token_budget": 4096,
  "branch": "feat/outbox",
  "ticket": "PAY-1842",
  "changed_files": ["src/payments/outbox.go"],
  "working_files": ["src/payments/publisher.go"],
  "agent_id": "cursor-agent"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| task | string | yes | Agent task description |
| query | string | no | Search text; defaults to `task` |
| scope | object | yes | `{ kind, key }` |
| token_budget | integer | no | Default 4096 |
| branch | string | no | Echoed; not used for retrieval filter |
| ticket | string | no | Included in search text |
| changed_files | string[] | no | Included in search text |
| working_files | string[] | no | Included in search text |
| agent_id | string | no | Echoed; never an authority factor |

V1 bodies with only `task` + `scope` (+ optional `query` / `token_budget`)
MUST still return `200` when valid.

Authz: read + membership on `scope` (unchanged).

### Response `200`

```json
{
  "task": "Implement payment outbox handler",
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "branch": "feat/outbox",
  "ticket": "PAY-1842",
  "changed_files": ["src/payments/outbox.go"],
  "working_files": ["src/payments/publisher.go"],
  "agent_id": "cursor-agent",
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
        "source_type": "human_statement"
      },
      "scope": { "kind": "repository", "key": "github.com/acme/payments" },
      "evidence": [{ "type": "adr", "value": "ADR-023" }],
      "provenance_refs": []
    }
  ],
  "sections": [
    {
      "id": "architecture",
      "items": [
        {
          "id": "uuid-arch",
          "statement": "Hexagonal architecture with ports.",
          "source": "governance",
          "authority_score": 0.9,
          "trust_band": "high",
          "authority_factors": {},
          "scope": { "kind": "repository", "key": "github.com/acme/payments" },
          "evidence": [],
          "provenance_refs": []
        }
      ]
    }
  ],
  "sources": [{ "type": "adr", "value": "ADR-023" }],
  "meta": {
    "token_budget": 4096,
    "estimated_tokens": 120,
    "items_included": 1,
    "items_total_ranked": 3,
    "unclassified_count": 0
  },
  "warnings": [],
  "conflicts": []
}
```

Notes:

- `items` always present (possibly `[]`); membership is the budgeted ranked list.
- `sections` always present as an array; only objects with non-empty `items`.
  Empty section ids (`conventions`, `gotchas`, …) are omitted entirely.
  Valid packet ids: `architecture`, `decisions`, `conventions`, `task_context`, `gotchas`.
- `sources` omitted when empty (`omitempty`). Echo fields omitted when empty.
- `conflicts` always present (`[]` when none). No `observed_drift` / `stale` keys.
- `meta.unclassified_count` is additive (0 when all budgeted items classified).

## MCP — `memlore.get_for_task`

### Arguments

| Field | Type | Required |
|-------|------|----------|
| task | string | yes |
| query | string | no |
| scope | object | yes |
| token_budget | integer | no |
| branch | string | no |
| ticket | string | no |
| changed_files | string[] | no |
| working_files | string[] | no |
| agent_id | string | no |
| actor_id | string | yes in local mode (unchanged) |
| access_token | string | OIDC |

Success payload identical to REST `200`. Tool count remains 10.

## CLI — `memlore context`

```text
memlore context --task "Implement outbox" --repository github.com/acme/payments
memlore context --task "Implement outbox" --repository github.com/acme/payments \
  --query "outbox" --ticket PAY-1842 --branch feat/outbox \
  --changed-file src/payments/outbox.go --working-file src/payments/publisher.go \
  --token-budget 2048 --agent-id cursor-agent
```

Prints a text briefing (task header + section headings + statements +
conflicts). Exit non-zero on validation or connectivity errors.
