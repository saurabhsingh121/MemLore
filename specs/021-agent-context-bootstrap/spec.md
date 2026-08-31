# Feature Specification: Agent Context Bootstrap / richer get_for_task (F021)

**Feature Branch**: `021-agent-context-bootstrap`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F021  
**Depends on**: F007 (context compiler v1 DONE), F020 (repository intelligence profile DONE)  
**Extends**: F007  
**Input**: Make compiled task context the primary coding-agent entry point: richer optional inputs and a named engineering packet, without breaking existing compile clients.

## Goal

A coding agent starting a task should receive a **compiled context packet** with named engineering sections — relevant architecture, applicable decisions, conventions, task-specific knowledge, gotchas, conflicts, and sources — rather than a flat bag of similar text. The agent starts informed instead of reconstructing repository history. MemLore must not invent statements, must not silently promote candidates to canonical, and must not let popularity or agent identity override authority.

This feature **extends** compile v1. Existing callers that send only task, scope, optional query, and optional token budget MUST still succeed. Historical IDs F007/F109 stay closed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Named task briefing instead of a flat dump (Priority: P1)

A coding agent (or developer) asks for context for a concrete task in a repository they are authorized to read. MemLore returns the familiar compiled packet **and** named engineering sections when supporting current knowledge exists. Empty sections are omitted. Nothing is invented. Verified architecture still outranks an unverified inference. Conflicts among current statements are listed, not dropped. Superseded knowledge stays out of the briefing.

**Why this priority**: This is the product value of F021 — reduce context-discovery cost and tokens by giving agents a structured packet they can use immediately.

**Independent Test**: Store current lore covering architecture, a decision, a convention, a gotcha, and a task-specific statement; compile for a matching task; verify named sections appear, empty section types are absent, the flat item list still contains the budgeted statements, and a superseded entry does not appear.

**Acceptance Scenarios**:

1. **Given** current lore in a repository that includes an architecture statement, an ADR-backed decision, a coding convention, a known gotcha, and a statement about the payment outbox, **When** context is compiled for the task "implement payment outbox", **Then** the packet includes named sections for architecture, decisions, conventions, gotchas, and task context as applicable, each with those statements and evidence when present.
2. **Given** no current knowledge that maps to coding conventions, **When** context is compiled, **Then** the packet does not include a conventions section (omitted, not padded).
3. **Given** a verified human-authored architecture statement and an unverified agent inference in the same packet, **When** both are included, **Then** the verified architecture is ordered ahead of the unverified inference.
4. **Given** two current disagreeing statements in the retrieval set, **When** context is compiled, **Then** both sides appear in the conflicts list and neither side is silently dropped from that list.
5. **Given** a superseded lore entry and a current successor, **When** context is compiled, **Then** the superseded statement does not appear in items or named sections.

---

### User Story 2 - Richer optional inputs without breaking old callers (Priority: P1)

Callers may supply branch, ticket, changed files, working files, query, token budget, and agent identity in addition to the existing required task and scope. Unspecified optional fields MUST NOT fail the request. File paths and ticket text influence which leftover items land in Task Context (and the retrieval query). Agent identity is recorded for later feedback only — it MUST NOT change ranking, trust, or what is treated as canonical. A caller that still sends only task + scope (+ optional query/budget) MUST receive a successful packet with the original fields.

**Why this priority**: Agents already call compile/`get_for_task` with the v1 body; this slice must enrich that path, not fork a new one.

**Independent Test**: Compile with the v1 body and assert success plus original fields; compile with files/ticket and assert Task Context prefers matching statements; compile with agent identity and assert scores match an identical request without identity.

**Acceptance Scenarios**:

1. **Given** a v1 compile request with only task, scope, query, and token budget, **When** it is submitted, **Then** the response is successful and still includes `task`, `query`, `scope`, `items`, `meta`, `warnings`, and `conflicts`.
2. **Given** optional branch, ticket, files, and agent identity omitted, **When** a valid task+scope request is submitted, **Then** the request is not rejected for missing those fields.
3. **Given** lore about `src/payments/outbox.go` and unrelated lore, **When** compile includes that path as a changed or working file, **Then** Task Context prefers the file-relevant statement over unrelated leftovers.
4. **Given** two otherwise identical compile requests that differ only by agent identity, **When** both complete, **Then** item membership, order, and authority scores match.
5. **Given** a caller-supplied token budget, **When** more current knowledge exists than fits, **Then** the packet stays within that budget (same estimation rules as compile v1).

---

### User Story 3 - Same packet on REST, MCP, and CLI (Priority: P2)

Authorized callers get the same packet through REST compile, MCP `get_for_task`, and a developer CLI briefing command. JSON payloads match across REST and MCP. CLI prints a compact human-readable briefing of populated sections. Membership and read authorization match other lore read operations. No new MCP tool is added.

**Why this priority**: Constitution: CLI + REST for humans; MCP where agents participate. Parity prevents split-brain. Agents already know `get_for_task`.

**Independent Test**: REST and MCP contract tests with the same inputs; CLI unit test of formatted output from a fixture packet.

**Acceptance Scenarios**:

1. **Given** the same task, scope, files, and token budget, **When** REST and MCP are called, **Then** section ids, item ids, statements, conflicts, and meta counts match.
2. **Given** a caller without membership on the requested scope (OIDC membership mode), **When** they compile, **Then** access is denied the same way as other reads for that scope.
3. **Given** a compiled packet with architecture and conflicts, **When** CLI formats it, **Then** the printed briefing includes the task, those section headings, and conflict statements.
4. **Given** the MCP tool list, **When** tools are enumerated, **Then** `get_for_task` is still the compile tool and the tool count is unchanged (no new tool).

---

### Edge Cases

- Optional fields that are empty strings or empty lists are treated as unspecified (not a validation error).
- `branch` is accepted and echoed when provided; v1 does not filter knowledge by git branch (no branch-scoped lore yet).
- Observed implementation drift is not invented: F050 is not built; do not emit a drift section from code analysis. Existing conflict groups remain the honesty surface for disagreement.
- Potentially stale knowledge is not resurrected into the briefing. Default retrieval stays current-only. No `include_stale` is added in this slice.
- Statements that match no briefing or task-context rule are omitted from named sections and counted as unclassified; they MAY still appear in the flat `items` list when they survived ranking and budget (backward-compatible compile list). They are not dumped into a generic "other" section.
- Token budget omitted uses default 4096. Conflicts remain listed even if a side is omitted from `items` due to budget.
- Graph-service unavailability yields the existing retrieval warning and a governance-only packet when governance hits exist.
- Duplicate equivalent statements across governance and graph are deduplicated as in compile v1 (governance preferred).
- Scope kinds other than repository remain valid for compile (v1 contract). Repository-wide briefing merge applies when the requested scope is a repository.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Callers MUST continue to provide a non-empty task and a scope `{kind, key}`. Query remains optional and MUST default to task when omitted. Token budget remains optional and MUST default to 4096 estimated tokens.
- **FR-002**: The system MUST accept these additional optional inputs without requiring them: branch, ticket, changed files, working files, agent identity. Unspecified optional fields MUST NOT fail the request.
- **FR-003**: The system MUST compile the packet on read from existing current knowledge (governance + graph). It MUST NOT persist compiled packets or invent statements not present in retrieved knowledge.
- **FR-004**: Retrieval MUST keep dual-plane search. Search text MUST be the query (or task) plus ticket and file paths when those are provided. When the scope is a repository, the system MUST also merge current repository lore that classifies as architecture, decisions, conventions, or gotchas so high-authority briefing knowledge is present even if the task query is narrow. Ranking and budget MUST still use explainable authority (task-specific items remain in the ranked set via search; authority MUST NOT be overridden by agent identity, usage, or popularity).
- **FR-005**: Default retrieval MUST omit superseded and invalidated lore. Conflicts among current statements in the retrieval set MUST be surfaced on the packet. Drift findings MUST NOT be fabricated from implementation analysis.
- **FR-006**: After ranking and token budget, the system MUST classify included items into packet sections using the existing F020 classifier (same section ids and first-match cues). Packet section ids are: `architecture`, `decisions`, `conventions`, `gotchas`, `task_context`. Task Context MUST contain leftover task-relevant items that are not architecture, decisions, conventions, or gotchas (relevance: statement or evidence matches task, query, ticket, or a supplied file path; when no files/ticket/query beyond task exist, leftover search-retrieved items that are not those four briefing types qualify).
- **FR-007**: Sections with zero items MUST be omitted from the packet body (not emitted as empty placeholders). The packet MUST NOT include `observed_drift` or `stale` sections in this slice.
- **FR-008**: The existing flat `items` list MUST remain: budgeted ranked items as in compile v1. Named `sections` are additive. `items` SHOULD list the same statements that were budgeted (not a breaking replacement of `items` by sections-only).
- **FR-009**: Evidence/sources MUST be available without a second knowledge store: unique evidence references from included items, emitted as an additive `sources` list when at least one exists (omitted when empty).
- **FR-010**: Included items MUST retain statement, source plane, authority score, trust band, authority factors, evidence, and provenance refs. Agent identity MUST NOT appear in authority factors and MUST NOT change scores.
- **FR-011**: Ranking MUST use existing explainable authority evaluation. Usage/popularity MUST NOT override authority. Agent identity is provenance for a future feedback feature only.
- **FR-012**: REST MUST keep `POST /v1/context/compile` with additive JSON fields. Existing v1 bodies MUST still return 200 when valid.
- **FR-013**: MCP MUST enrich `memlore.get_for_task` with the same additive fields and payload parity to REST. Local-mode actor rules stay as today. MCP tool count MUST stay 10 (no new tool).
- **FR-014**: CLI MUST expose a developer briefing command `memlore context` requiring `--task` and `--repository`, with optional `--query`, `--ticket`, `--branch`, repeatable `--changed-file` / `--working-file`, `--token-budget`, and `--agent-id`, printing a compact human-readable briefing of populated sections (JSON not required on CLI for this slice).
- **FR-015**: Authorization MUST remain read permission plus membership on the requested scope (same as current compile).
- **FR-016**: Response MUST include existing compile fields plus additive `sections` (non-empty only), optional echoed inputs when provided, optional `sources`, and additive meta `unclassified_count` (items in the budgeted list that matched no packet section).
- **FR-017**: Workflow profiles (`coding` / `debugging` / …) and a new drop-order ladder across section types are out of scope (F022 / F023). This slice does not add a `profile` field.

### Key Entities

- **ContextPacket**: On-read compiled briefing for a task: identity of the request, budgeted items, named sections, conflicts, sources, warnings, token meta.
- **PacketSection**: Named group of compiled items (`architecture`, `decisions`, `conventions`, `gotchas`, `task_context`); omitted when empty.
- **CompileRequest**: Task + scope plus optional query, budget, branch, ticket, files, and agent identity.
- **ConflictGroup**: Disagreeing current statements in one scope (existing compile entity).
- **SourceRef**: Deduplicated evidence reference cited by included items.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An agent can obtain a task briefing that includes at least two distinct populated named sections when corresponding current knowledge exists, without opening underlying source documents.
- **SC-002**: A v1 compile body (task + scope only, or task + scope + query + budget) succeeds and still exposes the original packet fields; new fields are additive.
- **SC-003**: In a fixture with three current items and one superseded item, named sections and `items` contain only current statements; the superseded item never appears.
- **SC-004**: Empty section types are absent from the packet; a fixture with no conventions knowledge has no conventions section.
- **SC-005**: REST and MCP contract tests pass with identical section membership and item ids for the same fixture; MCP tool count remains 10.
- **SC-006**: CLI prints the task identity and at least one section heading for a fixture that has classified items.
- **SC-007**: Two compiles that differ only by agent identity produce the same items, sections, and authority scores.
- **SC-008**: `go test ./...` is green.

## Assumptions

- F007 compile v1 and F020 profile classifier remain the foundation; this slice does not reopen F007 or replace F020.
- Packet section ids reuse F020 ids where they map: Relevant Architecture → `architecture`, Applicable Decisions → `decisions`, Coding Conventions → `conventions`, Known Gotchas → `gotchas`. Task Context is a new id `task_context` for leftover task-relevant items. Conflicts stay the existing top-level `conflicts` array (the Conflicts packet area). Evidence / Sources is the additive `sources` list, not a classified lore section.
- Observed Implementation Drift and Potentially Stale Knowledge are omitted until F050 / an explicit `include_stale` (not added here). Default packet is current-only.
- `branch` is stored on the request/echo only; lore is not yet branch-scoped.
- File lists are opaque path strings (no repository checkout or code analysis).
- Agent identity is an opaque string (`agent_id`); it is echoed when provided and not persisted as a new table in this slice (F060 will persist feedback later).
- Repository-scope merge uses list-by-scope current lore filtered to F020 briefing classes (architecture, decisions, conventions, gotchas), unioned with task search hits, then existing rank/dedup/budget. Non-repository scopes stay search-only (v1).
- Token estimation and default budget match compile v1 (character/4 plus per-item overhead). Retrieval limit stays the existing compile default.
- CLI `context` uses the same local Postgres/graph wiring as `memlore profile` / `memlore mcp` (not an HTTP client to `serve`).
- Web UI is out of scope (F120). F022 profiles and F023 drop-order ladder are out of scope.
- Graph-service remains optional; degradation warning is reused.
