# Feature Specification: ADR Auto-Ingestion (F032)

**Feature Branch**: `032-adr-ingest`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F032  
**Depends on**: F001 (lore model), F003 (authority), F008/F110 (supersede/invalidate DONE), F004/F107 (outbox DONE), F030 (git ingest producer pattern DONE), F031 (PR ingest producer pattern DONE), F114 (membership), F020 (profile DONE), F021 (get_for_task DONE)  
**Does not reopen**: F007 compile v1, F020 repository profile, F021 named packet, F030 git ingest, F031 PR ingest (additive fields/routes only if ADR ingest status must be stored)  
**Input**: Turn the team's existing Architecture Decision Records into governed lore without copy-paste. Discover ADRs from configured paths, extract decision knowledge, keep the original ADR as evidence (`evidence.type=adr`), and apply an explicit trusted-source policy for accepted ADRs.

## Goal

MemLore turns the team's existing Architecture Decision Records into governed lore without copy-paste. Operators point MemLore at a local working copy bound to an existing repository scope. MemLore discovers ADR files under default paths (`docs/adr/`, `adr/`, `architecture/decisions/`) plus optional extra paths, extracts decision, status, context, alternatives, consequences, supersession, and affected components when those are present in the file, and stores accepted ADRs as governed architecture-decision lore.

This slice **does** apply an explicit trusted-source policy for **accepted / adopted** ADRs (constitution V allows this). Trusted-source ingest MUST still preserve evidence and remain auditable. Git-derived and PR-derived observations (F030/F031) MUST remain distinguishable and MUST NOT be upgraded by this slice. Constitution VI: do not overwrite history; map ADR supersession onto lore supersession where possible; do not auto-resolve conflicts. Popularity/usage MUST NOT override authority.

This is capture of the team's existing decision corpus, not a generic document crawler (F033), not a suggested-lore review queue (F035), and not a first-class Decision aggregate (F040).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Trusted-source capture of accepted ADRs (Priority: P1)

An operator points MemLore at a local working copy bound to an existing repository scope they are authorized to write. MemLore scans default ADR directories plus any extra directories the operator named, skips templates, READMEs, and files that are not ADRs, and conservatively extracts decision knowledge from files that actually state a decision and a status. Each **accepted / adopted** ADR is stored as governed lore for that repository: origin `architecture_decision`, evidence type `adr` pointing at the original ADR identity (path and/or id), and **verified** status under the trusted-source policy. The triggering operator is the actor. Draft, proposed, and rejected ADRs are skipped. Uncertain parses (no Decision/Status headings, empty body) are skipped — extraction does not invent a decision the file does not state. Git and PR observational lore is not relabeled.

**Why this priority**: This is the product value of F032 — flywheel capture of the existing decision corpus as high-authority lore without copy-paste, without silently promoting git/PR observations, and without dumping every markdown file.

**Independent Test**: Ingest a fixture tree containing one accepted MADR/Nygard ADR, one draft ADR, one README, and one template; assert one verified `architecture_decision` lore entry with `adr` evidence from the accepted file; assert zero lore from draft, README, and template; assert origin is not `repository_observation` and not `human_authored`.

**Acceptance Scenarios**:

1. **Given** a local directory bound to repository scope `github.com/acme/payments` containing `docs/adr/0001-use-postgres.md` whose Status is Accepted and whose Decision heading states a choice, **When** ADR ingest runs for that repository, **Then** MemLore stores one lore entry whose statement is supported by that Decision text, whose origin is `architecture_decision`, whose verification status is `verified`, whose actor is the triggering operator, and whose evidence includes type `adr` identifying that ADR.
2. **Given** an ADR whose Status is Draft, Proposed, or Rejected, **When** ingest runs, **Then** that file produces no lore.
3. **Given** `docs/adr/README.md`, `docs/adr/template.md`, or `docs/adr/NNNN-title.md`, **When** ingest runs, **Then** those files are skipped and produce no lore.
4. **Given** a markdown file under an ADR directory with no Status and no Decision heading (or an empty body), **When** ingest runs, **Then** it is skipped; MemLore does not guess a decision.
5. **Given** a stored accepted-ADR lore entry, **When** a caller inspects origin, verification, and evidence, **Then** origin is `architecture_decision`, status is `verified`, evidence includes `adr`, and the entry is not `repository_observation` and was not created through human `CreateLore` / `remember`.
6. **Given** existing git-derived or PR-derived observational lore in the same repository, **When** ADR ingest runs, **Then** those observations are not upgraded, relabeled as architecture decisions, or auto-verified.

---

### User Story 2 - Idempotent re-ingest, content change, and safe retry (Priority: P1)

The same repository can be ingested again. An unchanged ADR (same repository + relative path + checksum) is a no-op. A changed accepted ADR (same ADR identity, different checksum) creates a **new** lore version and supersedes the prior **ingest-created** ADR lore for that identity — never a silent overwrite of the old row. Failed ingest retries without duplicating already-stored lore. Operators may name extra ADR directories beyond the defaults.

**Why this priority**: Flywheel capture is useless if retries create duplicate decisions or if editing an ADR silently rewrites history.

**Independent Test**: Ingest a fixture twice with unchanged files; assert lore count is unchanged. Change the accepted ADR body, ingest again; assert a new current lore entry exists, the previous ingest-created entry is superseded (not deleted), and both remain auditable. Simulate a failed run after one ADR is stored, retry, and assert that ADR still has exactly one current lore entry.

**Acceptance Scenarios**:

1. **Given** `docs/adr/0001-use-postgres.md` already produced verified lore, **When** ingest runs again and the file checksum is unchanged, **Then** no second lore entry is created and the existing entry is not modified.
2. **Given** that same ADR file's Decision text has changed, **When** ingest runs again, **Then** a new lore entry is created, the previous ingest-created ADR lore is superseded by it, and the previous row is not overwritten or deleted.
3. **Given** ingest failed after ADR `0001` was stored, **When** the operator retries, **Then** ADR `0001` is not duplicated as a second current entry and remaining unprocessed ADRs may still be considered.
4. **Given** a processed ADR was skipped as draft, **When** ingest runs again and the file is still draft, **Then** it is not re-extracted into lore.
5. **Given** an extra directory `architecture/records/` named on the trigger, **When** ingest runs, **Then** accepted ADRs there are considered in addition to the default path set.

---

### User Story 3 - ADR supersession maps to lore supersession (Priority: P2)

When an accepted ADR file states that it supersedes another ADR (for example “Supersedes ADR-0003”), and ingest-created lore already exists for that predecessor, MemLore applies existing lore supersession so the predecessor is no longer current. History is preserved. Human-authored lore is not auto-superseded in this slice, even if an ADR appears to name it, unless a later feature (F035/F040) handles that case. Deprecated or superseded ADR files are ingested so history exists, but they are not current canonical.

**Why this priority**: Constitution VI — do not overwrite history; map documented supersession onto governed supersession where the predecessor is ingest-created.

**Independent Test**: Ingest ADR-0003 (accepted) then ADR-0007 (accepted, “Supersedes ADR-0003”); assert 0007 is current verified architecture-decision lore and 0003's ingest-created lore is superseded, not deleted. Seed a human-authored lore entry and an ADR that names it; assert the human entry remains current.

**Acceptance Scenarios**:

1. **Given** ingest-created lore for ADR-0003 and a newly ingested accepted ADR-0007 whose file states it supersedes ADR-0003, **When** ingest stores 0007, **Then** 0003's ingest-created lore is superseded by 0007's lore and both rows remain in the store.
2. **Given** a human-authored lore entry (not ingest-created) and an ADR that appears to name it, **When** ADR ingest runs, **Then** the human-authored entry is not auto-superseded.
3. **Given** an ADR whose Status is Deprecated or Superseded, **When** ingest runs, **Then** lore is stored so the decision remains in history, and it is not treated as current canonical (superseded or invalidated when the file supports that; otherwise stored but not current verified canonical).
4. **Given** two accepted ADRs that conflict without a supersession link, **When** ingest stores both, **Then** neither is silently discarded or auto-resolved; both may be current and conflict remains visible.

---

### User Story 4 - Operators trigger ADR ingest and inspect status (Priority: P2)

An authorized writer or admin triggers ADR ingest for a repository they belong to, supplying the local working-copy path and optional extra ADR directories. They can read ADR ingest run status (started, succeeded, failed, counts of files seen / skipped / lore stored, error summary). Readers with membership can list ADR ingest runs and list ADR-derived lore for that repository. Unauthorized or cross-tenant callers are denied. Local-mode actor rules match F030/F031. Git and PR ingest routes and CLI remain working. Agents cannot create ADR-derived lore through a new MCP write tool. Human create remains human-only.

**Why this priority**: Constitution VIII: ingest is an observable operation. CLI + REST are the P0 surfaces; MCP stays at 10 tools.

**Independent Test**: CLI trigger + status on a fixture; REST trigger + list ADR runs + list ADR-derived candidates (`evidence_type=adr`); membership deny for a foreign repository; existing git and PR ingest still work; MCP tool count unchanged; `POST /v1/lore-entries` still human-authored.

**Acceptance Scenarios**:

1. **Given** a writer with membership on repository `github.com/acme/payments` and a valid local path, **When** they run `memlore ingest adr --repository github.com/acme/payments --path <dir>`, **Then** an ingest run is recorded and accepted ADRs (if any) are stored for that scope.
2. **Given** a completed or failed ADR run, **When** they run `memlore ingest status --repository github.com/acme/payments --kind adr` or GET ADR ingest status over REST, **Then** they see run state, file/lore counts, and an error summary when failed.
3. **Given** a reader with membership on that repository, **When** they list ADR ingest runs or ADR-derived lore (`evidence_type=adr`), **Then** the list is scoped to that repository and does not include other tenants’ runs or lore.
4. **Given** a principal without membership on the requested repository (OIDC membership mode), **When** they trigger ADR ingest or list ADR status, **Then** access is denied the same way as other writes/reads for that scope.
5. **Given** the MCP tool list, **When** tools are enumerated, **Then** the count remains 10 and there is no new write tool that creates ADR-derived lore.
6. **Given** existing git and PR ingest CLI/REST, **When** F032 ships, **Then** git and PR trigger, run list, and default `ingest status` (git) still behave as before.
7. **Given** `POST /v1/lore-entries`, **When** a caller creates lore, **Then** it remains human-authored only; ADR extracts are not writable through that route.

---

### User Story 5 - Trust boundary: accepted ADRs outrank git/PR observations (Priority: P2)

Accepted-ADR lore is visible as governed lore and compiles as high-authority architecture (F003 already treats `architecture_decision` origin and `adr` evidence as maximum source strength). Default compile / `get_for_task` ranking continues to put verified ADR/architecture above unverified repository observations. This slice does not change compile ranking formulas. Git and PR observations stay observational and unverified. Popularity/usage does not override that authority.

**Why this priority**: Constitution V and IX — accepted ADRs are the team's decision corpus; git/PR capture must not look equally canonical in the agent briefing.

**Independent Test**: Seed an ingested accepted ADR (verified `architecture_decision` + `adr` evidence) and an ingested git or PR observation in the same repository; compile for a matching task; assert the ADR-derived lore is ordered ahead of the observation. Characterization only — do not reopen F007/F020/F021 ranking formulas.

**Acceptance Scenarios**:

1. **Given** a verified ADR-derived architecture statement and an unverified git- or PR-derived observation in the same repository, **When** context is compiled for a matching task, **Then** the ADR-derived architecture is ordered ahead of the observation.
2. **Given** only git/PR observations and no ADR ingest, **When** compile runs, **Then** those observations retain `repository_observation` origin and unverified status (F032 has not upgraded them).
3. **Given** compile ranking tests from F030/F031, **When** F032 ships, **Then** verified architecture still outranks git/PR observations; this slice adds a characterization that ingested accepted ADR lore does likewise.

---

### Edge Cases

- Local path that is not a directory, not readable, or missing: ingest run fails with a clear error; no lore is written.
- None of the configured ADR directories exist: run succeeds with zero lore stored (empty discovery), unless the operator-supplied `--path` root itself is missing (then fail as above).
- Empty ADR directories: run succeeds with zero lore; processed set may be empty.
- ADR body longer than the lore statement limit: store a faithful extract of the Decision (title + decision section); skip the file if truncation would invent or drop the decision.
- Multiple Decision sections in one file: at most **one** lore entry per ADR file in v1 (the primary Decision). Conservative volume. Do not split context/alternatives/consequences into separate lore entries (that is F040).
- Concurrent ADR ingest runs for the same repository: at most one active ADR ingest run; a second trigger is rejected as conflict. Git or PR ingest for the same repository MAY run independently (separate operations).
- Graph-service down: lore, supersession, audit, and outbox still commit on the governance plane; graph catch-up remains the existing worker.
- Non-repository scopes: ingest is repository-scoped only; other scope kinds are rejected.
- Nested files under default dirs: ingest markdown files in the configured directories (non-recursive beyond one extra level of dated folders is allowed if present); skip unrelated trees.
- Front matter `status: accepted` without a Status heading: treat as a stated status (common MADR); do not require both.
- Status synonyms (Accepted / Adopted / Approved): trusted. Unknown status tokens: skip (do not guess).
- `memlore worker` remains outbox → graph publisher only; ADR scanning is not folded into the worker.
- REST create-lore remains human-authored only.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Operators MUST be able to ingest ADRs from a **local working copy** bound to an existing repository scope key. GitHub Contents API, clone-from-URL, and remote fetch are out of scope. The operator supplies the local path (F030-style).
- **FR-002**: Discovery MUST scan the default path set `docs/adr/`, `adr/`, and `architecture/decisions/` relative to the supplied working copy, plus optional extra directories named on the trigger. This is not a generic documentation crawler (F033).
- **FR-003**: Discovery MUST skip README files, ADR templates (`template.md`, `NNNN-title.md` and equivalent placeholders), and markdown that is not an ADR (no usable Status/Decision). Conservative beats volume.
- **FR-004**: Extraction MUST capture, when present in the file: title/decision, status, context, alternatives, consequences, supersession links, and affected components. Extraction MUST NOT invent a decision the file does not state. Uncertain parses MUST be skipped (F035 is later).
- **FR-005**: At most **one** lore entry MAY be created per ADR identity (repository + relative path / ADR id) in this slice. Context, alternatives, and consequences belong in that entry's statement (faithful extract), not as split lore rows.
- **FR-006**: **Trusted-source policy (accepted ADRs)**: files whose status is Accepted, Adopted, or Approved MUST be persisted as governed lore with origin `architecture_decision`, evidence type `adr`, and verification status `verified`. Created-by is the triggering operator. This MUST remain auditable (create audit + evidence). This policy MUST NOT apply to git/PR observational lore or to arbitrary markdown (F033).
- **FR-007**: Files whose status is Draft, Proposed, or Rejected MUST be skipped (no lore). Files whose status is Deprecated or Superseded MUST be ingested so history is preserved, but MUST NOT be current canonical (apply supersession or invalidation when the file states it; otherwise store without treating them as current verified canonical).
- **FR-008**: Each stored item MUST include evidence of type `adr` whose value identifies the original ADR (id/path slug, e.g. `0001-dual-plane` or the relative path). Optional additional `path` evidence for the file path is allowed. Existing evidence types remain valid.
- **FR-009**: Human create MUST remain human-only: `CreateLore` / MCP `remember` / `POST /v1/lore-entries` require `human_authored`. ADR ingest MUST use a dedicated trusted-source constructor (origin `architecture_decision`, evidence `adr`, verified). `NewLoreEntry` MUST NOT be loosened to accept any origin. `NewObservationalLoreEntry` MUST NOT accept ADR evidence as a substitute for this constructor.
- **FR-010**: Re-ingest of the same ADR (same repository + relative path + unchanged checksum) MUST be a no-op. A changed accepted ADR (same identity, different checksum) MUST create a new lore version and supersede prior **ingest-created** ADR lore for that identity. History MUST NOT be silently overwritten.
- **FR-011**: The system MUST persist additive operational records for ADR ingest runs, per-repository cursor/progress, and processed ADR identity (relative path + checksum) so re-runs are incremental and retries do not duplicate lore. These tables MUST NOT overload git SHA or PR number tables.
- **FR-012**: Creating ADR-derived lore MUST follow the same governance write unit as other lore creates: lore + audit + outbox episode ingest in one unit of work (plus supersede audit when chaining). The existing graph worker publishes outbox only.
- **FR-013**: When an ADR file states that it supersedes another ADR and ingest-created lore exists for that predecessor, the system MUST map that onto existing lore supersession. Human-authored predecessors MUST NOT be auto-superseded in v1. Conflicts without a supersession link MUST NOT be auto-resolved.
- **FR-014**: CLI MUST expose `memlore ingest adr --repository <key> --path <dir>` with optional extra ADR directories (`--adr-dir`). CLI MUST extend `memlore ingest status --repository <key>` with `--kind git|pr|adr` (default remains `git`). `--kind adr` prints a human-readable latest ADR-run summary (JSON not required).
- **FR-015**: REST MUST provide membership-scoped trigger and status: `POST /v1/ingest/adr` (write), list/get ADR ingest runs (e.g. `/v1/ingest/adr-runs`, read), and list ADR-derived items via the existing candidates endpoint with additive `evidence_type=adr`. Existing git and PR ingest routes MUST keep working. Repository keys MUST NOT be placed in the URL path. Trigger MUST NOT silently mark non-accepted files as canonical.
- **FR-016**: MCP tool count MUST remain 10. No new MCP tool. No MCP write that creates or promotes ADR extracts. `memlore.remember` MUST NOT gain ADR-extract semantics.
- **FR-017**: Triggering ingest requires **write** permission (writer or admin) plus F114 membership on the repository scope. Listing runs and ADR-derived lore requires **read** plus membership. Local mode: mutating routes require `X-Memlore-Actor`; membership off; actor is trusted admin (same as F030/F031). CLI uses the same local Postgres wiring as `memlore profile` / `memlore ingest git` (not an HTTP client to `serve`).
- **FR-018**: Compile / `get_for_task` default ranking MUST continue to place verified architecture / ADR-sourced lore above unverified repository observations. This slice MUST NOT change compile ranking formulas, packet section ids, or F007/F020/F021/F030/F031 contracts. A characterization test MUST assert ingested accepted ADR lore outranks git/PR observations.
- **FR-019**: Ingest MUST be observable: structured logs for start/complete/fail with repository key, run id, counts, and error; ingest run records expose status to operators.
- **FR-020**: Concurrent ADR ingest for the same repository MUST NOT double-write lore (one active ADR ingest run per repository).
- **FR-021**: F033 documentation ingest, F035 accept/reject queue, F040 first-class Decision aggregate, and web UI (F120) are out of scope. Do not implement a generic markdown crawler.

### Key Entities

- **ADRSnapshot**: An ADR file read from the working copy: relative path, checksum, title, status, decision, context, alternatives, consequences, supersession references, affected components (when present).
- **ADRDerivedLore**: Governed lore derived from one ADR: statement (faithful decision extract), origin `architecture_decision`, evidence including `adr`, scoped to the repository. Accepted/adopted items are verified (trusted-source). Deprecated/superseded items are historical, not current canonical.
- **ADRIngestRun**: An observable operation for one repository: actor, local path, extra directories, state (running / succeeded / failed), counts (files seen, skipped, lore stored, superseded), error summary.
- **ADRIngestCursor**: Per-repository watermark of ingest progress used for incremental re-runs.
- **ProcessedADR**: Per-repository record that a relative path + checksum was considered (lore stored or skipped), used for idempotency.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fixture with one accepted ADR, one draft ADR, one README, and one template yields exactly one stored current lore entry after ingest, with origin `architecture_decision`, verification `verified`, and evidence type `adr`.
- **SC-002**: Ingesting the same fixture twice with unchanged files yields the same current lore count (no duplicates) and the same `adr` evidence on that entry.
- **SC-003**: Changing the accepted ADR and re-ingesting yields a new current lore entry; the previous ingest-created entry is superseded, not deleted or overwritten.
- **SC-004**: A compile fixture with that ingested accepted ADR plus a git or PR observation ranks the ADR-derived lore ahead of the observation; git/PR items remain `repository_observation` and unverified.
- **SC-005**: REST contract tests cover ADR trigger, ADR run status, ADR-derived candidate list (`evidence_type=adr`), and membership denial for a foreign repository; CLI prints an ADR status summary for a completed run; git and PR ingest routes still pass their contract tests.
- **SC-006**: MCP tool count remains 10; `POST /v1/lore-entries` still creates human-authored lore only.
- **SC-007**: A failed-then-retried ingest of a fixture does not duplicate the lore that was stored before the failure.
- **SC-008**: An accepted ADR that states it supersedes a previously ingested ADR leaves the predecessor ingest-created lore superseded and the successor current; a human-authored entry is not auto-superseded.
- **SC-009**: Missing local path yields a failed observable run with no lore written.
- **SC-010**: `go test ./...` and `go vet ./...` are green.

## Assumptions

- **Trusted-source policy (v1 public contract)**: Accepted / Adopted / Approved ADRs persist as `architecture_decision` + `adr` evidence + **verified**. This is the constitution V exception called out in the tracker (“Accepted ADRs receive high source authority”). Auto-verify is limited to this policy; it is not a general ingest promotion. Actor = triggering operator; create is audited.
- **Skip vs historical**: Draft / Proposed / Rejected → skip. Deprecated / Superseded → ingest as historical (not current canonical). Unknown or missing status with no Decision → skip.
- **Status tokens (v1, documented for testers)**: trusted = `accepted`, `adopted`, `approved` (case-insensitive). Skip = `draft`, `proposed`, `rejected`, `withdrawn`. Historical = `deprecated`, `superseded`, `superceded`. Other tokens → skip.
- **One lore entry per ADR file**: do not split Decision / Context / Consequences into multiple lore rows (F040). Statement is a faithful extract: title plus Decision, with Context / Alternatives / Consequences appended when they fit the statement limit; skip if the Decision itself cannot be represented faithfully.
- **Source (v1)**: Local filesystem working copy + `t.TempDir()` fixtures. No GitHub Contents API. Operator supplies `--path` as the repo root; ADR dirs are relative to that root.
- **Idempotency key**: repository scope + relative path + content checksum. Unchanged checksum → no-op (even if extractor rules later change — v1 does not rewrite). Changed checksum + same ADR identity + prior ingest-created lore → new version + supersede.
- **ADR identity for evidence**: prefer filename slug without extension (e.g. `0001-use-postgres`); if the file states an ADR id (ADR-0001 / 0001), use that slug consistently so supersession can match.
- **Supersession matching**: only chain ingest-created ADR lore by ADR id/slug. Do not auto-supersede `human_authored` / `human_verified` lore in v1.
- **Parser**: Conservative heading/front-matter parse of common MADR and Nygard templates (Status, Context, Decision, Consequences, Alternatives, Supersedes). Stdlib only; no ADR SDK. Do not invent sections that are absent.
- **Operational tables**: Additive `adr_ingest_runs`, `adr_ingest_cursors`, `adr_ingest_files` (path + checksum). Do not store paths in `git_ingest_shas` or `pr_ingest_prs`.
- **Constructor**: New sibling of `NewObservationalLoreEntry` (e.g. `NewArchitectureDecisionLoreEntry`) requiring origin `architecture_decision`, at least one `adr` evidence, and verified status with VerifiedBy/VerifiedAt = triggering actor / now. Do not loosen `NewLoreEntry`.
- **Operator interface**: Dedicated CLI `memlore ingest adr` rather than folding ADR scanning into `memlore worker`. REST trigger runs ingest in-process for v1. Status extends `--kind adr`; default remains git.
- **Compile**: No F007/F020/F021/F030/F031 ranking formula changes. F003 already maps `architecture_decision` origin or `adr` evidence to ADR source type and max evidence strength. Add a characterization test that ingested accepted ADR lore outranks git/PR observations.
- **MCP**: No 11th tool; status is CLI + REST.
- **Web UI**: Out of scope (F120).
- **Out of slice**: F033 docs ingest, F035 review queue (uncertain extracts are skipped, not queued), F040 Decision aggregate.
- Historical IDs F004/F107 (outbox) and F032's unrelated historical numbering are not reused; this is product F032.
