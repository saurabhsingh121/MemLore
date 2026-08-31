# Quickstart: F006 semantic graph retrieval

## Prerequisites

- Postgres migrated (`memlore migrate`)
- Optional: graph-service + Neo4j for live graph plane
- Branch `019-semantic-graph-retrieval`

## Verify relevance (unit)

```bash
go test ./internal/application/queries/ -run Knowledge -count=1
```

## Manual smoke (local mode)

```bash
go run ./cmd/memlore serve
```

Create two lore entries in the same scope — one about “outbox”, one about
“unrelated topic”. Then:

```bash
curl -s localhost:8080/v1/knowledge-search \
  -H 'content-type: application/json' \
  -d '{"query":"outbox","scope":{"kind":"team","key":"platform"},"limit":10}'
```

Expect only the outbox lore under `governance.items`.

## Scope-less

```bash
curl -s localhost:8080/v1/knowledge-search \
  -H 'content-type: application/json' \
  -d '{"query":"outbox","limit":10}'
```

Expect governance hits without requiring `scope` (local mode).

## Receipt collapse

With worker + graph-service processing outbox, search a statement that was
ingested; expect `graph_receipt` on the lore and that fact absent from
`graph.items`.
