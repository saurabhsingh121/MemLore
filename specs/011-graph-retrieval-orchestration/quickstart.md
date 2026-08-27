# Quickstart: Graph Retrieval Orchestration (F108)

## Prerequisites

```bash
./bin/memlore migrate
docker compose up -d postgres neo4j graph-service
export OPENAI_API_KEY=...
```

## Start services

```bash
./bin/memlore serve &
./bin/memlore worker &
```

## Create lore (outbox → graph)

```bash
curl -s -X POST http://127.0.0.1:8080/v1/lore-entries \
  -H 'Content-Type: application/json' -H 'X-Memlore-Actor: alice' \
  -d '{"statement":"Use outbox for payments.","scope":{"kind":"repository","key":"github.com/acme/payments"}}'
```

Wait for worker + Graphiti indexing, then:

## Knowledge search (REST)

```bash
curl -s -X POST http://127.0.0.1:8080/v1/knowledge-search \
  -H 'Content-Type: application/json' \
  -d '{"query":"payment outbox","scope":{"kind":"repository","key":"github.com/acme/payments"}}' | jq
```

## Knowledge search (MCP)

Call `memlore.knowledge_search` with the same arguments (plus `actor_id`).

## Tests

```bash
go test ./...
go test -tags=integration ./internal/application/queries/...
```
