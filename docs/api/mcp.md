# MCP API

MemLore exposes a **domain** MCP surface for coding agents. Graphiti and Neo4j
internals are not part of the product contract.

**Transport (this slice)**: local **stdio** via `memlore mcp` (Go binary; build with `scripts/install-memlore.sh`).
Streamable HTTP MCP is out of scope.

Mutating tools require an explicit non-empty `actor_id` argument **in local
mode** (OIDC unset). When OIDC is configured, pass `access_token` (or set
`MEMLORE_ACCESS_TOKEN`); `actor_id` is ignored for identity. Roles:
`reader` / `writer` / `admin` (see `specs/015-oidc-rbac/`).

Actor is not inferred from the environment except via `MEMLORE_ACCESS_TOKEN`
when OIDC is on.

| Tool | Purpose | Status |
|------|---------|--------|
| `memlore.remember` | Store human-authored lore with provenance | implemented |
| `memlore.get` | Fetch by id | implemented |
| `memlore.verify` | Verify (self-verify allowed; idempotent re-verify) | implemented |
| `memlore.explain` | Lore fields, chronological audits, and explainable authority evaluation (no NL summary) | implemented |
| `memlore.search` | Exact scope list (`kind`+`key`); current only unless `include_stale` | implemented |
| `memlore.knowledge_search` | Dual-plane knowledge search (query-relevant governance + graph; optional `include_stale`; optional `graph_receipt`) | implemented |
| `memlore.get_for_task` | Compiled named context packet for a task (`sections`, `sources`; never packs stale; additive files/ticket/agent_id) | implemented |
| `memlore.repo_profile` | Compact repository intelligence profile (named sections; omit empty) | implemented |
| `memlore.supersede` | Replace while preserving history | implemented |
| `memlore.invalidate` | Mark invalid without deleting evidence | implemented |

`memlore.remember` always stores origin `human_authored` (parity with REST
create).

Default retrieval (`search`, `knowledge_search`, `get_for_task`) omits
superseded and invalidated lore. `get` and `explain` still return history.
`get_for_task` surfaces structural conflict groups when current statements in
the same scope disagree. Compiled items include `trust_band` and explainable
`authority_factors`. Named `sections` (`architecture`, `decisions`,
`conventions`, `task_context`, `gotchas`) are omitted when empty. Optional
`branch`, `ticket`, `changed_files`, `working_files`, and `agent_id` do not
fail the request when omitted; `agent_id` is never an authority factor.
`memlore.explain` adds `trust_band`, `authority_score`,
`authority_factors`, and `factor_breakdown`.

## Local attach

```bash
./scripts/install-memlore.sh
./bin/memlore mcp
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
payload schemas. Knowledge search:
`specs/019-semantic-graph-retrieval/contracts/knowledge-search-v2.md`.
Context compile: `specs/012-context-compiler/contracts/context-compile.md`.
F021 packet: `specs/021-agent-context-bootstrap/contracts/context-packet.md`.
Repository profile: `specs/020-repo-intelligence-profile/contracts/repository-profile.md`.
Invalidate and supersede:
`specs/013-supersede-invalidate/contracts/lifecycle-lore.md`.
Temporal filter + conflicts:
`specs/014-conflict-filtering/contracts/`.
Auth + RBAC: `specs/015-oidc-rbac/contracts/auth-rbac.md`.
Authority evaluation: `specs/016-authority-factors/contracts/authority-evaluation.md`.
