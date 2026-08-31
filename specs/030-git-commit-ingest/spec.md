# Feature Specification: Git Commit Ingestion (F030)

**Feature Branch**: `030-git-commit-ingest`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F030  
**Depends on**: F004/F107 (outbox DONE), F005/F106 (graph ingest DONE), F003 (authority), F114 (membership), F020 (profile DONE), F021 (get_for_task DONE)  
**Does not reopen**: F007 compile v1, F020 repository profile, F021 named packet (no compile ranking changes in this slice)  
**Input**: Recover engineering *why* from git history that never became an ADR by ingesting commits from a configured repository as candidate / observational knowledge with evidence links to SHAs.

## Goal

MemLore recovers rationale, migration context, bug explanations, component relationships, and technical constraints from git commit history that never became an ADR. Commits from a configured repository are ingested as **candidate / observational** knowledge, each with an evidence link to the commit SHA.

Git-derived observations MUST remain distinguishable from human-authored or human-verified knowledge. They MUST NOT silently become canonical. They MUST NOT auto-verify. Compile / `get_for_task` MUST continue to rank verified architecture above git observations. Failed ingest retries MUST NOT duplicate already-stored candidates. Re-ingest of the same SHA is idempotent.

This is capture, not promotion. F035 (suggested-lore review queue), F031 (PR ingest), and F032 (ADR ingest) are out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conservative commit capture as observational lore (Priority: P1)

An operator points MemLore at a local git working copy that is bound to an existing repository scope they are authorized to write. MemLore reads commits (author, SHA, timestamp, message, changed paths), skips noisy history, and extracts a small number of candidate engineering statements that are actually supported by the commit text and paths. Each accepted candidate is stored as governed lore for that repository: observational origin, unverified, with evidence pointing at the commit SHA. Nothing is marked human-authored, human-verified, or canonical. Extraction does not invent a decision the commit does not support.

**Why this priority**: This is the product value of F030 — recover *why* from history without polluting the authority plane.

**Independent Test**: Ingest a fixture repository containing one rationale commit, one chore/version bump, and one merge commit; assert one observational unverified candidate with SHA evidence from the rationale commit; assert zero candidates from the noisy SHAs; assert origin is not human-authored and verification is not verified.

**Acceptance Scenarios**:

1. **Given** a local git directory bound to repository scope `github.com/acme/payments` containing a commit whose message explains why a migration was required and lists changed paths, **When** ingest runs for that repository, **Then** MemLore stores a candidate whose statement is supported by that commit text, whose origin is observational (`repository_observation`), whose verification status is unverified, and whose evidence includes the commit SHA.
2. **Given** commits that are merge-only, empty-message, or noisy chore/version/CI bumps (documented skip rules), **When** ingest runs, **Then** those SHAs produce no candidate lore.
3. **Given** a commit whose message is a terse conventional subject with no rationale in the body, **When** ingest runs, **Then** no candidate is invented from paths or heuristics that the text does not support.
4. **Given** a stored git-derived candidate, **When** a caller inspects origin and verification, **Then** it is not `human_authored`, not `human_verified`, not verified, and not treated as an architecture decision by ingest itself.

---

### User Story 2 - Idempotent re-ingest and safe retry (Priority: P1)

The same repository can be ingested again. Already-processed SHAs are not re-extracted into duplicate lore, whether the SHA previously produced a candidate or was skipped as noisy. If a run fails after some SHAs were accepted, retry continues from a per-repository watermark and does not duplicate accepted candidates. Incremental runs only consider commits newer than the stored cursor (or unprocessed SHAs) so operators can re-run safely.

**Why this priority**: Flywheel capture is useless if retries create duplicate “why” statements or if a failed batch cannot be resumed.

**Independent Test**: Ingest a fixture twice; assert candidate count is unchanged and the same SHA maps to the same lore. Simulate a failed run after one SHA is stored, retry, and assert that SHA still has exactly one candidate.

**Acceptance Scenarios**:

1. **Given** SHA `abc` already produced a candidate, **When** ingest runs again for the same repository, **Then** no second lore entry is created for that SHA.
2. **Given** SHA `def` was processed and skipped as noisy, **When** ingest runs again, **Then** it is not re-extracted and still produces no lore.
3. **Given** ingest failed after SHA `abc` was stored, **When** the operator retries, **Then** SHA `abc` is not duplicated and remaining unprocessed SHAs may still be considered.
4. **Given** a repository whose cursor already covers commits through SHA `old`, **When** a new rationale commit `new` is added and ingest runs, **Then** only `new` (and any other unprocessed SHAs) can produce new candidates.

---

### User Story 3 - Operators trigger ingest and inspect status (Priority: P2)

An authorized writer or admin triggers git ingest for a repository they belong to, supplying the local clone path. They can read ingest run status (started, succeeded, failed, counts of commits seen / skipped / candidates stored, cursor position, error summary). Readers with membership can list runs and list git-derived candidates for that repository. Unauthorized or cross-tenant callers are denied. Local-mode actor rules match other writes (`X-Memlore-Actor` required). Agents cannot promote extracts to canonical through a new MCP write tool.

**Why this priority**: Constitution VIII: ingest is an observable operation. CLI + REST are the P0 surfaces; MCP stays at 10 tools.

**Independent Test**: CLI trigger + status on a fixture repo; REST trigger + list runs + list candidates; membership deny for a foreign repository; MCP tool count unchanged.

**Acceptance Scenarios**:

1. **Given** a writer with membership on repository `github.com/acme/payments` and a valid local clone path, **When** they run `memlore ingest git --repository github.com/acme/payments --path <clone>`, **Then** an ingest run is recorded and candidates (if any) are stored for that scope.
2. **Given** a completed or failed run, **When** they run `memlore ingest status --repository github.com/acme/payments` or GET ingest status over REST, **Then** they see run state, commit/candidate counts, cursor, and an error summary when failed.
3. **Given** a reader with membership on that repository, **When** they list ingest runs or git-derived candidates, **Then** the list is scoped to that repository and does not include other tenants’ runs or lore.
4. **Given** a principal without membership on the requested repository (OIDC membership mode), **When** they trigger ingest or list status, **Then** access is denied the same way as other writes/reads for that scope.
5. **Given** the MCP tool list, **When** tools are enumerated, **Then** the count remains 10 and there is no new write tool that creates git-derived lore.

---

### User Story 4 - Trust boundary: compile still prefers verified architecture (Priority: P2)

Git-derived candidates are visible as governed lore (so they are not a hidden second store), but default compile / `get_for_task` ranking continues to put verified human/ADR architecture above unverified repository observations. Ingest does not auto-verify, does not apply trusted-source policy, and does not change compile contracts except that observational lore may appear among lower-ranked items when it is current and in budget. F035 accept/reject is not built; candidates remain unverified observational lore until a later feature promotes them.

**Why this priority**: Constitution V is non-negotiable. Capture must not look like canonical architecture in the agent briefing.

**Independent Test**: Seed a verified architecture statement and an ingested git observation in the same repository; compile for a matching task; assert the verified architecture is ordered ahead of the git observation; assert the git item’s origin is `repository_observation` and status is unverified.

**Acceptance Scenarios**:

1. **Given** a verified human-authored architecture statement and an unverified git-derived observation in the same repository, **When** context is compiled for a matching task, **Then** the verified architecture is ordered ahead of the git observation.
2. **Given** only git-derived unverified observations, **When** they appear in compile, **Then** they retain observational origin and unverified status (they are not relabeled as architecture decisions by ingest).
3. **Given** ingest has stored candidates, **When** a caller uses the existing verify API, **Then** ingest itself has not already marked them verified; promotion remains a separate human action (existing verify, or F035 later).

---

### Edge Cases

- Local path that is not a git directory, not readable, or not a directory: ingest run fails with a clear error; no lore is written.
- Empty repository (no commits): run succeeds with zero candidates; cursor remains unset or records “nothing processed”.
- Commit message longer than the lore statement limit: candidate is skipped or truncated only if the truncated text remains a faithful extract of the message (do not paraphrase into a new claim); prefer skip when truncation would drop the rationale.
- SHA already stored from a previous run with different extraction rules: v1 treats SHA as processed; do not rewrite the existing candidate (constitution VI: do not overwrite history).
- Detached HEAD / shallow clone: ingest the reachable history from HEAD; do not fetch remotes.
- Binary-only or generated-file path changes with no rationale message: skip (noisy / unsupported).
- Multiple qualifying paragraphs in one commit: at most one candidate per SHA in v1 (the primary message extract). Conservative volume.
- Concurrent ingest runs for the same repository: at most one active run; a second trigger is rejected or queued as failed-conflict, not duplicated writers.
- Graph-service down: lore and outbox still commit on the governance plane (same as other creates); graph catch-up remains the existing worker.
- Non-repository scopes: ingest is repository-scoped only; other scope kinds are rejected.
- REST create-lore remains human-authored only; git extracts MUST NOT be writable via `POST /v1/lore-entries` as if they were human ADRs.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Operators MUST be able to ingest commits from a **local git directory** bound to an existing repository scope key. GitHub API / remote forge adapters are out of scope (Epic H / F031).
- **FR-002**: For each processed commit, the system MUST capture author, full SHA, timestamp, message, and changed-path metadata (paths may be stored on the ingest record and/or as `path` evidence; SHA evidence is mandatory on every candidate).
- **FR-003**: Extraction MUST produce **candidate** engineering knowledge, not a dump of every commit message. Skip rules (v1): merge commits; empty or whitespace-only messages; conventional subjects that are only `chore`, `ci`, `style`, `build`, `bump`, or version-bump with no body rationale. Keep commits whose subject or body contains supported rationale cues (e.g. because / so that / workaround / migration / breaking / to fix / instead of / avoid / constraint) **and** whose extract is a faithful use of the commit text (no invented decisions).
- **FR-004**: At most **one** candidate lore entry MAY be created per SHA in this slice. The statement MUST be derived from the commit message text (subject plus body), not from model inference or path-only guesses.
- **FR-005**: Each stored candidate MUST have origin `repository_observation`, verification status `unverified`, and MUST NOT set `human_verified` origin or verified status. Ingest MUST NOT use trusted-source policy (that is F032 for accepted ADRs).
- **FR-006**: Each stored candidate MUST include evidence of type `commit` whose value is the full commit SHA. Optional additional `path` evidence MAY list changed paths. Existing types `url`, `path`, `adr` remain valid; `commit` is additive.
- **FR-007**: Re-ingest of the same SHA for the same repository scope MUST be idempotent: one candidate (or zero, if skipped), never a duplicate. Processed SHAs include skipped noisy SHAs.
- **FR-008**: The system MUST persist a per-repository ingest cursor (watermark: last processed commit timestamp and SHA) so re-runs are incremental. Failed runs MUST be retryable without duplicating already-stored candidates.
- **FR-009**: Candidates MUST be persisted as governed **lore entries** in the existing knowledge store (not a parallel `suggested_lore` knowledge table). A small operational ingest-run / processed-SHA index is allowed for cursor, idempotency, and status; it is not a second knowledge plane. Compile continues to read lore entries.
- **FR-010**: Creating a git-derived lore entry MUST follow the same governance write unit as other lore creates: lore + audit + outbox episode ingest in one unit of work, so the existing graph worker can publish. Human `CreateLore` / MCP `remember` / REST create remain human-authored only.
- **FR-011**: CLI MUST expose `memlore ingest git --repository <key> --path <dir>` with optional `--max-commits` (positive integer cap for a run). CLI MUST expose `memlore ingest status --repository <key>` printing a human-readable latest-run summary (JSON not required).
- **FR-012**: REST MUST provide membership-scoped trigger and status: trigger ingest (write), list/get ingest runs (read), and list git-derived candidates for a repository (read). Trigger MUST NOT silently mark lore canonical. Exact paths are specified in the contract; existing `POST /v1/lore-entries` MUST remain human-authored.
- **FR-013**: MCP tool count MUST remain 10. No new MCP tool. No MCP write that creates or promotes git extracts.
- **FR-014**: Triggering ingest requires **write** permission (writer or admin) plus F114 membership on the repository scope. Listing runs and candidates requires **read** plus membership. Local mode: mutating routes require `X-Memlore-Actor`; membership off; actor is trusted admin (same as other writes). CLI uses the same local Postgres wiring as `memlore profile` / `memlore context` (not an HTTP client to `serve`).
- **FR-015**: Compile / `get_for_task` default ranking MUST continue to place verified architecture above unverified repository observations (existing authority evaluation). This slice MUST NOT change compile ranking formulas, packet section ids, or F007/F021 contracts.
- **FR-016**: Ingest MUST be observable: structured logs for start/complete/fail with repository key, run id, counts, and error; ingest run records expose status to operators.
- **FR-017**: Concurrent ingest for the same repository MUST NOT double-write candidates (one active run per repository).
- **FR-018**: F031 PR ingest, F032 ADR ingest, F035 accept/reject queue, web UI, and GitHub-first forge adapters are out of scope.

### Key Entities

- **GitCommitSnapshot**: A commit read from a local repository: author, full SHA, timestamp, message, changed paths.
- **CommitCandidate**: Observational lore derived from one SHA: statement, origin `repository_observation`, unverified, evidence including `commit` SHA, scoped to the repository.
- **IngestRun**: An observable operation for one repository: actor, local path, state (running / succeeded / failed), counts (commits seen, skipped, candidates stored), cursor after the run, error summary.
- **IngestCursor**: Per-repository watermark of the last successfully processed commit (SHA + timestamp) used for incremental re-runs.
- **ProcessedSHA**: Per-repository record that a SHA was considered (candidate stored or skipped), used for idempotency.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fixture repository with one rationale commit and two noisy commits (merge + chore bump) yields exactly one stored candidate after ingest.
- **SC-002**: Ingesting the same fixture twice yields the same candidate count (no duplicates) and the same SHA evidence on that candidate.
- **SC-003**: Every stored git-derived candidate has origin `repository_observation`, verification `unverified`, and evidence type `commit` pointing at the SHA.
- **SC-004**: A compile fixture with verified architecture plus that git observation ranks the verified architecture ahead of the observation.
- **SC-005**: REST contract tests cover trigger, run status, candidate list, and membership denial for a foreign repository; CLI prints a status summary for a completed run.
- **SC-006**: MCP tool count remains 10; `POST /v1/lore-entries` still creates human-authored lore only.
- **SC-007**: A failed-then-retried ingest of a fixture does not duplicate the candidate that was stored before the failure.
- **SC-008**: `go test ./...` and `go vet ./...` are green.

## Assumptions

- **Knowledge store**: Persist candidates as `lore_entries` with observational origin + unverified status so F003 authority already downranks them vs verified human/ADR. F035 is not required to *hold* candidates. A parallel suggested-lore knowledge table is rejected because compile could not see it without extra work.
- **Commit source (v1)**: Local git directory bound to an existing repository scope key. No GitHub API, no clone-from-URL, no fetch. Operator supplies `--path`. This ships faster and is testable without forge credentials.
- **Evidence**: New type `commit` with value = full SHA. Optional `path` evidence for changed files. Do not overload `url` to mean SHA.
- **Extractor**: Deterministic heuristic on message text + skip rules; not an LLM. Conservative: at most one candidate per SHA; skip when unsupported.
- **Operator interface**: Dedicated CLI `memlore ingest git` / `memlore ingest status` rather than folding git polling into `memlore worker`. `memlore worker` remains the outbox → graph publisher. REST trigger runs ingest in-process for v1 (local git, optional `--max-commits`) and records the run.
- **Cursor**: Processed-SHA set plus last SHA/timestamp; incremental by committer/author timestamp then SHA order as recorded by git log from HEAD.
- **Created-by**: The triggering actor (CLI/REST actor id). Origin still `repository_observation`.
- **Compile**: No F007/F020/F021 contract changes. Git observations may appear in `items` or task_context when they rank into budget; they must not outrank verified architecture via ingest-side flags.
- **MCP**: No 11th tool; status is CLI + REST.
- **Web UI**: Out of scope (F120).
- **Shared ingest primitive**: Only what F030 needs (run + cursor + processed SHA + observational lore write). F031/F032/F035 wait.
- **Skip-rule details** (v1, documented for testers):
  - Merge: second parent exists, or subject matches `Merge ` / `Merge pull request`.
  - Noisy prefix: subject matches conventional-commit type `chore|ci|style|build` with no body, or subject is only a version bump (`v1.2.3`, `bump version`, `release x.y`).
  - Rationale cues (case-insensitive, message subject or body): `because`, `so that`, `workaround`, `migration`, `breaking`, `to fix`, `instead of`, `avoid`, `constraint`, `why:`.
  - If skip and cue both match, skip wins (noisy).
- Historical IDs F004/F107 (outbox) are not reused; this is product F030.
