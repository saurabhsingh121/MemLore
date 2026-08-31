# Data Model: Knowledge search enrichment (F006 remainder)

No new PostgreSQL tables. Uses existing `lore_entries` and graph facts via port.

## Governance candidate

| Field | Source | Notes |
|-------|--------|-------|
| LoreEntry | `lore_entries` | Unchanged schema |
| relevance | derived | Token/substring match vs query |
| graph_receipt | derived | Optional; from linked GraphFact |

## Graph fact (port)

| Field | Notes |
|-------|-------|
| ID, Statement, Score, Scope, ProvenanceRefs | Existing `ports.GraphFact` |

## Graph receipt (API additive)

| Field | Type | Notes |
|-------|------|-------|
| id | string | Graph fact id |
| statement | string | Fact statement |
| score | float | Graph score |
| provenance_refs | string[] | As returned by graph |

Attached onto a knowledge-search governance item when that lore id appears in
`provenance_refs` and the lore is returned as primary.

## SearchKnowledgeResult (application)

| Field | Notes |
|-------|-------|
| Query, Scope, Warnings | Unchanged |
| Governance | `[]GovernanceHit` — LoreEntry + optional receipt |
| Graph | Facts **not** collapsed into receipts |

## Repository port addition

```text
SearchRelevant(ctx, opts) ([]LoreEntry, error)

opts:
  Query string
  Scope *domain.Scope   # nil = all scopes (caller applies membership)
  Limit int             # high-water or final; handler may re-limit
  IncludeStale bool     # or filter after fetch like ListByScope
```

Memory + Postgres implement. ListByScope remains for exact list (`memlore.search`).

## Authz / temporal

Unchanged F114 membership and F112 stale filtering applied after candidate
fetch / before response.
