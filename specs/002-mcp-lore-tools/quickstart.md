# Quickstart: MCP Lore Tools (local)

## Prerequisites

- Docker Compose (Postgres on host port **15432**)
- `uv` + Python 3.12
- Repo on branch `002-mcp-lore-tools`
- Feature 001 migrations applied (`uv run alembic upgrade head`)

## Steps (once implemented)

```bash
cp .env.example .env
docker compose up -d postgres
uv sync
uv run alembic upgrade head
```

### Start the stdio MCP server

```bash
uv run memlore mcp
```

This process speaks MCP on stdin/stdout. Attach a coding agent with:

```json
{
  "mcpServers": {
    "memlore": {
      "command": "uv",
      "args": ["run", "memlore", "mcp"],
      "cwd": "<repo-root>"
    }
  }
}
```

REST remains available in another terminal if needed:

```bash
uv run memlore serve
```

### Five-tool path (SC-001)

From an MCP client (agent, or contract harness), complete this sequence in
local development in **under 2 minutes**:

1. `memlore.remember` with `statement`, `scope` `{kind, key}`, required
   `actor_id`, optional `evidence`.
2. `memlore.get` with the returned `id`.
3. `memlore.verify` with the same `id` and `actor_id`.
4. `memlore.explain` with the `id` (fields + chronological `audits`; no NL
   summary).
5. `memlore.search` for that scope (`items` includes the entry).

Do **not** omit `actor_id` on remember/verify and do **not** rely on env
defaults.

## Verify acceptance locally

```bash
uv run pytest tests/unit tests/contract tests/integration tests/e2e -q
```

Expect behaviors in `contracts/mcp-lore-tools.md`. The stdio e2e test skips
if Postgres is not reachable (`docker compose up -d postgres`).
