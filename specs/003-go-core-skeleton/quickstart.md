# Quickstart: Go Core Skeleton (F101)

**Prerequisites**: Go 1.25+, existing Python/uv setup unchanged

## Verify Go module

```bash
go mod download
go test ./...
go vet ./...
```

Expected: all packages pass; no database required for default unit tests.

## Run stub binary (after implementation)

```bash
go run ./cmd/memlore
# or
go build -o bin/memlore ./cmd/memlore && ./bin/memlore version
```

## Apply goose migration (optional local check)

```bash
docker compose up -d postgres
export DSN="postgresql://memlore:memlore@localhost:15432/memlore?sslmode=disable"
goose -dir migrations postgres "$DSN" up
goose -dir migrations postgres "$DSN" status
```

Confirm tables: `lore_entries`, `audit_records`.

**Note**: If your dev DB was created by Alembic, use a **separate database**
for goose-first testing to avoid double-migration conflicts.

## sqlc (after implementation)

```bash
sqlc generate
go test ./...
```

## Python unchanged

```bash
uv sync
uv run pytest
uv run memlore serve   # still Python
uv run memlore mcp     # still Python
```

F101 does not switch the default runtime to Go.

## CI parity

Same checks as GitHub Actions:

```bash
go test ./...
go vet ./...
uv run ruff check src tests
uv run ruff format --check src tests
uv run mypy
uv run pytest
```
