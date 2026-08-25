# Feature Specification: MCP Lore Tools

**Feature Branch**: `002-mcp-lore-tools`  
**Created**: 2026-08-25  
**Status**: Draft  
**Input**: User description: "MCP lore tools: remember, get, verify, explain (and optional search) for coding agents over existing lore services"

## Clarifications

### Session 2026-08-25

- Q: How is actor identity supplied on MCP mutating tools? → A: Required `actor_id` tool argument on `remember` and `verify`.
- Q: Is `memlore.search` in this feature’s MVP? → A: Yes — include `memlore.search` (exact scope list) in this feature MVP.
- Q: What does `memlore.explain` return? → A: Structured lore entry fields plus chronological audit list (no generated narrative).
- Q: How do coding agents run/connect to MemLore MCP locally? → A: Local stdio MCP server via `memlore mcp` CLI.
- Q: What origin does `memlore.remember` store? → A: Always `human_authored` (parity with REST create); agent origins deferred.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Remember lore via MCP (Priority: P1)

A coding agent stores a human-authored engineering knowledge statement for a
scope through MemLore’s MCP tool surface, with actor identity and optional
evidence, and receives a stable lore entry id with provenance.

**Why this priority**: Agents need a first-class write path without calling REST
directly; this is the core of MCP parity with the existing create capability.

**Independent Test**: Call `memlore.remember` with statement, scope, actor, and
optional evidence; confirm a lore entry is created and returned with id and
provenance matching existing governance rules.

**Acceptance Scenarios**:

1. **Given** a valid statement, scope (`kind`+`key`), and required `actor_id`,
   **When** the agent calls `memlore.remember`, **Then** MemLore stores a
   human-authored lore entry and returns id, statement, scope, origin,
   verification status, and evidence.
2. **Given** invalid input (missing statement, scope, or `actor_id`), **When**
   `memlore.remember` is called, **Then** the tool returns a clear validation
   failure and stores nothing.
3. **Given** an existing identical statement in the same scope, **When**
   `memlore.remember` is called again, **Then** a new distinct entry id is
   created (duplicates allowed, same as REST).

---

### User Story 2 - Get and explain lore via MCP (Priority: P1)

A coding agent fetches a lore entry by id and can request an explanation of
provenance (who authored, verification state, evidence, and audit trail).

**Why this priority**: Agents must retrieve and justify knowledge before acting
on it; explainability is a MemLore differentiator.

**Independent Test**: Create an entry, call `memlore.get` and `memlore.explain`;
confirm content/provenance and audit actions are returned; unknown id fails
clearly.

**Acceptance Scenarios**:

1. **Given** an existing lore entry id, **When** `memlore.get` is called,
   **Then** the full lore entry (statement, scope, origin, verification,
   evidence, timestamps) is returned.
2. **Given** an existing lore entry id, **When** `memlore.explain` is called,
   **Then** the response includes the lore entry fields plus a chronological
   list of audit actions for that entry (no generated natural-language
   narrative required).
3. **Given** an unknown id, **When** `memlore.get` or `memlore.explain` is
   called, **Then** the tool reports not found without leaking internals.

---

### User Story 3 - Verify lore via MCP (Priority: P1)

A human or agent-assisted workflow marks a lore entry verified through MCP,
preserving authorship origin and supporting self-verify and idempotent
re-verify as already defined for REST.

**Why this priority**: Verification is required for authority distinction;
agents and operators need MCP access to the same behavior.

**Independent Test**: Verify an unverified entry via `memlore.verify`; confirm
status/verifier/time; re-verify does not create a second verify audit.

**Acceptance Scenarios**:

1. **Given** an unverified entry and a required non-empty `actor_id`, **When**
   `memlore.verify` is called, **Then** the entry becomes verified with verifier
   and time recorded; origin remains `human_authored`.
2. **Given** an already verified entry, **When** `memlore.verify` is called
   again, **Then** the entry remains verified with original verification
   metadata (idempotent no-op).
3. **Given** missing `actor_id` or unknown id, **When** `memlore.verify` is
   called, **Then** the tool returns validation or not-found failure
   appropriately.

---

### User Story 4 - Search/list lore via MCP (Priority: P1)

A coding agent lists lore entries for an exact scope through MCP to discover
relevant knowledge before coding.

**Why this priority**: Included in this feature’s MVP so agents can discover
scoped lore without REST.

**Independent Test**: Create entries in two scopes; `memlore.search` for one
scope returns only matching entries.

**Acceptance Scenarios**:

1. **Given** entries in scopes A and B, **When** `memlore.search` is called for
   scope A (`kind`+`key`), **Then** only scope A entries are returned.
2. **Given** a scope with no entries, **When** `memlore.search` is called,
   **Then** an empty list is returned (not an error).
3. **Given** missing scope kind or key, **When** `memlore.search` is called,
   **Then** the tool returns a validation failure.

---

### User Story 5 - Run local MCP server (Priority: P1)

A developer starts MemLore’s MCP server locally so a coding agent can attach via
stdio and use the lore tools.

**Why this priority**: Without a runnable local MCP entrypoint, tools cannot be
exercised by agents even if implemented.

**Independent Test**: Start `memlore mcp`, confirm the server advertises the
MemLore lore tools and responds to a remember/get round-trip from a test client.

**Acceptance Scenarios**:

1. **Given** a working local MemLore governance database, **When** a developer
   runs the MemLore MCP CLI entrypoint, **Then** an MCP server starts on stdio
   and lists the lore tools for this feature.
2. **Given** the MCP server is running, **When** a client calls
   `memlore.remember` then `memlore.get`, **Then** both succeed using the same
   governance rules as REST.

---

### Edge Cases

- MCP tools MUST NOT expose Graphiti or Neo4j internals.
- Actor identity is supplied via required `actor_id` on mutating tools
  (`remember`, `verify`); omitting it is a validation error.
- Tool errors MUST be actionable: `validation_error` for bad input,
  `not_found` for unknown ids. There is no third `unavailable` tool code.
  Infrastructure failures (e.g. database unreachable) MUST surface as a
  generic tool failure without leaking internals (stack traces, SQL,
  Graphiti/Neo4j).
- `get_for_task`, supersede, invalidate, and context compilation are out of
  scope for this feature; `memlore.search` (exact scope list) is in scope.
- Agent-authored create origins remain out of scope; `memlore.remember` always
  creates `human_authored` entries (same as REST create).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose MCP tool `memlore.remember` that creates a
  scoped human-authored lore entry using the same governance rules as REST
  create (scope kind+key, optional evidence type+value, duplicates allowed).
- **FR-002**: System MUST expose MCP tool `memlore.get` that returns a lore
  entry by id.
- **FR-003**: System MUST expose MCP tool `memlore.verify` that verifies a lore
  entry with the same semantics as REST verify (self-verify allowed; idempotent
  re-verify).
- **FR-004**: System MUST expose MCP tool `memlore.explain` that returns the
  lore entry’s fields together with its chronological audit records for the
  given id. A generated natural-language summary is out of scope.
- **FR-005**: System MUST require a non-empty `actor_id` tool argument on
  mutating MCP tools (`memlore.remember`, `memlore.verify`). Actor MUST NOT be
  inferred from environment defaults or omitted session metadata alone.
- **FR-006**: System MUST return clear not-found outcomes for unknown ids on
  get/explain/verify.
- **FR-007**: System MUST NOT expose raw Graphiti/Neo4j operations through these
  tools.
- **FR-008**: System MUST expose MCP tool `memlore.search` that lists lore
  entries by exact scope `kind`+`key` (empty list when none; validation error
  when scope is incomplete). Semantic/vector search is out of scope.
- **FR-009**: MCP tool behavior MUST remain consistent with existing REST lore
  contracts for overlapping operations (no divergent business rules).
- **FR-010**: Mutating MCP operations MUST continue to write audit records as
  already required by the governance plane.
- **FR-011**: System MUST provide a local CLI entrypoint that runs the MemLore
  MCP server over stdio for coding-agent attachment.
- **FR-012**: Streamable HTTP MCP transport is out of scope for this feature.

### Key Entities

- Reuses existing **Lore Entry**, **Scope**, **Evidence Reference**, **Actor**,
  **Audit Record** from the governance plane.
- **MCP Tool Invocation**: An agent-initiated call with tool name, arguments,
  and structured result or error.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A coding agent can remember, get, verify, explain, and search lore
  end-to-end via MCP in under 2 minutes in local development.
- **SC-002**: 100% of overlapping create/get/verify/list acceptance rules from
  the REST lore slice remain true when exercised through MCP contract tests.
- **SC-003**: 100% of unknown-id MCP get/explain/verify calls return not-found
  (not empty success payloads).
- **SC-004**: 0 Graphiti/Neo4j internal APIs are reachable via the MemLore MCP
  tool list for this feature.
- **SC-006**: A developer can start the local stdio MCP server with a single
  documented CLI command and complete a remember→get tool round-trip in local
  development.

## Assumptions

- This feature is an MCP adapter over existing application services; no new
  lore domain rules beyond transport/actor wiring.
- Transport for this feature is local **stdio** MCP via a `memlore mcp` (or
  equivalent) CLI command; Streamable HTTP MCP is deferred.
- Soft-auth actor identity is supplied as a required `actor_id` tool argument on
  mutating calls; OIDC remains out of scope.
- Agent-authored or agent-inferred creation remains out of scope;
  `memlore.remember` always persists origin `human_authored`.
- `memlore.get_for_task`, supersede, invalidate, conflict detection, and
  Graphiti sync remain future features.
- `memlore.search` is in MVP and maps to exact scope list for this slice, not
  semantic retrieval.
- REST continues to work unchanged; MCP is additive.
