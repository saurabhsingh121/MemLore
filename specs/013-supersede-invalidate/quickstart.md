# Quickstart: F110 invalidate + supersede

Requires Postgres if exercising persistence; contract tests use in-memory UoW.

```bash
# Unit + contract (no Postgres)
go test ./...

# Optional integration
go test -tags=integration ./internal/infrastructure/postgres/ ./internal/adapters/http/

# Migration
./bin/memlore migrate
```

## Manual REST

```bash
# create
curl -s -X POST http://localhost:8080/v1/lore-entries \
  -H 'Content-Type: application/json' \
  -H 'X-Memlore-Actor: alice' \
  -d '{"statement":"Use outbox","scope":{"kind":"repository","key":"r1"}}'

# invalidate
curl -s -X POST http://localhost:8080/v1/lore-entries/{id}/invalidate \
  -H 'X-Memlore-Actor: alice'

# supersede (on a different current entry)
curl -s -X POST http://localhost:8080/v1/lore-entries/{id}/supersede \
  -H 'Content-Type: application/json' \
  -H 'X-Memlore-Actor: alice' \
  -d '{"statement":"Use transactional outbox for all events"}'
```

## Manual MCP

`memlore mcp` must list nine tools. Call `memlore.remember`, then
`memlore.supersede` / `memlore.invalidate`, then `memlore.explain`.
