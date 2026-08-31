# Testing

MemLore Core tests live next to Go packages (`internal/.../*_test.go`).
Graph-service tests live in `graph-service/tests/`.

Prefer many unit tests, fewer integration tests, and only necessary e2e tests.

```bash
go test ./...
go test -tags=integration ./...   # needs Postgres + migrate

cd graph-service && uv run pytest -m "not integration"
```
