# MemLore

Shared engineering memory for humans and AI coding agents.

MemLore preserves the *why* behind the code: architectural decisions, engineering
context, provenance, authority, temporal history, conflicts, superseded
knowledge, and agent observations — so coding agents retrieve trustworthy
context with evidence, not just similar text.

## Why MemLore?

Generic AI memory stores embeddings and snippets. Engineering work needs more:

- **Provenance** — who said it, human or agent, with evidence
- **Authority** — verified architecture outranks unverified inference
- **Temporal truth** — supersede facts without erasing history
- **Conflicts** — surface drift between policy and implementation
- **Scopes** — organization → team → project → repository → task

## Quick start

Requirements: Go 1.25+, Docker / Docker Compose. Python 3.12 + [uv](https://docs.astral.sh/uv/)
only if you run `graph-service/`.

```bash
cp .env.example .env
docker compose up -d postgres
./scripts/install-memlore.sh
./bin/memlore migrate
./bin/memlore serve
```

Health check: `GET http://127.0.0.1:8080/health`

## Connect a coding agent

Start the stdio MCP server (Postgres must be up and migrated):

```bash
./bin/memlore mcp
```

Tools: `memlore.remember`, `get`, `verify`, `explain`, `search`,
`knowledge_search`, `get_for_task`, `supersede`, `invalidate`.
Mutating tools require `actor_id` in local mode (OIDC optional).
See [docs/api/mcp.md](docs/api/mcp.md).

## CLI

```bash
./bin/memlore serve     # REST on :8080
./bin/memlore mcp       # stdio MCP
./bin/memlore migrate   # goose (embedded)
./bin/memlore worker    # outbox → graph-service
```

Or `go run ./cmd/memlore <command>` from this repo.

## Documentation

| Area | Link |
|------|------|
| Architecture overview | [docs/architecture/overview.md](docs/architecture/overview.md) |
| ADRs | [docs/adr/](docs/adr/) |
| Development setup | [docs/development/setup.md](docs/development/setup.md) |
| TDD | [docs/development/tdd.md](docs/development/tdd.md) |
| Constitution | [.specify/memory/constitution.md](.specify/memory/constitution.md) |

## Status

MemLore Core is **Go** (REST, MCP, migrations, worker, authority, lifecycle,
OIDC/RBAC). Knowledge plane is a thin Python **graph-service** (Graphiti/Neo4j).

## License

Apache-2.0 (planned; confirm before public release).
