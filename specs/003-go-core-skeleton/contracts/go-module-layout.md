# Contract: Go Module Layout (F101)

**Feature**: `003-go-core-skeleton`  
**Module**: `github.com/memlore/memlore`  
**Go version**: 1.25+

## Required paths (F101)

| Path | Requirement |
|------|-------------|
| `go.mod` | Module root; `go 1.25` directive |
| `go.sum` | Present after `go mod tidy` |
| `sqlc.yaml` | Points at `db/queries/` and Postgres engine |
| `cmd/memlore/main.go` | Runnable entrypoint (version/health stub OK) |
| `migrations/*.sql` | goose-format; includes `00001_lore_audit` |
| `db/queries/*.sql` | At least one query for sqlc smoke |
| `internal/domain/` | Package exists; no infra imports |
| `internal/application/` | Package exists |
| `internal/adapters/` | Package exists |
| `internal/infrastructure/postgres/` | Package exists; may hold sqlc output |

## Forbidden in F101

- HTTP lore handlers (F104)
- MCP server (F105)
- `graph-service/` implementation
- Empty package directories without `.go` files
- Domain imports of `net/http`, `github.com/jackc/pgx`, MCP SDK

## CI contract

`.github/workflows/ci.yml` MUST include:

```yaml
- go test ./...
- go vet ./...
```

Python jobs (ruff, mypy, pytest) MUST remain.

## Command contract

From repository root:

```bash
go test ./...
go vet ./...
```

Both exit 0 on clean checkout after `go mod download`.

Optional (documented in quickstart):

```bash
goose -dir migrations postgres "$DSN" up
sqlc generate
```

## Future expansion (not F101)

```text
cmd/worker/main.go
internal/domain/lore/
internal/application/commands/
internal/adapters/http/
internal/adapters/mcp/
graph-service/
```
