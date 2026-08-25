# Quickstart: Graph Service (F106)

## Prerequisites

- Docker Compose (Neo4j)
- `OPENAI_API_KEY` for Graphiti ingest/search
- uv 0.5+

## Local dev

```bash
# Start Neo4j (+ optional graph-service container)
docker compose up -d neo4j

# Or run graph-service locally:
cd graph-service
uv sync
export MEMLORE_NEO4J_URI=bolt://localhost:7687
export MEMLORE_NEO4J_USER=neo4j
export MEMLORE_NEO4J_PASSWORD=memlore-dev-password
export OPENAI_API_KEY=sk-...
uv run uvicorn graph_service.adapters.http.app:create_app --factory --host 127.0.0.1 --port 8090
```

## Smoke test

```bash
curl -s http://127.0.0.1:8090/health | jq

curl -s -X POST http://127.0.0.1:8090/episodes \
  -H 'Content-Type: application/json' \
  -d '{"statement":"Use outbox for payments.","scope":{"kind":"repository","key":"github.com/acme/payments"}}' | jq

curl -s -X POST http://127.0.0.1:8090/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"payment outbox"}' | jq
```

## Tests

```bash
cd graph-service
uv run pytest                    # unit + contract (no Neo4j)
uv run pytest -m integration     # requires Neo4j + OPENAI_API_KEY
```

Go contract test (integration tag):

```bash
export MEMLORE_GRAPH_SERVICE_URL=http://127.0.0.1:8090
go test -tags=integration ./internal/infrastructure/graphclient/...
```

## Docker Compose full stack

```bash
export OPENAI_API_KEY=sk-...
docker compose up -d neo4j graph-service
curl -s http://127.0.0.1:8090/health
```
