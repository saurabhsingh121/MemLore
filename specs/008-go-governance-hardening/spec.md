# Feature Specification: Go Governance Hardening (F106a)

**Branch**: `008-go-governance-hardening`  
**Depends on**: F104 (REST), F105 (MCP)

## Goal

Make Go the production path for governance-plane MemLore: reliable migrations,
CI integration coverage, installable binary for cross-project MCP, and Python
adapter deprecation notices.

## Acceptance Criteria

- `memlore migrate` applies embedded goose migrations (works from any cwd)
- Alembic integration test proves `upgrade head` creates tables on empty DB
- CI runs Go integration tests against Postgres service
- Python `memlore serve` / `memlore mcp` print deprecation guidance to stderr
- `scripts/install-memlore.sh` builds `bin/memlore` for MCP cross-project use
- Setup docs cover MCP config with absolute binary path

## Out of Scope

- Removing Python adapters (deprecation only)
- Graph-service / knowledge plane (F106)
- OIDC auth (F010)
