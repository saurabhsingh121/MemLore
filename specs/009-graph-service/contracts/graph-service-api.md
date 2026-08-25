# REST Contract: Graph Service

Base path: `/`  
Content-Type: `application/json`

Error envelope:

```json
{
  "error": {
    "code": "validation_error|not_found|service_unavailable|internal_error",
    "message": "human-readable summary"
  }
}
```

## GET /health

**Response 200**:

```json
{
  "status": "ok",
  "service": "graph-service",
  "neo4j": "ok"
}
```

`neo4j` may be `unavailable`; `status` may be `degraded` when Neo4j is down.

## POST /episodes

Ingest a lore episode into the knowledge graph.

**Request**:

```json
{
  "statement": "string (1..8000)",
  "scope": { "kind": "repository", "key": "github.com/org/repo" },
  "metadata": {},
  "provenance_refs": ["lore-entry-uuid"],
  "reference_time": "2026-08-25T12:00:00Z"
}
```

`metadata`, `provenance_refs`, `reference_time` optional.

**Responses**:
- `201` → `EpisodeResponse`
- `400` → validation_error
- `503` → service_unavailable (Neo4j / Graphiti unavailable)

### EpisodeResponse

```json
{
  "episode_id": "string"
}
```

## POST /search

Semantic / graph search with MemLore-shaped results.

**Request**:

```json
{
  "query": "string",
  "scope": { "kind": "repository", "key": "github.com/org/repo" },
  "limit": 10
}
```

`scope` and `limit` optional (`limit` default 10, max 50).

**Responses**:
- `200` → `SearchResponse`
- `400` → validation_error
- `503` → service_unavailable

### SearchResponse

```json
{
  "results": [
    {
      "id": "string",
      "statement": "string",
      "score": 0.95,
      "scope": { "kind": "repository", "key": "github.com/org/repo" },
      "provenance_refs": []
    }
  ]
}
```

**Contract rule**: Response JSON MUST NOT contain Graphiti-specific keys such as
`EntityEdge`, `uuid`, `group_id`, `fact_embedding`, or `episodes`.

## GET /facts/{id}

Retrieve a single fact by id.

**Responses**:
- `200` → `GraphFactResponse` (same shape as search result item)
- `404` → not_found
- `503` → service_unavailable

## Go KnowledgeGraph Port

See `internal/application/ports/knowledge_graph.go` — mirrors ingest, search,
get fact, health. No Graphiti imports.
