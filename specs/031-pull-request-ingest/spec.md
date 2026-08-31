# Feature Specification: Pull Request Ingestion (F031)

**Feature Branch**: `031-pull-request-ingest`  
**Created**: 2026-09-01  
**Status**: Implemented  
**Product ID**: F031  
**Depends on**: F030 (git ingest producer pattern DONE), F004/F107 (outbox DONE), F005/F106 (graph ingest DONE), F003 (authority), F114 (membership), F020 (profile DONE), F021 (get_for_task DONE)  
**Does not reopen**: F007 compile v1, F020 repository profile, F021 named packet, F030 git ingest routes (additive fields/routes only)  
**Input**: Recover engineering *why* from pull-request discussion that never became an ADR by ingesting merged GitHub PRs as candidate / observational knowledge with evidence links to the original PR (and to review comments actually used).

## Goal

MemLore recovers rationale, trade-offs, linked-issue context, and review decisions from pull-request discussion that never became an ADR. PRs are often the real decision record. Merged PRs from a configured GitHub repository are ingested as **candidate / observational** knowledge, each with an evidence link to the original PR.

PR-derived observations MUST remain distinguishable from human-authored or human-verified knowledge. They MUST NOT silently become canonical. They MUST NOT auto-verify. Compile / `get_for_task` MUST continue to rank verified architecture above PR observations. Failed ingest retries MUST NOT duplicate already-stored candidates. Re-ingest of the same PR (same repository + number) is idempotent.

This is capture, not promotion. F032 (ADR ingest), F035 (suggested-lore review queue), F054 (GitHub PR check), and F074 (review bot) are out of scope.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conservative PR capture as observational lore (Priority: P1)

An operator points MemLore at an existing repository scope that maps to a GitHub repository they are authorized to write. MemLore reads merged pull requests (title, description, linked issues, changed files, review discussion used as evidence, merged state, author, timestamps), skips noisy or unmerged PRs, and extracts a small number of candidate engineering statements that are actually supported by the PR text and review comments. Each accepted candidate is stored as governed lore for that repository: observational origin, unverified, with evidence pointing at the original PR. Review comments that actually contributed to the extract are also linked. Linked issues are stored as evidence when present. Nothing is marked human-authored, human-verified, or canonical. Extraction does not invent a decision the PR does not support.

**Why this priority**: This is the product value of F031 — recover *why* from PR discussion without polluting the authority plane.

**Independent Test**: Ingest a fixture of merged PRs containing one rationale PR (title/body or review explains a decision), one dependabot/chore bump with no review rationale, and one unmerged PR; assert one observational unverified candidate with PR evidence from the rationale PR; assert zero candidates from the noisy and unmerged PRs; assert origin is not human-authored and verification is not verified.

**Acceptance Scenarios**:

1. **Given** repository scope `github.com/acme/payments` mapped to GitHub `acme/payments` containing a merged PR whose title or description explains why a change was required, **When** ingest runs for that repository, **Then** MemLore stores a candidate whose statement is supported by that PR text, whose origin is observational (`repository_observation`), whose verification status is unverified, and whose evidence includes a link to that PR.
2. **Given** a merged PR whose title/body is a noisy dependabot or chore bump with no human review rationale, **When** ingest runs, **Then** that PR produces no candidate lore.
3. **Given** an open, draft, or closed-unmerged PR, **When** ingest runs, **Then** it is skipped and is not stored as landed implementation.
4. **Given** a merged PR whose extract used a specific review comment, **When** the candidate is stored, **Then** evidence includes both the PR link and a link to that review comment.
5. **Given** a merged PR that names a linked issue or ticket, **When** a candidate is stored, **Then** that issue/ticket is stored as evidence.
6. **Given** a stored PR-derived candidate, **When** a caller inspects origin and verification, **Then** it is not `human_authored`, not `human_verified`, not verified, and not treated as an architecture decision by ingest itself.

---

### User Story 2 - Idempotent re-ingest and safe retry (Priority: P1)

The same repository can be ingested again. Already-processed PRs are not re-extracted into duplicate lore, whether the PR previously produced a candidate or was skipped as noisy or unmerged. If a run fails after some PRs were accepted, retry continues from a per-repository watermark and does not duplicate accepted candidates. Incremental runs only consider merged PRs newer than the stored cursor (or unprocessed PRs). Operators may ingest a single PR by number.

**Why this priority**: Flywheel capture is useless if retries create duplicate “why” statements or if a failed batch cannot be resumed.

**Independent Test**: Ingest a fixture twice; assert candidate count is unchanged and the same PR maps to the same lore. Simulate a failed run after one PR is stored, retry, and assert that PR still has exactly one candidate.

**Acceptance Scenarios**:

1. **Given** PR `#1842` already produced a candidate, **When** ingest runs again for the same repository, **Then** no second lore entry is created for that PR.
2. **Given** PR `#99` was processed and skipped as noisy, **When** ingest runs again, **Then** it is not re-extracted and still produces no lore.
3. **Given** ingest failed after PR `#1842` was stored, **When** the operator retries, **Then** PR `#1842` is not duplicated and remaining unprocessed merged PRs may still be considered.
4. **Given** a repository whose cursor already covers merged PRs through a known merged time, **When** a new rationale PR is merged and ingest runs, **Then** only that new PR (and any other unprocessed merged PRs) can produce new candidates.
5. **Given** an operator requests a single PR by number that is already processed, **When** ingest runs, **Then** the run succeeds without creating a duplicate.

---

### User Story 3 - Operators trigger PR ingest and inspect status (Priority: P2)

An authorized writer or admin triggers PR ingest for a GitHub-mapped repository they belong to. They can read PR ingest run status (started, succeeded, failed, counts of PRs seen / skipped / candidates stored, cursor position, error summary). Readers with membership can list PR ingest runs and list PR-derived candidates for that repository. Unauthorized or cross-tenant callers are denied. Local-mode actor rules match other writes. Git commit ingest routes and CLI remain working. Agents cannot promote extracts to canonical through a new MCP write tool.

**Why this priority**: Constitution VIII: ingest is an observable operation. CLI + REST are the P0 surfaces; MCP stays at 10 tools.

**Independent Test**: CLI trigger + status on a fixture; REST trigger + list PR runs + list PR-derived candidates; membership deny for a foreign repository; existing git ingest still works; MCP tool count unchanged.

**Acceptance Scenarios**:

1. **Given** a writer with membership on repository `github.com/acme/payments`, **When** they run `memlore ingest pr --repository github.com/acme/payments`, **Then** an ingest run is recorded and candidates (if any) are stored for that scope.
2. **Given** a completed or failed PR run, **When** they run `memlore ingest status --repository github.com/acme/payments --kind pr` or GET PR ingest status over REST, **Then** they see run state, PR/candidate counts, cursor, and an error summary when failed.
3. **Given** a reader with membership on that repository, **When** they list PR ingest runs or PR-derived candidates, **Then** the list is scoped to that repository and does not include other tenants’ runs or lore.
4. **Given** a principal without membership on the requested repository (OIDC membership mode), **When** they trigger PR ingest or list PR status, **Then** access is denied the same way as other writes/reads for that scope.
5. **Given** the MCP tool list, **When** tools are enumerated, **Then** the count remains 10 and there is no new write tool that creates PR-derived lore.
6. **Given** existing git ingest CLI/REST, **When** F031 ships, **Then** git trigger, git run list, and default `ingest status` (git) still behave as before.

---

### User Story 4 - Trust boundary: compile still prefers verified architecture (Priority: P2)

PR-derived candidates are visible as governed lore (so they are not a hidden second store), but default compile / `get_for_task` ranking continues to put verified human/ADR architecture above unverified repository observations. Ingest does not auto-verify, does not apply trusted-source policy, and does not change compile contracts except that observational lore may appear among lower-ranked items when it is current and in budget. F035 accept/reject is not built; candidates remain unverified observational lore until a later feature promotes them.

**Why this priority**: Constitution V is non-negotiable. Capture must not look like canonical architecture in the agent briefing.

**Independent Test**: Seed a verified architecture statement and an ingested PR observation in the same repository; compile for a matching task; assert the verified architecture is ordered ahead of the PR observation; assert the PR item’s origin is `repository_observation` and status is unverified.

**Acceptance Scenarios**:

1. **Given** a verified human-authored architecture statement and an unverified PR-derived observation in the same repository, **When** context is compiled for a matching task, **Then** the verified architecture is ordered ahead of the PR observation.
2. **Given** only PR-derived unverified observations, **When** they appear in compile, **Then** they retain observational origin and unverified status (they are not relabeled as architecture decisions by ingest).
3. **Given** ingest has stored candidates, **When** a caller uses the existing verify API, **Then** ingest itself has not already marked them verified; promotion remains a separate human action (existing verify, or F035 later).

---

### Edge Cases

- Repository scope key that does not map to `github.com/{owner}/{repo}`: ingest run fails with a clear error; no lore is written.
- Missing or invalid GitHub credential: run fails observably; the credential is never written to logs or status text beyond “not configured” / “unauthorized”.
- GitHub repository not found or not accessible with the credential: run fails with a clear error; no lore is written.
- Empty set of merged PRs: run succeeds with zero candidates; cursor remains unset or records “nothing processed”.
- PR title+body longer than the lore statement limit: candidate is skipped (do not invent a shorter claim).
- PR already stored from a previous run with different extraction rules: v1 treats the PR as processed; do not rewrite the existing candidate (constitution VI: do not overwrite history).
- Multiple qualifying paragraphs or review threads in one PR: at most one candidate per PR in v1. Conservative volume.
- Concurrent PR ingest runs for the same repository: at most one active PR ingest run; a second trigger is rejected as conflict, not duplicated writers. A git ingest run for the same repository MAY run independently (separate operation).
- Graph-service down: lore and outbox still commit on the governance plane (same as other creates); graph catch-up remains the existing worker.
- Non-repository scopes: ingest is repository-scoped only; other scope kinds are rejected.
- REST create-lore remains human-authored only; PR extracts MUST NOT be writable via `POST /v1/lore-entries` as if they were human ADRs.
- GitHub rate limit or transient API failure mid-run: run fails; already-accepted PRs are not duplicated on retry.
- Bot-authored merged PRs (dependabot, renovate, github-actions): skipped even if the title contains a cue word.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Operators MUST be able to ingest **merged** pull requests from **GitHub** for an existing repository scope key. GitLab, Bitbucket, and other forges are out of scope. Webhooks and GitHub App installation are out of scope (F054).
- **FR-002**: Repository identity MUST bind to the existing repository scope key. Mapping for v1: scope key `github.com/{owner}/{repo}` → GitHub `{owner}/{repo}`. Other key shapes MUST be rejected.
- **FR-003**: v1 MUST ingest **merged PRs only**. Open, draft, and closed-unmerged PRs MUST be skipped and MUST NOT be treated as landed implementation.
- **FR-004**: For each processed merged PR, the system MUST be able to capture title, description, linked issues, changed files, review discussion used as evidence, merged state, author, and timestamps.
- **FR-005**: Extraction MUST produce **candidate** engineering knowledge, not a dump of every PR description. Skip rules (v1): unmerged; bot authors (dependabot, renovate, github-actions); conventional subjects that are only `chore`, `ci`, `style`, `build`, `deps`, or version-bump with no body **and** no human review rationale. Keep PRs whose title, body, or a human review comment contains supported rationale cues **and** whose extract is a faithful use of that text (no invented decisions).
- **FR-006**: At most **one** candidate lore entry MAY be created per PR (repository + number) in this slice. The statement MUST be derived from PR title, body, and selected review comments that were actually used — not from model inference or path-only guesses.
- **FR-007**: Each stored candidate MUST have origin `repository_observation`, verification status `unverified`, and MUST NOT set `human_verified` origin or verified status. Ingest MUST NOT use trusted-source policy (that is F032 for accepted ADRs).
- **FR-008**: Each stored candidate MUST include evidence of type `pr` whose value identifies the GitHub PR as `{owner}/{repo}#{number}`. Optional additional evidence: `url` for a review comment actually used; `path` for changed files; `url` for linked issues/tickets when present. Existing types `url`, `path`, `adr`, `commit` remain valid; `pr` is additive.
- **FR-009**: Re-ingest of the same PR for the same repository scope MUST be idempotent: one candidate (or zero, if skipped), never a duplicate. Processed PRs include skipped noisy and unmerged PRs that were explicitly requested or listed.
- **FR-010**: The system MUST persist a per-repository PR ingest cursor (watermark: last processed merged time and PR number) so re-runs are incremental. Failed runs MUST be retryable without duplicating already-stored candidates. Operational tables for PR runs, cursor, and processed PRs MUST be additive and MUST NOT overload git processed-SHA storage.
- **FR-011**: Candidates MUST be persisted as governed **lore entries** in the existing knowledge store (not a parallel `suggested_lore` knowledge table). Compile continues to read lore entries.
- **FR-012**: Creating a PR-derived lore entry MUST follow the same governance write unit as other lore creates: lore + audit + outbox episode ingest in one unit of work, so the existing graph worker can publish. Human `CreateLore` / MCP `remember` / REST create remain human-authored only. Observational create is the F030 path extended to accept PR evidence (not a silent reuse of `CreateLore`).
- **FR-013**: CLI MUST expose `memlore ingest pr --repository <key>` with optional `--pr <n>` (single PR) and optional `--max-prs` (positive integer cap for a run). CLI MUST extend `memlore ingest status --repository <key>` with optional `--kind git|pr` (default `git`, preserving F030 output). `--kind pr` prints a human-readable latest PR-run summary (JSON not required).
- **FR-014**: REST MUST provide membership-scoped trigger and status: trigger PR ingest (write), list/get PR ingest runs (read), and list PR-derived candidates for a repository (read). Trigger MUST NOT silently mark lore canonical. Existing `POST /v1/ingest/git` and git run listing MUST keep working. Existing `POST /v1/lore-entries` MUST remain human-authored. Repository keys MUST NOT be placed in the URL path.
- **FR-015**: MCP tool count MUST remain 10. No new MCP tool. No MCP write that creates or promotes PR extracts.
- **FR-016**: Triggering ingest requires **write** permission (writer or admin) plus F114 membership on the repository scope. Listing runs and candidates requires **read** plus membership. Local mode: mutating routes require `X-Memlore-Actor`; membership off; actor is trusted admin (same as other writes). CLI uses the same local Postgres wiring as `memlore profile` / `memlore ingest git` (not an HTTP client to `serve`).
- **FR-017**: Compile / `get_for_task` default ranking MUST continue to place verified architecture above unverified repository observations (existing authority evaluation). This slice MUST NOT change compile ranking formulas, packet section ids, or F007/F021/F030 contracts.
- **FR-018**: Ingest MUST be observable: structured logs for start/complete/fail with repository key, run id, counts, and error; ingest run records expose status to operators. GitHub credentials MUST never appear in logs.
- **FR-019**: Concurrent PR ingest for the same repository MUST NOT double-write candidates (one active PR ingest run per repository).
- **FR-020**: GitHub access uses an operator-supplied credential (environment token). `memlore worker` remains outbox → graph publisher only; PR polling MUST NOT be folded into the worker.
- **FR-021**: F032 ADR ingest, F035 accept/reject queue, F054 GitHub checks, F074 review bot, web UI, and non-GitHub forges are out of scope.

### Key Entities

- **PullRequestSnapshot**: A pull request read from GitHub: number, title, description, author, bot flag, merged state, merged timestamp, HTML URL, changed paths, review comments, linked issue URLs.
- **PRCandidate**: Observational lore derived from one PR: statement, origin `repository_observation`, unverified, evidence including `pr` (`{owner}/{repo}#{number}`), scoped to the repository.
- **PRIngestRun**: An observable operation for one repository: actor, state (running / succeeded / failed), counts (PRs seen, skipped, candidates stored), optional single-PR filter, cursor after the run, error summary.
- **PRIngestCursor**: Per-repository watermark of the last successfully processed merged PR (number + merged timestamp) used for incremental re-runs.
- **ProcessedPR**: Per-repository record that a PR number was considered (candidate stored or skipped), used for idempotency.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A fixture with one rationale merged PR, one noisy merged PR, and one unmerged PR yields exactly one stored candidate after ingest.
- **SC-002**: Ingesting the same fixture twice yields the same candidate count (no duplicates) and the same PR evidence on that candidate.
- **SC-003**: Every stored PR-derived candidate has origin `repository_observation`, verification `unverified`, and evidence type `pr` pointing at `{owner}/{repo}#{number}`.
- **SC-004**: A compile fixture with verified architecture plus that PR observation ranks the verified architecture ahead of the observation.
- **SC-005**: REST contract tests cover PR trigger, PR run status, PR-derived candidate list, and membership denial for a foreign repository; CLI prints a PR status summary for a completed run; git ingest routes still pass their contract tests.
- **SC-006**: MCP tool count remains 10; `POST /v1/lore-entries` still creates human-authored lore only.
- **SC-007**: A failed-then-retried ingest of a fixture does not duplicate the candidate that was stored before the failure.
- **SC-008**: `go test ./...` and `go vet ./...` are green.

## Assumptions

- **Knowledge store**: Persist candidates as `lore_entries` with observational origin + unverified status so F003 authority already downranks them vs verified human/ADR. F035 is not required to *hold* candidates.
- **PR source (v1)**: GitHub only, operator-triggered (CLI/REST), credential from environment (`MEMLORE_GITHUB_TOKEN` then `GITHUB_TOKEN`). No webhooks, no GitHub App, no `gh` CLI as the product adapter (HTTP port is testable without network). No GitLab/Bitbucket.
- **Merged-only**: Unmerged PRs are skipped, not labeled as not-landed lore. This is the public contract for v1.
- **Origin**: Same `repository_observation` as F030. A new origin is not required; evidence type distinguishes git vs PR.
- **Evidence**: New type `pr` with value `{owner}/{repo}#{number}` (e.g. `acme/payments#1842`). Do not overload `url` as the only PR identifier. Optional `url` for review comments actually used and for linked issues. Optional `path` for files (capped, same conservatism as git ingest).
- **Operational tables**: Additive `pr_ingest_runs`, `pr_ingest_cursors`, `pr_ingest_prs`. Do not store PR numbers in `git_ingest_shas` (SHA-shaped, different identity, independent concurrency).
- **Extractor**: Deterministic heuristic on title + body + human review comments + skip rules; not an LLM. Conservative: at most one candidate per PR; skip when unsupported. Do not reuse the commit-message extractor unchanged.
- **Operator interface**: Dedicated CLI `memlore ingest pr` rather than folding PR polling into `memlore worker`. REST trigger runs ingest in-process for v1. Status extends `memlore ingest status --kind pr`; default remains git.
- **Cursor**: Processed-PR set plus last merged timestamp and PR number; incremental by merged_at then number.
- **Created-by**: The triggering actor (CLI/REST actor id). Origin still `repository_observation`.
- **Compile**: No F007/F020/F021/F030 contract changes. PR observations may appear in compile items when they rank into budget; they must not outrank verified architecture via ingest-side flags. Add a characterization test (same pattern as F030 git).
- **MCP**: No 11th tool; status is CLI + REST.
- **Web UI**: Out of scope (F120).
- **Skip-rule details** (v1, documented for testers):
  - Unmerged: not merged (open, draft, closed without merge).
  - Bot: author login is `dependabot`, `dependabot[bot]`, `renovate`, `renovate[bot]`, `github-actions`, `github-actions[bot]`, or GitHub user type Bot for those logins.
  - Noisy prefix: title matches conventional-commit type `chore|ci|style|build|deps` with no body rationale and no human review rationale, or title is only a version bump.
  - Rationale cues (case-insensitive, title, body, or human review comment): `because`, `so that`, `workaround`, `migration`, `breaking`, `to fix`, `instead of`, `avoid`, `constraint`, `why:`, `decided`, `we should`.
  - Unmerged and bot skip win over cues. Noisy title loses if body or a human review comment has a cue.
- Historical IDs F004/F107 (outbox) are not reused; this is product F031.
