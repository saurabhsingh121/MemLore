# Graph Service API

MemLore knowledge-plane HTTP API. Go Core calls this service via the
`KnowledgeGraph` port — never Graphiti directly.

**OpenAPI**: [`graph-service/openapi.yaml`](../../graph-service/openapi.yaml)  
**Contract spec**: [`specs/009-graph-service/contracts/graph-service-api.md`](../../specs/009-graph-service/contracts/graph-service-api.md)

## Base URL

Local default: `http://127.0.0.1:8090` (`MEMLORE_GRAPH_SERVICE_URL`)

## Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/health` | Service + Neo4j connectivity |
| POST | `/episodes` | Ingest lore episode into graph |
| POST | `/search` | Semantic/graph search |
| GET | `/facts/{id}` | Retrieve fact by id |

## Episode ingest

```http
POST /episodes
Content-Type: application/json

{
  "statement": "Payment events must use the transactional outbox.",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "metadata": {},
  "provenance_refs": ["lore-entry-uuid"]
}
```

Response `201`:

```json
{ "episode_id": "..." }
```

## Search

```http
POST /search
Content-Type: application/json

{
  "query": "payment outbox",
  "scope": { "kind": "repository", "key": "github.com/acme/payments" },
  "limit": 10
}
```

Response `200`:

```json
{
  "results": [
    {
      "id": "...",
      "statement": "...",
      "score": 0.95,
      "scope": { "kind": "repository", "key": "github.com/acme/payments" },
      "provenance_refs": []
    }
  ]
}
```

Responses use MemLore field names only — no Graphiti types (`EntityEdge`,
`group_id`, etc.) in JSON.

## Configuration

| Variable | Default | Notes |
|----------|---------|-------|
| `MEMLORE_NEO4J_URI` | `bolt://localhost:7687` | Neo4j bolt URI |
| `MEMLORE_NEO4J_USER` | `neo4j` | Neo4j user |
| `MEMLORE_NEO4J_PASSWORD` | `memlore-dev-password` | Neo4j password |
| `OPENAI_API_KEY` | — | Required for Graphiti ingest/search |
| `MEMLORE_GRAPH_SERVICE_URL` | `http://127.0.0.1:8090` | Go client base URL |

## Go client

```go
client := graphclient.NewClient(os.Getenv("MEMLORE_GRAPH_SERVICE_URL"), nil)
episodeID, err := client.IngestEpisode(ctx, ports.EpisodeIngestRequest{...})
```

See `internal/application/ports/knowledge_graph.go` and
`internal/infrastructure/graphclient/`.

## Local development

```bash
docker compose up -d neo4j graph-service
curl -s http://127.0.0.1:8090/health
```

See [`specs/009-graph-service/quickstart.md`](../../specs/009-graph-service/quickstart.md).
