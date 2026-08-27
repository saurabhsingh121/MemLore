# Data Model: Graph Retrieval Orchestration (F108)

## SearchKnowledgeQuery (application input)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| query | string | yes | Semantic search text |
| scope | Scope | no | When set, triggers governance list |
| limit | int | no | Default 10; applied per plane |

## KnowledgeSearchResponse (API output)

| Field | Type | Notes |
|-------|------|-------|
| query | string | Echo of input |
| scope | Scope \| null | Echo when provided |
| governance.items | LoreEntry[] | Postgres lore entries |
| graph.items | GraphFact[] | Knowledge plane hits |
| warnings | string[] | e.g. `graph_service_unavailable` |

## GraphFact (knowledge plane item)

| Field | Type | Notes |
|-------|------|-------|
| id | string | Fact identifier |
| statement | string | Natural language fact |
| score | number | Relevance score (higher = better) |
| scope | Scope \| null | Optional scope filter echo |
| provenance_refs | string[] | Provenance references |

## Relationships

- `SearchKnowledgeHandler` → `ListLoreByScopeHandler` (governance, conditional)
- `SearchKnowledgeHandler` → `ports.KnowledgeGraph.Search` (knowledge, always)
- Presenters map `domain.LoreEntry` and `ports.GraphFact` to JSON shapes
