# Development setup

## Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/)
- Go 1.25+ (MemLore Core skeleton)
- Docker / Docker Compose

## Bootstrap

```bash
cp .env.example .env
docker compose up -d postgres
uv sync
./scripts/install-memlore.sh   # builds bin/memlore for MCP cross-project use
./bin/memlore migrate          # canonical schema apply (embedded goose)
uv run alembic upgrade head    # legacy Python path; same DDL
uv run pytest
./bin/memlore serve            # Go REST on :8080
./bin/memlore mcp              # Go MCP stdio
```

Postgres is published on host port **15432** by default (see `docker-compose.yml`
and `.env.example`) to avoid clashes with local Postgres installs.

### Cross-project MCP (e.g. another repo in Cursor)

Build once from this repo:

```bash
./scripts/install-memlore.sh
```

In the **consumer** project's `.cursor/mcp.json` use an **absolute binary path**
(Cursor ignores `cwd` for `go run`):

```json
{
  "mcpServers": {
    "memlore": {
      "type": "stdio",
      "command": "/absolute/path/to/memlore/bin/memlore",
      "args": ["mcp"],
      "env": {
        "MEMLORE_POSTGRES_DSN": "postgresql://memlore:memlore@localhost:15432/memlore"
      }
    }
  }
}
```

Run `./bin/memlore migrate` before first use if tables are missing.

Legacy Python adapters (`uv run memlore serve` / `uv run memlore mcp`) still work
but print deprecation notices; prefer Go binaries.

Integration tests need Postgres:

```bash
docker compose up -d postgres
uv run pytest -m integration
# or skip them when DB is down (auto-skip via fixture)
uv run pytest
```

Optional override: `MEMLORE_TEST_DATABASE_URL=postgresql+psycopg://...`

## Quality commands

### Python

```bash
uv run ruff check src tests
uv run ruff format src tests
uv run mypy
uv run pytest
```

### Go (MemLore Core)

```bash
go test ./...
go vet ./...
go run ./cmd/memlore version
```

Optional Postgres migration check (use a fresh database; see
`specs/003-go-core-skeleton/quickstart.md`):

```bash
go test -tags=integration ./migrations/...
go test -tags=integration ./internal/infrastructure/postgres/...
```

## Spec Kit

This repo uses GitHub Spec Kit with the Cursor integration.

Typical feature flow:

1. `/speckit-constitution` (already ratified v1.0.0)
2. `/speckit-specify`
3. `/speckit-clarify` (if needed)
4. `/speckit-checklist`
5. `/speckit-plan`
6. `/speckit-tasks`
7. `/speckit-analyze`
8. `/speckit-implement`

Constitution: `.specify/memory/constitution.md`
