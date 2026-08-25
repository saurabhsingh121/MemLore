# Data Model: F106 Graph Service

Knowledge plane entities (Neo4j via Graphiti — not PostgreSQL).

## Episode (ingest)

| Field | Type | Notes |
|-------|------|-------|
| episode_id | string | Returned by API; used as Graphiti episode name |
| statement | string | Primary lore text |
| scope | Scope | kind + key |
| metadata | object | Optional JSON object |
| provenance_refs | string[] | Lore entry ids, ADR refs, etc. |
| reference_time | datetime | Optional; defaults to ingest time |

## Scope

| Field | Type |
|-------|------|
| kind | enum aligned with governance (team, repository, …) |
| key | string |

Stored as Graphiti `group_id` = `{kind}:{key}`.

## GraphFact (search / get result)

| Field | Type | Notes |
|-------|------|-------|
| id | string | Edge/fact uuid |
| statement | string | Human-readable fact text |
| score | float | Search relevance 0–1 |
| scope | Scope | Optional if group known |
| provenance_refs | string[] | When available |

## Health

| Field | Type |
|-------|------|
| status | `ok` \| `degraded` |
| neo4j | `ok` \| `unavailable` |
| service | `graph-service` |
