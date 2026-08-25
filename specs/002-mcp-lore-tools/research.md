# Research: MCP Lore Tools

**Feature**: `002-mcp-lore-tools`  
**Date**: 2026-08-25

## R1 — MCP SDK and server API

**Decision**: Add the official MCP Python SDK v2 as a runtime dependency
(`mcp>=2.1,<3`). Build the server with `MCPServer` (`from mcp.server import
MCPServer`), not the v1 `FastMCP` import path and not the standalone
`fastmcp` package.

**Rationale**: ADR 0002 already names the official MCP Python SDK. v2 is the
current stable line (2026-07-28 protocol), ships stdio plus an in-memory
`Client` for contract tests, and maps tool failures via `ToolError` so agents
see recoverable errors. `mcp` 1.x (`FastMCP` at `mcp.server.fastmcp`) is
legacy; a bare `pip install mcp` now resolves 2.x.

**Alternatives considered**:
- `mcp>=1.28,<2` — still works, but fights the default install and uses a
  renamed API we would migrate off immediately.
- Standalone `fastmcp` (jlowin) — extra product surface; not the stack ADR.
- Hand-rolled JSON-RPC over stdio — more code, no protocol/test helpers.

## R2 — Transport and CLI entrypoint

**Decision**: Local **stdio** only. Add CLI subcommand `memlore mcp` that
constructs the MCP server and runs `server.run(transport="stdio")`. Do not
expose Streamable HTTP, SSE, or a `--http` flag in this feature.

**Rationale**: Spec clarification and FR-011/FR-012. Coding agents attach by
spawning a process (Cursor/Claude `mcpServers.command`). Stdio uses stdout as
the protocol stream, so application logs MUST go to **stderr**.

**Alternatives considered**:
- Streamable HTTP MCP — deferred by spec.
- Separate `memlore-mcp` console script — unnecessary; existing `memlore`
  CLI already owns local developer commands.

## R3 — Tool names

**Decision**: Register tools with dotted names exactly as specified:
`memlore.remember`, `memlore.get`, `memlore.verify`, `memlore.explain`,
`memlore.search`. Do not also register Graphiti, Neo4j, or ADR-deferred tools
(`get_for_task`, `supersede`, `invalidate`).

**Rationale**: Spec FR-001–FR-008 and ADR 0003/0004. SEP-986 allows `.` in
MCP tool names (1–64 chars). Explicit `name=` on `@mcp.tool` avoids the SDK
defaulting to the Python function name.

**Alternatives considered**:
- Snake_case only (`remember`) — weaker product namespace; contradicts ADR.
- Expose full ADR 0003 catalog now — out of scope.

## R4 — Actor identity on mutating tools

**Decision**: Require a non-empty `actor_id: str` **tool argument** on
`memlore.remember` and `memlore.verify`. Do not read actor from environment,
CLI flags, or session metadata. Read tools (`get`, `explain`, `search`) do
not take `actor_id`.

**Rationale**: Spec clarification and FR-005. Maps 1:1 onto existing
`CreateLoreCommand.actor_id` / `VerifyLoreCommand.actor_id` (REST uses
`X-Memlore-Actor` instead). Empty/whitespace `actor_id` is a validation
error after strip, same as application services.

**Alternatives considered**:
- Default actor from `MEMLORE_ACTOR` — rejected in clarify.
- MCP session identity / OIDC — out of scope.

## R5 — Adapter over existing application services

**Decision**: MCP is a new adapter (`src/memlore/adapters/mcp/`) that calls
the existing handlers via `AppContainer` + one Unit of Work per tool call.
No new domain rules, no Alembic migrations, no Graphiti.

| Tool | Handler(s) |
|------|------------|
| `memlore.remember` | `CreateLoreHandler` (`origin` always `human_authored`) |
| `memlore.get` | `GetLoreHandler` |
| `memlore.verify` | `VerifyLoreHandler` |
| `memlore.search` | `ListLoreByScopeHandler` |
| `memlore.explain` | `GetLoreHandler` + `ListAuditsHandler` in the **same** UoW |

**Rationale**: Spec assumption: adapter-only. Explain is a view over existing
get + chronological audits, not a new domain entity. Composing two queries in
the adapter avoids a speculative application service.

**Alternatives considered**:
- New `ExplainLoreHandler` — slightly cleaner, not required.
- Reimplement create/verify in the MCP layer — would diverge from REST
  (violates FR-009).

## R6 — Payload parity with REST

**Decision**: Success payloads use the same JSON field names and enums as
`LoreEntryResponse` / `AuditRecordResponse` in
`specs/001-scoped-lore-entry/contracts/rest-lore-entries.md`. Search returns
`{"items": [...]}`. Explain returns all lore-entry fields **plus**
`audits: [AuditRecord, ...]` (chronological; no generated narrative).

Reuse the existing Pydantic response models in `adapters/rest/schemas.py`
from the MCP adapter (they have no FastAPI dependency). Do not invent a
parallel DTO package unless a third adapter appears.

**Rationale**: FR-009 (no divergent business rules / overlapping contracts).
Importing REST schemas is a small layering smell vs extracting
`adapters/common/`; extraction is deferred to keep this slice small.

**Alternatives considered**:
- Nested `{ "entry": {...}, "audits": [...] }` for explain — clearer
  nesting, but spec wording is “entry fields plus audit list”.
- Separate MCP-only field names — would break SC-002 parity.

## R7 — Errors (agent-visible, no internals)

**Decision**: Map `ValidationError` and `NotFoundError` to SDK `ToolError`
(`from mcp.server.mcpserver.exceptions import ToolError`) so the call returns
`is_error=true` with an actionable message. Message format:

```text
{code}: {message}
```

where `code` is `validation_error` or `not_found` (same codes as REST). Do
**not** raise `MCPError` / JSON-RPC errors for missing ids or bad input —
those hide the message from the model. Unexpected exceptions stay uncaught so
the SDK treats them as crashes (generic is_error, traceback on stderr only).
There is no `unavailable` tool error code.

Missing required tool args (including omitted `actor_id`) are rejected by the
MCP input schema before the handler runs; contract tests still assert a
failed/invalid outcome, not a stored row.

**Rationale**: Spec edge cases distinguish validation vs not-found vs
infrastructure crash; MCP SDK guidance: `ToolError` for recoveries the agent
can fix, uncaught exceptions for unexpected unavailability.

**Alternatives considered**:
- JSON-RPC `-32602` for not-found — host sees it, the agent often does not.
- Return error JSON with `is_error=false` — looks like success.

## R8 — Testing strategy

**Decision**:
- **Unit**: MCP error mapping / argument wiring against `InMemoryUnitOfWork`
  (same fakes as REST).
- **Contract** (primary): in-memory `mcp.Client(server)` against
  `create_mcp_server(memory_container)` covering list-tools, remember / get /
  verify / explain / search, duplicates, empty search, missing actor, unknown
  id. This is the MCP analogue of REST `TestClient` tests (SC-002).
- **e2e (sparse)**: one stdio round-trip that spawns `memlore mcp` (skip if
  Postgres is down), remember → get → verify → explain → search — US5 /
  SC-001 path / SC-006. Wall-clock “under 2 minutes” is a local quickstart
  check, not a CI timer.
- No new Postgres repository tests; persistence is unchanged.

**Rationale**: Constitution test pyramid; MCP SDK documents `Client(server)`
as the in-process harness. Stdio subprocess is the only way to prove the CLI
entrypoint.

**Alternatives considered**:
- Only stdio subprocess tests — slower, harder to isolate.
- SQLite MCP tests — rejected in 001 for JSONB/enum drift; still rejected.

## R9 — New dependency justification

**Decision**: Runtime: `mcp>=2.1,<3` (MIT). No `mcp[cli]` extra (we own
`memlore mcp`). Dev stack already has `pytest-asyncio` for async contract
tests.

**Rationale**: Constitution dependency policy; required to speak MCP. License
and purpose are clear.

## R10 — Observability and stdio hygiene

**Decision**: Structured `log_operation` on each tool (`operation` values
`mcp.remember`, `mcp.get`, `mcp.verify`, `mcp.explain`, `mcp.search`) with
`actor_id` when present and `lore_entry_id` when known. Configure the MCP
CLI process so logging handlers write to **stderr** only.

**Rationale**: Constitution VIII; stdout is the MCP byte stream.

## R11 — Documentation (same unit of work)

**Decision**: Update `docs/api/mcp.md` from “planned” to implemented subset;
add CLI + agent-config snippet to `README.md` and `docs/development/setup.md`.
No new ADR: 0003 already decided domain MCP over Graphiti; this feature
implements the in-scope subset over stdio.

**Rationale**: Constitution IV. A new transport ADR would repeat clarify
outcomes already in the spec.
