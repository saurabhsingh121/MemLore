# Feature Specification: Go MCP Lore Tools (F105)

**Branch**: `007-go-mcp-lore-tools`  
**Depends on**: F104 (REST/application layer)

## Goal

Expose the five lore MCP tools via Go stdio server `memlore mcp`, matching
`specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md`.

## Acceptance Criteria

- Tools: `memlore.remember`, `get`, `verify`, `explain`, `search`
- Error envelope: `validation_error: {message}`, `not_found: {message}` as tool errors
- Go contract tests mirror Python `tests/contract/test_mcp_*.py`
- `go run ./cmd/memlore mcp` (stdio; logs on stderr)
- Python `uv run memlore mcp` remains available

## Out of Scope

- Streamable HTTP MCP
- Graphiti/Neo4j tools
- Disabling Python MCP
