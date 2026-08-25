# MCP API

MemLore exposes a **domain** MCP surface for coding agents. Graphiti and Neo4j
internals are not part of the product contract.

**Transport (this slice)**: local **stdio** via `uv run memlore mcp`.
Streamable HTTP MCP is out of scope.

Mutating tools require an explicit non-empty `actor_id` argument. Actor is not
inferred from the environment.

| Tool | Purpose | Status |
|------|---------|--------|
| `memlore.remember` | Store human-authored lore with provenance | implemented |
| `memlore.get` | Fetch by id | implemented |
| `memlore.verify` | Verify (self-verify allowed; idempotent re-verify) | implemented |
| `memlore.explain` | Lore fields plus chronological audits (no NL summary) | implemented |
| `memlore.search` | Exact scope list (`kind`+`key`) | implemented |
| `memlore.get_for_task` | Compiled context packet | deferred |
| `memlore.supersede` | Replace while preserving history | deferred |
| `memlore.invalidate` | Mark invalid without deleting evidence | deferred |

`memlore.remember` always stores origin `human_authored` (parity with REST
create).

## Local attach

```bash
uv run memlore mcp
```

Example agent config:

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

Errors use `{code}: {message}` with `validation_error` or `not_found`.
Infrastructure failures are generic tool failures without leaking internals.

See `specs/002-mcp-lore-tools/contracts/mcp-lore-tools.md` for argument and
payload schemas.
