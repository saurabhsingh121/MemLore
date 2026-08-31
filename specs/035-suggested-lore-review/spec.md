# Feature Specification: Suggested Lore Review Queue (F035)

**Feature Branch**: `035-suggested-lore-review`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F035  
**Depends on**: F030 (git ingest DONE), F031 (PR ingest DONE), F032 (ADR ingest DONE — trusted-source policy remains), F003 (authority), F008/F110 (supersede/invalidate), F010 (RBAC), F114 (membership)  
**Does not reopen**: F007 compile v1, F020 repository profile, F021 named packet, F030 git ingest, F031 PR ingest, F032 ADR ingest (additive review-queue fields/routes only). Git SHA / PR number / ADR path idempotency tables stay frozen.  
**Input**: Automatically extracted knowledge must not silently become authoritative. Provide a human Accept / Edit / Reject path so git and PR observational lore can be promoted with human verification provenance, without using in-place verify as a fake promotion.

## Goal

MemLore already captures git- and PR-derived observations as **candidate lore**. Those observations remain distinguishable from human-authored and human-verified knowledge. Without this feature there is no promotion path except marking verification status on the same observational row, which does not record that a human accepted the claim and does not change origin.

F035 is that human review queue. Authorized reviewers list pending suggestions, **Accept** a faithful extract as human-verified knowledge, **Edit then Accept** a corrected statement as human-authored knowledge, or **Reject** an extract so it is not offered again. Nothing in the queue is canonical until Accept. Accepted Architecture Decision Records already trusted by F032 stay trusted and **never** appear as review items. Existing `verify` remains a verification-status flip only and is **not** Accept.

This is promotion, not capture. F033 documentation ingest, F034 chat ingest, F040 first-class decision model, F022 packet profiles, and F120 product web UI are out of scope.

## Clarifications

### Session 2026-09-01

Decisions encoded from the F035 implementation prompt. No remaining product questions block planning.

- **Accept vs verify**: Accept is not `verify`. Verify only sets verification status; origin stays observational. Accept records human verification provenance by creating a new current lore entry and superseding the observational predecessor (F110). The historical observation is not rewritten.
- **Accept as-stated**: Successor origin is `human_verified`, statement matches the extract, evidence from the predecessor is preserved, verification is `verified`, actor is the reviewer.
- **Edit then Accept**: If the reviewer supplies a statement that differs from the extract, successor origin is `human_authored` (the human wrote the trusted wording). Evidence from the predecessor is still preserved. The raw extract is not labeled as if a human originally authored it.
- **Reject**: Records a negative review decision for the extract identity. Observational lore is **not** invalidated or deleted (the observation that git/PR said this remains historically true). The extract is not re-queued.
- **Queue vs ingest listing**: Dedicated review-queue workflow. `GET /v1/ingest/candidates` remains ingest listing, not Accept/Edit/Reject. F032 accepted-ADR rows listed there are not review items.
- **Queue identity**: Review items are the current observational lore entries (git `commit` / PR `pr` evidence). A review-decision overlay stores accept/reject so lifecycle does not collide with ingest listing.
- **Reject/re-ingest key**: repository scope + evidence type + evidence identity + statement checksum. Re-ingest of an already-rejected extract does not resurrect a pending item.
- **Confidence**: Producers do not store a 0–1 score. Queue MUST NOT invent confidence. The field is omitted until a producer supplies it.
- **Reason**: Git/PR statements are the lore statement. Queue MUST NOT invent a separate reason. Reason is omitted in this slice; later producers MAY supply a distinct reason without changing Accept/Reject semantics.
- **Uncertain ADR extracts**: F032 continues to skip them. This slice does not enqueue skipped ADRs and does not change F032. The queue is designed so a later producer can reuse it.
- **Who may Accept/Reject**: Writer or admin plus F114 membership (same as ingest trigger and supersede). Not the admin-only verify permission. Readers with membership may list the queue.
- **MCP**: Tool count stays 10. No new list or mutate review tool. Mutating Accept/Reject is a human (CLI/REST) action.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reviewers see pending suggested lore (Priority: P1)

An authorized reader or writer who belongs to a repository lists the suggested-lore review queue for that repository. Each pending item shows the candidate statement, evidence (for example a PR number or commit SHA), and source type (git commit or pull request). Confidence and a separate reason appear only when they were actually stored — they are not fabricated. Accepted ADR lore is absent. Items already accepted or rejected are absent from the default pending list. A reviewer without membership on that repository cannot see its queue. Cross-tenant listing is denied.

**Why this priority**: Reviewers cannot promote what they cannot see. The queue is the product surface that separates candidates from trusted knowledge.

**Independent Test**: Seed one git observation, one PR observation, one F032 accepted-ADR entry, and one already-rejected extract in the same repository; list the pending queue; assert the two observations appear with statement, evidence, and source type; assert the ADR and rejected extract do not; assert a foreign-repository caller is denied.

**Acceptance Scenarios**:

1. **Given** current unverified observational lore from git ingest (evidence type `commit`) and PR ingest (evidence type `pr`) in repository `github.com/acme/payments`, **When** a member lists the pending review queue for that repository, **Then** both items appear with their statement, evidence, and source type, and neither is presented as canonical or high-authority.
2. **Given** verified `architecture_decision` lore from F032 accepted-ADR ingest in the same repository, **When** the pending queue is listed, **Then** that ADR lore is not a review-queue item.
3. **Given** a producer that did not store confidence or a separate reason, **When** a pending item is shown, **Then** confidence and reason are omitted rather than filled with guessed values.
4. **Given** a principal without membership on the requested repository (OIDC membership mode), **When** they list the queue, **Then** access is denied the same way as other reads for that scope; no other tenant’s items leak.
5. **Given** ingest candidates listing (`evidence_type=commit|pr|adr`), **When** F035 ships, **Then** that listing still lists current observational/ADR-evidence lore and is not the Accept/Edit/Reject workflow.

---

### User Story 2 - Accept promotes an extract with human verification provenance (Priority: P1)

A reviewer accepts a pending git or PR extract as stated. MemLore creates a **new** current lore entry whose origin is `human_verified`, whose verification status is `verified`, whose actor is the reviewer, and whose evidence is the same evidence the extract had (commit SHA and/or PR identity, plus any linked paths or issues). The observational predecessor is superseded, not overwritten or deleted. Compile continues to rank verified architecture decisions above leftover unverified observations; the newly accepted item outranks remaining unverified observations. Existing `verify` on the observational row is unchanged in meaning and is not what Accept does.

**Why this priority**: This is the constitution V promotion path. Capture without a real Accept would leave teams using verify-in-place as a false “this was always human.”

**Independent Test**: Seed an unverified git observation; accept it as stated; assert a new current `human_verified` + verified entry exists with the same statement and commit evidence; assert the observational row is superseded (still fetchable); assert `verify` on a different observation still only flips verification status and leaves origin `repository_observation`.

**Acceptance Scenarios**:

1. **Given** a pending observational extract “Payment events use transactional outbox” with PR evidence `acme/payments#1842`, **When** a writer with membership accepts it as stated, **Then** MemLore creates a new current lore entry with that statement, origin `human_verified`, verification `verified`, reviewer as actor, and the same PR evidence; the observational predecessor is superseded and still stored.
2. **Given** that accepted successor, **When** a caller inspects origin and verification, **Then** it is not `repository_observation` and was not created through human `remember` / `POST /v1/lore-entries` (those remain `human_authored` only).
3. **Given** an unverified observational row, **When** a caller uses existing `verify` instead of Accept, **Then** origin remains `repository_observation` and the row is not treated as F035 Accept.
4. **Given** a verified architecture-decision ADR plus remaining unverified observations plus the newly accepted `human_verified` entry, **When** context is compiled for a matching task, **Then** the accepted item outranks remaining unverified observations, and F032 accepted-ADR ranking versus observations is unchanged (ADR architecture still outranks leftover git/PR observations). Ranking formulas themselves are not modified.
5. **Given** Accept that creates lore, **When** the write completes, **Then** lore, audit, and knowledge-plane outbox for the mutation occur in one unit of work; the graph worker is not called from the review command.

---

### User Story 3 - Edit then Accept records the human-authored statement (Priority: P1)

A reviewer disagrees with the wording of an extract, supplies a corrected statement, and accepts. The trusted current lore uses the **human-authored** statement, not the raw extract labeled as if a human wrote it. Evidence from the extract is preserved on the successor. The observational predecessor is superseded and remains the historical extract.

**Why this priority**: Constitution V — do not convert an inference into a human-authored fact. Edit is how humans take responsibility for different wording.

**Independent Test**: Seed a PR observation with statement A; accept with edited statement B; assert current lore statement is B, origin `human_authored`, verified, evidence still includes the PR; assert predecessor statement remains A and is superseded.

**Acceptance Scenarios**:

1. **Given** a pending extract with statement A, **When** a reviewer accepts with a different statement B, **Then** the successor’s statement is B, origin is `human_authored`, verification is `verified`, and predecessor evidence is present on the successor.
2. **Given** that same accept, **When** the predecessor is fetched, **Then** its statement is still A, origin is still `repository_observation`, and it is superseded (not rewritten).
3. **Given** an accept whose supplied statement is the same as the extract after ordinary whitespace normalization, **When** it is stored, **Then** it is treated as Accept-as-stated (`human_verified`), not as a spurious human-authored rewrite.
4. **Given** `POST /v1/lore-entries` / `remember`, **When** F035 ships, **Then** those paths still require origin `human_authored` and MUST NOT be the Accept path (Accept must not drop extract evidence by creating a bare human entry).

---

### User Story 4 - Reject records a durable negative decision (Priority: P1)

A reviewer rejects a pending extract. The item leaves the pending queue. The observational lore remains as an observation (not deleted, not silently invalidated). The same extract identity cannot reappear as pending after re-ingest. Rejecting an already-rejected item is idempotent. Accepting a rejected extract is rejected as invalid.

**Why this priority**: Without a negative decision, operators would be re-prompted forever and might “accept” noise to clear the queue.

**Independent Test**: Reject a git extract; list pending (absent); re-run git ingest for that SHA; list pending (still absent); reject again (idempotent); attempt accept (validation error); observational row still exists as observational.

**Acceptance Scenarios**:

1. **Given** a pending extract, **When** a writer with membership rejects it, **Then** it is absent from the pending queue and a negative review decision is stored for its extract identity (scope + evidence type + evidence identity + statement checksum).
2. **Given** that rejected extract, **When** ingest runs again for the same repository and would consider the same SHA or PR, **Then** no new pending review item appears for that extract.
3. **Given** an already-rejected extract, **When** reject is requested again, **Then** the decision is unchanged (idempotent) and no duplicate pending item is created.
4. **Given** a rejected extract, **When** accept is requested, **Then** the system returns a validation error and does not create human-verified lore.
5. **Given** a rejected extract’s observational lore row, **When** reject completes, **Then** that row is not deleted and is not rewritten into human origin; history of the observation remains.

---

### User Story 5 - Operators use CLI and REST; agents do not mutate the queue (Priority: P2)

Reviewers list, accept, edit-accept, and reject through CLI and REST. Local mode requires an actor header or CLI `--actor`. Membership scopes every operation. MCP tool count remains 10; there is no new agent tool to accept or reject. There is no product web UI.

**Why this priority**: Constitution: CLI + REST are sufficient governance surfaces; web UI is F120. Agents must not auto-promote candidates.

**Independent Test**: CLI list/accept/reject on a fixture; REST list/accept/edit-accept/reject plus membership denial; MCP tool enumeration still 10; existing git/PR/ADR ingest routes still pass.

**Acceptance Scenarios**:

1. **Given** a writer with membership, **When** they run `memlore review list --repository github.com/acme/payments`, **Then** they see pending items for that repository in a human-readable summary (JSON not required).
2. **Given** a pending item id, **When** they run `memlore review accept <id>` or `memlore review accept <id> --statement "…"`, **Then** Accept or Edit-then-Accept occurs as in stories 2–3.
3. **Given** a pending item id, **When** they run `memlore review reject <id>`, **Then** Reject occurs as in story 4.
4. **Given** REST review-queue list (read) and accept/reject (write), **When** a principal lacks membership, **Then** the call is denied; repository keys are not placed in the URL path (item id in the path is allowed).
5. **Given** the MCP tool list, **When** tools are enumerated, **Then** the count remains 10 and there is no new tool that lists or mutates the review queue.
6. **Given** existing git, PR, and ADR ingest CLI/REST, **When** F035 ships, **Then** those surfaces still behave as specified in F030–F032.

---

### Edge Cases

- Unknown review item id: not found; no lore written.
- Review item whose lore is already superseded or invalidated: not pending; Accept/Reject returns a validation error (except idempotent re-reject of a stored rejection, and idempotent re-accept that returns the existing successor).
- Accept of an item that is not observational git/PR lore (including F032 ADR lore): not found or validation error; ADR trusted-source entries are never promoted through this queue.
- Concurrent Accept of the same pending item: one successor; the second request is idempotent (same successor) or a conflict — never two current human-verified rows for one extract.
- Blank actor on mutate: validation error (local mode still requires `X-Memlore-Actor` / `--actor`).
- Empty or whitespace-only edited statement: validation error.
- Edited statement longer than the lore statement limit: validation error; predecessor unchanged.
- Reader role: may list; must not Accept or Reject.
- Admin without membership: existing F114 admin bypass still applies (JWT admin bypasses membership), same as other writes.
- Graph-service down: lore, supersession, review decision, audit, and outbox still commit on the governance plane; graph catch-up remains the existing worker.
- `memlore worker` remains outbox → graph publisher only; review is not folded into the worker.
- Queue scoped only to membership-visible repository (and, if listed, other scopes the caller belongs to). Cross-tenant leak is a hard fail.
- Human-authored lore created via `remember` is not a review-queue item.
- Observational lore that was only `verify`-flipped in place (origin still `repository_observation`) MAY still appear as pending until Accept or Reject — verify is not Accept. If that proves noisy, Reject remains available; this slice does not auto-hide verified-but-still-observational rows from the queue.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Authorized members MUST be able to list a **pending suggested-lore review queue** for a repository they may read. Each pending item MUST include statement, evidence, and source type (git commit or pull request). Confidence and a separate reason MUST be omitted unless actually stored.
- **FR-002**: Pending items MUST be current, not-superseded, not-invalidated observational lore with origin `repository_observation` and git (`commit`) or PR (`pr`) evidence. F032 accepted-ADR lore (`architecture_decision` / evidence `adr` / trusted-source verified) MUST NOT appear.
- **FR-003**: The ingest candidates listing MUST remain an ingest listing. F035 MUST provide a dedicated review workflow (list + Accept + Edit-then-Accept + Reject) rather than overloading ingest candidates as the promotion UI.
- **FR-004**: Accept-as-stated MUST create a new current lore entry with origin `human_verified`, verification `verified`, reviewer as actor, statement equal to the extract, and **copied predecessor evidence**. The observational predecessor MUST be superseded (F110), not overwritten or deleted.
- **FR-005**: Edit-then-Accept (statement different from the extract) MUST create a new current lore entry with origin `human_authored`, verification `verified`, reviewer as actor, the **edited** statement, and copied predecessor evidence. The predecessor extract MUST remain stored and superseded.
- **FR-006**: Accept MUST NOT use human create (`NewLoreEntry` / `remember` / `POST /v1/lore-entries`) in a way that drops extract evidence. Human create remains `human_authored` only for direct remember; Accept-as-stated MUST use `human_verified` provenance.
- **FR-007**: Accept MUST NOT be implemented as in-place `verify`. Existing verify MUST keep flipping `verification_status` only, leaving origin unchanged.
- **FR-008**: Reject MUST store a negative review decision keyed by extract identity: repository scope + evidence type + evidence identity + statement checksum. The same extract MUST NOT reappear as pending after re-ingest. Observational lore MUST NOT be deleted or origin-rewritten by Reject.
- **FR-009**: Re-accept of an already-accepted extract MUST be idempotent (return the existing successor; no second current row). Re-reject of an already-rejected extract MUST be idempotent. Accept after Reject MUST fail validation. Reject after Accept MUST fail validation.
- **FR-010**: Nothing in the pending queue MUST be treated as canonical or high-authority until Accept. Compile ranking **formulas** MUST stay frozen (F007). Characterization MUST show: (a) accepted review-queue lore outranks remaining unverified observations; (b) F032 accepted-ADR lore still outranks leftover git/PR observations as before.
- **FR-011**: CLI MUST expose `memlore review list --repository <key>`, `memlore review accept <id> [--statement <text>]`, and `memlore review reject <id>`, with optional `--actor`. Output is human-readable (JSON not required). CLI uses the same local Postgres wiring as `memlore ingest git` (not an HTTP client to `serve`).
- **FR-012**: REST MUST provide membership-scoped `GET` pending queue (read) and `POST` accept / reject (write). Optional accept body carries a replacement statement. Repository keys MUST NOT be placed in the URL path. Exact paths are specified in the contract.
- **FR-013**: MCP tool count MUST remain 10. No new MCP tool for listing or mutating review items. Agents MUST NOT gain a default Accept/Reject tool.
- **FR-014**: Listing requires **read** plus F114 membership. Accept and Reject require **write** (writer or admin) plus membership — the same bar as ingest trigger and supersede, not the admin-only verify permission. Local mode: mutating routes require `X-Memlore-Actor`; membership off; actor is trusted admin.
- **FR-015**: Accept that creates or supersedes lore MUST write lore + audit + outbox (existing lore-mutation / `episode.ingest` pattern) in one unit of work. Review MUST NOT call the graph service. `memlore worker` stays outbox-only.
- **FR-016**: Review MUST be observable: structured logs for list/accept/reject with repository key, item id, actor, and outcome; review decisions MUST be auditable.
- **FR-017**: Future producers (documentation ingest, session extraction, agent correction) SHOULD be able to reuse this queue later. This slice MUST NOT implement F033, F034, F040, F050, or F100 producers, and MUST NOT change F032 skip of uncertain ADR files.
- **FR-018**: Product web UI is out of scope (F120). F022 packet profiles are out of scope.

### Key Entities

- **SuggestedLoreItem**: A pending review candidate projected from current observational lore (git or PR evidence): statement, evidence, source type, optional confidence, optional reason, lore entry id, repository scope.
- **ReviewDecision**: Durable overlay for one extract identity: pending (implicit — no row or open), accepted (successor lore id), or rejected. Keyed by scope + evidence type + evidence identity + statement checksum. Records actor and time.
- **AcceptedLore**: Current governed lore produced by Accept: `human_verified` (as-stated) or `human_authored` (edited), verified, evidence preserved, reviewer actor.
- **ObservationalPredecessor**: The git/PR lore row that remains after Accept as superseded history (origin unchanged).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fixture with one git observation, one PR observation, one F032 accepted ADR, and one rejected extract yields exactly two pending review items; the ADR and rejected extract are absent.
- **SC-002**: Accept-as-stated yields one new current `human_verified` + verified entry with the extract’s evidence; the observational predecessor is superseded, not deleted; origin of the predecessor remains `repository_observation`.
- **SC-003**: Edit-then-Accept yields current lore whose statement is the edited text and whose origin is `human_authored`; predecessor statement is unchanged.
- **SC-004**: After Reject, pending list does not include that extract; re-ingest of the same SHA/PR does not resurrect it; observational lore still exists.
- **SC-005**: A compile fixture with an F032 accepted ADR, an F035-accepted item, and leftover unverified observations ranks the accepted review item above those observations and still ranks the ADR-derived architecture above leftover observations; ranking formulas are unchanged.
- **SC-006**: REST contract tests cover list, accept, edit-accept, reject, and membership denial for a foreign repository; CLI prints a pending list and can accept/reject; git/PR/ADR ingest routes still pass their contract tests.
- **SC-007**: MCP tool count remains 10; `POST /v1/lore-entries` still creates human-authored lore only; `POST /v1/lore-entries/{id}/verify` still does not change origin.
- **SC-008**: Cross-tenant list or mutate is denied; a member of repository A never receives repository B’s pending items.
- **SC-009**: Pending items never display an invented confidence score or invented reason.
- **SC-010**: `go test ./...` and `go vet ./...` are green.

## Assumptions

- **Supersede over mutate-in-place**: Accept creates a successor and supersedes the observational row (constitution VI / F110). In-place origin rewrite is rejected because it would pretend the observation was always human-side.
- **human_verified constructor**: Accept-as-stated is the first writer of origin `human_verified`. Direct remember stays `human_authored`. Verify-in-place stays observational.
- **Evidence copy**: Successor always receives the predecessor’s evidence list (constitution: do not drop evidence on Accept). Generic F110 supersede-without-evidence remains for the existing supersede API; Accept is a dedicated promotion that copies evidence.
- **Queue projection**: Git/PR observational rows **are** the pending items. A review-decision overlay (not ingest cursor tables, not a second knowledge plane) stores Accept/Reject. Compile continues to read lore entries; rejected observations may still appear as low-authority lore until a later feature hides them.
- **No fake metadata**: Confidence and reason are optional and absent in v1 because F030/F031 do not store them.
- **Uncertain ADRs**: Remain skipped by F032. Not queued here.
- **Authz**: Writer+membership for mutate matches ingest/supersede. Verify remaining admin-only is an F010 contract this slice does not reopen.
- **MCP**: Agents already search and get lore. A dedicated review-list tool is not required for this slice; keeping 10 tools is the default.
- **Surfaces**: CLI `memlore review …` and REST `/v1/review-queue` (exact paths in contract). No web UI.
- **Idempotency of ingest**: F030/F031 SHA/PR processed tables already prevent duplicate lore; F035 rejection identity is additional insurance so a future extract with the same identity cannot re-enter pending.
- **Out of slice**: F033, F034, F040, F022, F050, F100, F120, enqueueing skipped ADR parses, changing compile formulas, auto-canonical git/PR ingest.
