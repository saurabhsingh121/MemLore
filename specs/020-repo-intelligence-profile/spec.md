# Feature Specification: Repository Intelligence Profile (F020)

**Feature Branch**: `020-repo-intelligence-profile`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F020  
**Depends on**: F006 (retrieval), F007 (compile v1), F003 (authority), F114 (membership)  
**Input**: Compact compiled engineering overview for a repository from existing governed knowledge — not a second knowledge store and not invented content.

## Goal

Give a human or coding agent a compact **repository intelligence profile**: the
engineering context they need before exploring dozens of files. The profile is
compiled on read from current lore and graph knowledge for that repository,
token-budgeted, evidence-bearing, and silent about sections that have no
supporting knowledge.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Compact repository briefing (Priority: P1)

A developer or coding agent asks for an overview of a repository they are
authorized to read. MemLore returns a structured profile with only those
engineering sections that have supporting current knowledge (for example
architecture, important decisions, technologies, conventions, ownership,
gotchas, migrations, hotspots, operational risks, recent changes, related
services). Each included item shows the statement plus evidence when the
source had evidence. Empty sections are omitted. Nothing is invented.

**Why this priority**: This is the product value of F020 — useful overview
without archaeology — and it can ship on the existing retrieval/compile
foundation.

**Independent Test**: Store several current lore entries in one repository
scope covering at least two different section types; request a profile; verify
named sections appear with those statements and evidence, and that a section
with no matching knowledge is absent.

**Acceptance Scenarios**:

1. **Given** current lore in repository `payments-api` that includes an
   architecture statement and an ADR-backed decision, **When** a profile is
   requested for that repository, **Then** the profile identifies the
   repository and includes architecture and decisions sections containing those
   statements with evidence links where present.
2. **Given** no current knowledge that maps to "active migrations", **When** a
   profile is requested, **Then** the profile does not include a migrations
   section (omitted, not filled with a placeholder or guessed text).
3. **Given** a repository with no current lore and no graph hits, **When** a
   profile is requested by an authorized caller, **Then** the response succeeds
   with the repository identity, no invented sections, and empty/zero item
   counts rather than a not-found error.

---

### User Story 2 - Trust, history, and conflicts stay honest (Priority: P1)

The profile uses the same honesty rules as compiled task context: superseded
and invalidated knowledge is omitted from sections; disagreeing current
statements are surfaced as conflicts; each item keeps authority/trust metadata
and evidence. Usage popularity is not a ranking override.

**Why this priority**: Constitution V and VI — a pretty overview that hides
stale truth or invents canonical facts would be a product defect.

**Independent Test**: Mix current, superseded, and conflicting lore in one
repository; request a profile; assert stale items are absent from sections,
conflicts are listed, and remaining items carry trust metadata.

**Acceptance Scenarios**:

1. **Given** a superseded lore entry and a current successor in the same
   repository, **When** a profile is requested, **Then** only the current
   statement appears in sections.
2. **Given** two current disagreeing statements in that repository, **When** a
   profile is requested, **Then** both sides appear in a conflicts list and
   neither side is silently dropped from that list.
3. **Given** an unverified agent inference and a verified human-authored
   decision, **When** both are classified into the same section, **Then** the
   verified decision is ordered ahead of the unverified inference.

---

### User Story 3 - Same profile on REST, MCP, and CLI (Priority: P2)

Authorized callers can request the profile through REST, the MCP tool
`memlore.repo_profile`, and the CLI `memlore profile`. JSON payloads match
across REST and MCP. CLI prints a compact human-readable briefing of the same
sections. Membership and read authorization match other lore read operations.

**Why this priority**: Constitution: CLI + REST are sufficient; MCP where
agents participate. Parity prevents agent/human split-brain.

**Independent Test**: Contract tests for REST and MCP with the same inputs;
CLI unit test of formatted output from a fixture profile (no live server
required for formatting).

**Acceptance Scenarios**:

1. **Given** the same repository scope and token budget, **When** REST and MCP
   are called, **Then** section names, item ids, statements, and meta counts
   match.
2. **Given** a missing repository key, **When** any surface is called, **Then**
   the caller receives a validation error (not an empty invented profile).
3. **Given** OIDC membership mode and a caller without access to the
   repository, **When** they request a profile, **Then** access is denied the
   same way as compile/search for that scope.
4. **Given** a compiled profile with a decisions section, **When** CLI formats
   it, **Then** the printed briefing includes the repository identity and that
   section's statements.

---

### Edge Cases

- Scope kind other than `repository` is rejected with a validation error.
- Token budget omitted uses the same default as task compile (4096 estimated
  tokens). Budget that cannot fit all classified items drops lower-ranked
  items; conflicts remain listed even if a side is omitted from sections.
- Graph-service unavailability yields the existing retrieval warning and a
  governance-only profile when governance hits exist.
- Statements that match no section cues are omitted from named sections and
  counted as unclassified in meta (not dumped into a generic "other" memory
  bucket).
- Duplicate equivalent statements across governance and graph are deduplicated
  as in compile v1 (governance preferred).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Callers MUST request a profile for exactly one repository
  identity (`scope.kind` = `repository` and a non-empty `scope.key`).
- **FR-002**: The system MUST compile the profile on read from existing current
  knowledge for that repository (governance + graph). It MUST NOT persist a
  second profile store or invent statements not present in retrieved
  knowledge.
- **FR-003**: The system MUST classify retrieved items into at most one of these
  section ids when cues match: `architecture`, `decisions`, `technologies`,
  `conventions`, `ownership`, `gotchas`, `migrations`, `hotspots`,
  `operational_risks`, `recent_changes`, `related_services`.
- **FR-004**: Classification MUST use origin, evidence type, and statement
  cues in a documented deterministic order (first match wins). ADR evidence or
  architecture-decision origin maps to `decisions`.
- **FR-005**: Sections with zero classified items MUST be omitted from the
  profile body (not emitted as empty arrays/placeholders).
- **FR-006**: Included items MUST retain statement, source plane, authority
  score, trust band, authority factors, evidence, and provenance refs as in
  compiled task context.
- **FR-007**: Default retrieval MUST omit superseded and invalidated lore.
  Conflicts among current statements in the retrieval set MUST be surfaced on
  the profile.
- **FR-008**: An explicit token budget MUST cap included section items using
  the same estimation rules as task compile. Default budget is 4096.
- **FR-009**: Ranking MUST use existing explainable authority evaluation.
  Usage/popularity MUST NOT override authority.
- **FR-010**: REST MUST expose `POST /v1/repository-profile` with the same
  authorization as `POST /v1/context/compile`.
- **FR-011**: MCP MUST expose read-only `memlore.repo_profile` with payload
  parity to REST. Local mode follows existing actor rules for read tools that
  take `actor_id` (same as `get_for_task`).
- **FR-012**: CLI MUST expose `memlore profile --repository <key>` (optional
  `--token-budget`) printing a compact human-readable briefing. JSON is not
  required on CLI for this slice.
- **FR-013**: Response MUST include repository identity, included sections,
  conflicts, warnings, and meta (`token_budget`, `estimated_tokens`,
  `items_included`, `items_total_ranked`, `unclassified_count`).
- **FR-014**: `get_for_task` and knowledge search MUST remain unchanged except
  for MCP tool-list count increasing by one.

### Key Entities

- **RepositoryProfile**: On-read briefing for one repository: identity,
  classified sections, conflicts, warnings, token meta.
- **ProfileSection**: Named group of profile items; omitted when empty.
- **ProfileItem**: One classified knowledge statement with evidence and
  authority metadata (same fields as a compiled context item plus `section`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A reader can obtain a repository briefing that includes at least
  two distinct populated sections when corresponding current knowledge exists,
  without opening the underlying source documents.
- **SC-002**: In a fixture with three current items and one superseded item,
  the profile sections contain only the three current items (or fewer if
  budget/classification drops some); the superseded item never appears in a
  section.
- **SC-003**: Empty-knowledge repositories return a successful empty briefing
  in under a typical interactive wait (same order as compile); no fabricated
  architecture or decision text.
- **SC-004**: REST and MCP contract tests pass with identical section
  membership for the same fixture.
- **SC-005**: CLI prints repository identity and at least one section heading
  for a fixture that has classified items.
- **SC-006**: `go test ./...` is green.

## Assumptions

- Repository identity is the existing lore scope `{ kind: repository, key }`
  (for example `github.com/acme/payments`). No new repository registry is
  required.
- Profile retrieval reuses dual-plane search with a default overview query
  covering the section themes; callers do not supply a free-text query in this
  slice (F021 can add task-specific bootstrap later).
- Classification is cue-based and conservative: unmatched items are omitted
  from named sections (counted in `unclassified_count`) rather than forced
  into a generic dump.
- Token estimation and default budget match compile v1 (character/4 plus
  per-item overhead).
- Graph-service remains optional; degradation warning is reused.
- Web UI is out of scope (F120). Git ingest (F030–F032) is out of scope;
  profiles get richer as ingest lands later.
- CLI `profile` uses the same local Postgres/graph wiring as `memlore mcp`
  (not an HTTP client to `serve`).
