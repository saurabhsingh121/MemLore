# Feature Specification: First-Class Decision Model (F040)

**Feature Branch**: `040-decision-model`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F040  
**Depends on**: F001 (lore), F003 (authority), F008/F110 (supersede/invalidate DONE), F007/F021 (compile / `get_for_task` DONE), F032 (ADR ingest DONE — trusted-source policy remains), F035 (review queue DONE — promotion ≠ decision), F114 (membership)  
**Does not reopen**: F007 compile v1, F020 repository profile, F021 named packet, F030 git ingest, F031 PR ingest, F032 ADR ingest, F035 review queue (additive decision fields/routes only). Git SHA / PR number / ADR path idempotency tables stay frozen.  
**Input**: Agents must answer “why Kafka?” as an engineering **decision**, not a lore snippet. Provide a dedicated decision domain: question, choice, rationale, alternatives, consequences, owner, date, affected components, evidence, superseded predecessor, and current validity — without inventing a second competing truth beside trusted ADR lore.

## Goal

MemLore already stores engineering knowledge as lore entries (one statement + origin + evidence + lifecycle). Accepted ADRs become verified `architecture_decision` lore (F032). Git and PR extracts become observational lore and may be promoted through the review queue (F035). Context packets already have a `decisions` section, filled by heuristics (ADR evidence, architecture-decision origin, or keywords).

That is not enough. A decision is not a snippet. Humans and agents need a **first-class Decision**: the problem that was being solved, the choice, why, what else was considered, what follows, who owns it, when, what it affects, the evidence, whether a later decision replaced it, and whether it is still the current record.

F040 is that domain. It is Decision Intelligence (Epic E), not capture (Epic C) and not the signature `memlore why` command (F044). F044 comes after this model exists.

This slice **does** let operators create, retrieve, list current, and supersede decisions on REST and CLI. It **does** make current decisions participate in compile / `get_for_task` as first-class `decisions` section content. It **does** project existing F032 accepted-ADR lore as decisions so the team’s trusted corpus is queryable without duplicating the same choice as a second current fact. It **does not** auto-upgrade observational git/PR lore or F035 Accept into Decisions. It **does not** add an MCP write tool or an 11th MCP tool.

## Clarifications

### Session 2026-09-01

Decisions encoded from the F040 implementation prompt. No remaining product questions block planning.

- **Aggregate vs lore enrichment**: Decision is a dedicated first-class record with structured fields (question, choice, rationale, alternatives, consequences, owner, date, affected components). Alternatives and consequences MUST NOT be stuffed only into a lore statement. Each Decision is linked to exactly one lore entry so compile, authority, and explain continue to work on governed lore. Human create dual-writes Decision + lore. Supersede creates a new Decision (and new lore) rather than rewriting in place (constitution VI, F110).
- **Relationship to F032**: Current accepted-ADR lore **appears as** F040 decisions (same identity as the lore entry, evidence type `adr`). F040 MUST NOT duplicate that statement as a second current fact. Human-recorded decisions are distinguishable from ingest-created ADR decisions (constitution V). F032 trusted-source auto-verify is unchanged. A later enrichment mapper MAY fill empty optional fields on ADR-projected decisions; this slice does not re-parse ADR files.
- **F035 Accept**: Accept continues to create `human_verified` / `human_authored` lore, not a Decision. This slice does not auto-wrap every Accept. Reviewers who are recording a decision use create-decision explicitly. A review-queue “record as decision” action is deferred.
- **Validity**: “Current” means the decision is not superseded and not invalidated. This slice does not invent still-valid vs implementation-drift (F050) and does not claim code compliance.
- **MCP**: Tool count stays 10. No new `get_decision` / `remember_decision` tool. Current decisions feed the existing `get_for_task` `decisions` section. `explain` / `get` continue to work on the linked lore id. Mutating create/supersede is human (CLI/REST), not an agent-default tool.
- **F042 alternatives**: Stored as fields on the Decision in this slice (ordered considered options: label required, optional note). Not a separate alternatives product.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Record a human engineering decision (Priority: P1)

An authorized writer who belongs to a repository records a decision: the problem (for example “how should payment events leave the service?”), the choice (“transactional outbox”), optional rationale, optional alternatives considered (for example SQS, dual-write), optional consequences, owner, date, optional affected components, and optional evidence. MemLore stores that as a **current Decision**, distinguishable from ingest-created ADR decisions and from generic lore snippets. Evidence and owner are preserved. The decision is also represented as governed lore so authority and compile can see it. History is empty until someone supersedes it.

**Why this priority**: Without a human create path, F040 cannot exist for teams that have not ingested ADRs, and agents cannot later answer “why Kafka?” from a structured decision.

**Independent Test**: Create a decision with required fields and one alternative; retrieve it by id; assert question, choice, owner, date, alternative, and source kind `human` are present; assert it is current; assert a linked lore entry exists and is not `repository_observation`.

**Acceptance Scenarios**:

1. **Given** a writer with membership on repository `github.com/acme/payments`, **When** they record a decision with question “How should payment events be published?”, choice “Transactional outbox”, owner `alice`, and alternative “Dual-write to the topic”, **Then** MemLore stores one current Decision with those fields, source kind human, and preserved owner and evidence (if any).
2. **Given** that Decision, **When** a caller retrieves it by id, **Then** they see the structured fields, current validity, and that it is not an ingest-created ADR decision.
3. **Given** that create, **When** the write completes, **Then** the Decision, its linked lore, audit, and knowledge-plane outbox occur in one unit of work; the graph worker is not called from the create command.
4. **Given** `POST /v1/lore-entries` / `memlore.remember`, **When** F040 ships, **Then** those paths still require origin `human_authored` and MUST NOT become the Decision create path (they still create lore snippets, not Decisions).

---

### User Story 2 - Trusted ADR lore is queryable as decisions without a second current fact (Priority: P1)

A reader lists current decisions for a repository that already has F032 accepted-ADR lore. Each current accepted-ADR lore entry appears as a Decision (source kind ADR, evidence type `adr`). The same choice is not stored twice as two current facts. Human-recorded decisions in the same repository appear alongside ADR-projected ones and remain distinguishable. Draft/proposed ADR files that F032 skipped still do not appear. Observational git/PR lore and F035 pending queue items do not appear as Decisions.

**Why this priority**: Constitution V and the flywheel — F032 already captured the team’s decision corpus. F040 must project or link that corpus, not invent a competing truth or undo trusted-source ingest.

**Independent Test**: Seed one current verified ADR lore entry and one human-recorded Decision in the same repository; list current decisions; assert both appear once; assert the ADR item has evidence `adr` and source kind ADR; assert the human item is not labeled ingest-created; assert a git observation and a pending review item are absent.

**Acceptance Scenarios**:

1. **Given** current F032 accepted-ADR lore “Use PostgreSQL as the system of record” with evidence `adr` `0001-use-postgres`, **When** a member lists current decisions for that repository, **Then** that choice appears exactly once as a Decision with ADR evidence and ingest-created source, not as a second copied statement.
2. **Given** that ADR-projected Decision, **When** a caller retrieves it by the same id as the lore entry (or the Decision’s public id that identifies that lore), **Then** they get that Decision; a second current Decision with the same ADR identity MUST NOT exist.
3. **Given** a current human-recorded Decision in the same repository, **When** current decisions are listed, **Then** both the human Decision and the ADR-projected Decision appear, and callers can tell which is which.
4. **Given** unverified git or PR observational lore and a pending F035 review item in the same repository, **When** current decisions are listed, **Then** those items are not Decisions.
5. **Given** F032 ingest of an unchanged accepted ADR after F040 ships, **When** current decisions are listed, **Then** the ADR still appears once (idempotent projection; no duplicate current fact).

---

### User Story 3 - Supersede a decision without deleting history (Priority: P1)

An authorized writer replaces a current Decision with a new choice (or a refined statement of the same choice). MemLore records a **new** current Decision. The predecessor remains retrievable as history, linked as superseded by the successor. Linked lore is superseded the same way (not overwritten, not deleted). ADR-projected decisions may be superseded by a human-recorded successor; that is the explicit path to replace trusted ingest with a later human decision. Invalidated or already-superseded decisions cannot be superseded.

**Why this priority**: Constitution VI — do not overwrite historical truth. F044 and F041 will need previous decisions to remain discoverable.

**Independent Test**: Create a human Decision A; supersede it with Decision B; get A (superseded, still stored) and B (current); list current (only B). Separately supersede an ADR-projected Decision with a human successor; assert the ADR lore is superseded, not deleted, and list current shows the human successor once.

**Acceptance Scenarios**:

1. **Given** a current human Decision A, **When** a writer with membership supersedes it with question/choice for successor B, **Then** B is current, A is not current, A remains gettable, A points at B as the superseding decision, and neither row is deleted.
2. **Given** a current ADR-projected Decision, **When** a writer supersedes it with a human-recorded successor, **Then** the successor is current and human-sourced, the ADR-backed predecessor remains gettable as history with its `adr` evidence, and there is not a second current copy of the old ADR choice.
3. **Given** an already-superseded or invalidated Decision, **When** supersede is requested, **Then** the system returns a validation error and creates no successor.
4. **Given** that supersede write, **When** it completes, **Then** successor Decision, predecessor link, linked lore supersession, audit, and outbox occur in one unit of work; the graph worker is not called from the command.
5. **Given** list-current after supersede, **When** a member lists current decisions, **Then** only the successor appears; history remains available via get-by-id of the predecessor.

---

### User Story 4 - Current decisions brief agents in the existing decisions packet section (Priority: P1)

An agent (or operator) compiles context / `get_for_task` for a repository task. Current first-class Decisions in scope appear in the existing named `decisions` section — not as undifferentiated snippets and not as a new section id. Compile ranking formulas are unchanged. Current Decisions (including ADR-projected ones) outrank leftover unverified observational lore. F032 accepted-ADR authority is not weakened: verified ADR-backed architecture still outranks leftover git/PR observations. Keyword-heuristic leftover lore may still fill other sections; it MUST NOT duplicate a Decision already represented in `decisions`.

**Why this priority**: F044 (`memlore why`) needs a real decision in the briefing. F021 already promised a `decisions` section; F040 makes that section first-class.

**Independent Test**: Seed a current human Decision, a current ADR-projected Decision, and an unverified git observation; compile / `get_for_task`; assert both decisions appear under section `decisions`; assert the observation is not treated as equal to those decisions; assert ADR ranking versus observations is not weaker than before F040.

**Acceptance Scenarios**:

1. **Given** a current human Decision and a current ADR-projected Decision in repository `github.com/acme/payments`, **When** context is compiled / `get_for_task` runs for a matching task, **Then** both appear in the existing `decisions` section (same section id as F021).
2. **Given** those current Decisions plus leftover unverified git/PR observations, **When** compile ranks items, **Then** the current Decisions outrank those observations; ranking formulas themselves are not modified.
3. **Given** F032 accepted-ADR lore that F040 projects as a Decision, **When** compile runs, **Then** that ADR-backed item still outranks leftover unverified observations (authority not accidentally weakened).
4. **Given** a lore snippet that would previously match the `decisions` keyword heuristic but is **not** a first-class Decision, **When** a first-class Decision already represents that choice, **Then** the snippet is not shown as a second `decisions` item for the same current fact.
5. **Given** MCP `get_for_task`, **When** tools are enumerated, **Then** the tool count remains 10 and there is no new decision-specific tool.

---

### User Story 5 - Operators use CLI and REST; agents do not mutate decisions (Priority: P2)

Operators create, get, list-current, and supersede through CLI and REST. Local mode requires an actor header or CLI `--actor`. Membership scopes every operation. Get-by-id for a foreign tenant returns not found (no leak). Readers with membership may get and list; they must not create or supersede. MCP tool count remains 10; there is no new agent tool to create or supersede a Decision. There is no product web UI.

**Why this priority**: Constitution: CLI + REST are sufficient governance surfaces; web UI is F120. Agents must not silently mint canonical decisions.

**Independent Test**: CLI create/get/list/supersede on a fixture; REST create/get/list/supersede plus membership denial and cross-tenant 404; MCP tool enumeration still 10; existing ingest and review routes still pass.

**Acceptance Scenarios**:

1. **Given** a writer with membership, **When** they run `memlore decision create` with required fields (or the equivalent CLI flags), **Then** a current Decision is stored for that repository.
2. **Given** a Decision id, **When** they run `memlore decision get <id>` or `memlore decision list --repository <key>`, **Then** they see that Decision or the current list in a human-readable summary (JSON not required).
3. **Given** a current Decision id, **When** they run `memlore decision supersede <id>` with successor fields, **Then** supersede occurs as in story 3.
4. **Given** REST create (write), get/list (read), and supersede (write), **When** a principal lacks membership, **Then** the call is denied; get-by-id across tenants returns not found; repository keys are not placed in the URL path (decision id in the path is allowed).
5. **Given** the MCP tool list, **When** tools are enumerated, **Then** the count remains 10 and there is no new tool that creates, lists, or supersedes Decisions.
6. **Given** existing git, PR, ADR ingest, and review-queue CLI/REST, **When** F040 ships, **Then** those surfaces still behave as specified in F030–F035.

---

### Edge Cases

- Missing required fields on human create (question, choice, owner): validation error; nothing stored.
- Blank actor on mutate: validation error (local mode still requires `X-Memlore-Actor` / `--actor`).
- Unknown decision id: not found; no write.
- Get-by-id for a Decision in a repository the caller does not belong to: not found (hard fail; no leak).
- Reader role: may get and list; must not create or supersede.
- Admin without membership: existing F114 admin bypass still applies (JWT admin bypasses membership), same as other writes.
- Supersede of unknown id: not found.
- Concurrent supersede of the same current Decision: one successor; the second request is a validation error or conflict — never two current successors for one predecessor.
- Human create that restates an existing current ADR decision in different words: allowed as a separate Decision (conflict surfacing remains F112); F040 MUST NOT silently merge or overwrite the ADR fact. Operators who intend to replace the ADR decision MUST supersede it.
- F035 Accept of a git/PR extract: still lore only; no Decision is created.
- In-place `verify` of observational lore: still not a Decision.
- Deprecated/superseded ADR lore (F032 historical, not current canonical): not listed as current Decisions; may be gettable as historical if projected, otherwise omitted from list-current.
- Graph-service down: Decision, lore, audit, and outbox still commit on the governance plane; graph catch-up remains the existing worker.
- `memlore worker` remains outbox → graph publisher only; decision writes are not folded into the worker.
- Empty current list: success with zero items (not an error).
- Decision date omitted on human create: recorded-at is used as the decision date.
- Alternatives: empty list is valid; a listed alternative MUST have a non-empty label; optional note may be omitted.
- Web UI: out of scope (F120).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST represent an engineering **Decision** as a first-class record, distinct from a generic lore snippet. A Decision MUST be able to represent: question/problem, decision (the choice), rationale, alternatives considered, consequences, owner, date, affected components, evidence, superseded predecessor, and current validity. Optional fields MAY be empty; required fields on human create are question, choice, and owner.
- **FR-002**: Alternatives MUST be stored as structured fields on the Decision (ordered considered options: label required, optional note). Consequences MAY be a single optional text field in this slice. The system MUST NOT rely on stuffing alternatives/consequences only into a lore statement as the Decision model.
- **FR-003**: Human create MUST dual-write a Decision and a linked lore entry in the same unit of work so compile, authority, and explain continue to operate on governed lore. `CreateLore` / MCP `remember` / `POST /v1/lore-entries` remain human-authored lore only and MUST NOT be loosened or become Decision create.
- **FR-004**: Current F032 accepted-ADR lore (`architecture_decision` origin, `adr` evidence, current/not invalidated) MUST be queryable as Decisions. Projection MUST use the lore identity so the choice is not duplicated as a second current fact. Human-recorded vs ingest-created ADR Decisions MUST be distinguishable.
- **FR-005**: Observational git/PR lore, F035 pending review items, and F035 Accept successors MUST NOT be auto-upgraded into Decisions. Accept remains lore promotion only.
- **FR-006**: List-current MUST return Decisions that are not superseded and not invalidated, scoped to the requested repository (membership-visible). Historical (superseded/invalidated) Decisions MUST remain gettable by id when the caller is authorized.
- **FR-007**: Supersede MUST create a new Decision (and linked lore) and mark the predecessor Decision (and linked lore) superseded — never in-place rewrite or delete. Already-superseded or invalidated Decisions MUST be rejected. ADR-projected Decisions MAY be superseded by a human-recorded successor.
- **FR-008**: Create and supersede MUST write governance records + audit + knowledge-plane outbox in one unit of work. Commands MUST NOT call Graphiti / graph-service directly. The existing worker remains outbox-only.
- **FR-009**: Evidence and owner MUST be preserved on retrieve after create and after supersede (successor stores the owner and evidence supplied for the successor; predecessor evidence remains on the predecessor).
- **FR-010**: Current first-class Decisions MUST participate in compile / `get_for_task` in the existing `decisions` section (same section id as F021). This slice MUST NOT add a new packet section id and MUST NOT change compile ranking formulas. Characterization tests MUST assert: (a) current Decisions outrank leftover unverified observations; (b) F032 accepted-ADR authority versus leftover observations is not weakened; (c) a Decision already in `decisions` is not also shown as a duplicate snippet in that section.
- **FR-011**: CLI MUST expose `memlore decision create`, `memlore decision get <id>`, `memlore decision list --repository <key>`, and `memlore decision supersede <id>` with `--actor` for mutating commands in local mode. Output is human-readable (JSON not required).
- **FR-012**: REST MUST provide membership-scoped `POST /v1/decisions` (write), `GET /v1/decisions/{id}` (read), `GET /v1/decisions` list-current (read; `scope_kind` + `scope_key`), and `POST /v1/decisions/{id}/supersede` (write). Repository keys MUST NOT be placed in the URL path. Cross-tenant get-by-id MUST return not found.
- **FR-013**: Create and supersede require **write** permission (writer or admin) plus F114 membership on the repository scope. Get and list require **read** plus membership. Local mode: mutating routes require `X-Memlore-Actor`; membership off; actor is trusted admin (same as F035). CLI uses the same local Postgres wiring as `memlore review` / ingest (not an HTTP client to `serve`).
- **FR-014**: MCP tool count MUST remain 10. No new MCP tool. No MCP write that creates or supersedes Decisions. `memlore.remember` MUST NOT gain Decision semantics. Agent read of decisions is via existing `get_for_task` (and `get` / `explain` of linked lore).
- **FR-015**: F032 trusted-source policy MUST remain: accepted ADRs stay verified `architecture_decision` with `adr` evidence. This slice MUST NOT undo auto-verify, MUST NOT send ADR lore through the F035 queue, and MUST NOT require a human to re-enter ADR text to make it a Decision.
- **FR-016**: Operational tables for Decisions MUST be additive. They MUST NOT overload `lore_review_decisions`, git SHA, PR number, or ADR path/checksum ingest tables.
- **FR-017**: Current validity in this slice is solely “not superseded and not invalidated.” The system MUST NOT claim implementation compliance or drift status (F050).
- **FR-018**: F041 decision timeline as a separate product, F043 impact graph / graph-service relationships, F044 `memlore why`, F022 packet profiles, F033 docs ingest, F034 chat ingest, F050 drift, and F120 web UI are out of scope.

### Key Entities

- **Decision**: A first-class engineering decision in a repository scope: question/problem, choice, optional rationale, optional alternatives, optional consequences, owner, date, optional affected components, optional evidence, source kind (human vs ADR ingest), link to exactly one lore entry, optional predecessor/successor supersession, and derived current validity.
- **DecisionAlternative**: A considered option on a Decision: required label, optional note (why it lost / tradeoff). Ordered. F042-as-fields, not a separate aggregate.
- **DecisionSource**: Distinguishes human-recorded Decisions from ingest-created ADR projections (constitution V). Not a popularity or confidence score.
- **LinkedLore**: The governed lore row that carries origin, verification, evidence, and lifecycle for compile/authority/explain. Human create writes it; ADR projection reuses the existing F032 lore row.
- **CurrentDecisionSet**: The membership-scoped list of Decisions that are not superseded and not invalidated.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator can record a Decision with question, choice, owner, and one alternative, then retrieve that same Decision with those fields and human source kind, in one create-and-get pass.
- **SC-002**: A repository with one current accepted-ADR lore entry and one human-recorded Decision lists exactly those two as current Decisions (each once); a git observation and a pending review item are not listed.
- **SC-003**: Superseding a current Decision yields a new current successor; the predecessor remains retrievable, is not deleted, and is absent from list-current.
- **SC-004**: Compile / `get_for_task` for that repository places current Decisions in the existing `decisions` section; those Decisions outrank leftover unverified observations; ADR-backed authority versus observations is not weaker than before this feature.
- **SC-005**: REST contract tests cover create, get, list-current, supersede, membership denial, and cross-tenant get-by-id as not found; CLI covers create, get, list, and supersede on a fixture.
- **SC-006**: MCP tool count remains 10; `remember` / `POST /v1/lore-entries` still create human-authored lore only, not Decisions; F035 Accept still does not create a Decision.
- **SC-007**: Evidence and owner on a retrieved Decision match what was stored at create (and on the successor after supersede).
- **SC-008**: Create and supersede remain durable on the governance plane when the graph service is unavailable (outbox records the sync; the command does not call the graph).
- **SC-009**: `go test ./...` and `go vet ./...` are green.

## Assumptions

- **Dedicated Decision + lore pointer (v1 public contract)**: Human create persists a Decision aggregate and a linked lore entry. ADR-backed decisions are the existing current F032 lore rows exposed through the Decision API (projection), not a copied second statement. Plan may materialize a projection row that points at lore; product behavior is one current Decision identity per current ADR lore entry.
- **Human-created lore origin**: Linked lore for a human-recorded Decision uses origin `human_authored` and is created verified only if specify-adjacent policy says so. Default: human-recorded Decisions are **verified** (the writer is recording an explicit decision, not an extract) with actor = creating operator, remaining auditable. This is not the F032 trusted-source exception and not F035 Accept. It MUST stay distinguishable from `architecture_decision` + `adr`.
- **ADR projection fields**: Choice/statement comes from the lore statement. Question MAY be empty or a title-derived fallback already present in the statement; this slice does not re-parse ADR files to fill alternatives/consequences. Empty optional fields on ADR-projected Decisions are valid.
- **Decision id for ADR projection**: Public id equals the lore entry id so get-by-id is stable and non-duplicative.
- **Human Decision id**: New unique id; linked lore may share that id (one identity) to keep get/explain simple.
- **Scope**: Repository-scoped only in v1 (same as ingest/review). Other scope kinds are rejected on create/list.
- **List-current default**: Current only. Get-by-id returns historical records. No separate timeline product (F041).
- **Invalidate**: Operators MAY invalidate linked lore via existing lore invalidate; the Decision then is not current. This slice does not add a separate `decision invalidate` command (use supersede for replacement; use existing lore invalidate for “this was never right” if needed).
- **MCP**: Fold into `get_for_task` / `explain`; stay at 10 tools.
- **F042**: Alternatives-as-fields only; no query API “list all alternatives across decisions” in this slice.
- **Compile**: Feed existing `decisions` section; do not add `decision_records` as a new section id.
- **Conflict**: Two current Decisions that disagree remain both current (F112); F040 does not auto-resolve.
- **Web UI**: Out of scope (F120).
- **Out of slice**: F041, F043, F044, F022, F033, F034, F050, auto-wrap of F035 Accept, observational auto-canonical, graph-service relationship types, `memlore why`.
- Historical IDs F004, F107, and frozen F101–F114 are not reused; this is product F040.
