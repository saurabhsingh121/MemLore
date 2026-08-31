# MemLore Feature Development Tracker

**Last Updated**: 2026-09-01  
**Current Milestone**: M19 — product flywheel (next: F032 ADR ingest or F035 review queue; F022 is next Epic D compiler follow-up)  
**Current Release Target**: v0.9.0 engineering-intelligence flywheel  
**Foundation**: v0.8.0 knowledge plane + governance (F001–F010, F101–F114)

---

## Product Thesis

MemLore is not merely a database where agents store memories.

It is **the engineering intelligence layer that makes every coding agent
understand how your team actually builds software.**

```text
Generic memory
      ↓
"I remember something."

MemLore
      ↓
"I know what your team decided,
why they decided it,
whether it's still true,
what code implements it,
what contradicts it,
and what this agent needs to know right now."
```

The differentiator is not more memory primitives. It is engineering rationale,
provenance, authority, temporal validity, evidence, and implementation
awareness — compiled for the task at hand.

Foundation already shipped: scoped lore, authority, provenance, temporal
validity, conflict handling, Graphiti retrieval, context compilation,
verification, MCP, membership-scoped authz, and auditability. New work extends
that foundation around real engineering problems.

---

## Clarifications

### Session 2026-09-01

- Q: New product IDs F101–F103 and F110–F112 collide with completed platform IDs. → A: Keep F101–F114 frozen. Remap collisions to F115–F119 and F122.
- Q: How much completed TDD detail stays in this tracker? → A: Compact DONE Epic A/B tables with spec links. Full writeups only for new planned features.
- Q: P0 review-queue surface? → A: CLI + REST for P0 queues. Web UI waits for F120. GitHub checks/comments (F054/F074) are separate integration surfaces. MCP only where agents participate.

---

## Status Legend

| Status | Meaning |
|--------|---------|
| PLANNED | Identified; no spec yet |
| SPECIFYING | Spec Kit specify/clarify in progress |
| READY | Spec + plan + tasks; not coding |
| IN DEVELOPMENT | Active TDD implementation |
| BLOCKED | External dependency or decision needed |
| IN REVIEW | Implementation complete; verification pending |
| DONE | Acceptance criteria met; docs/tests updated |
| DEFERRED | Intentionally postponed |

---

## How to read this tracker

Work is organized by **product epic**, not a flat ID list.

| Epic | Name | Role |
|------|------|------|
| A | Memory Foundation | Completed control + knowledge planes |
| B | Trust & Governance | Completed authority, authn/z, audit |
| C | Knowledge Acquisition | Ingest git, PRs, ADRs, docs, agent sessions |
| D | Agent Intelligence | Profiles, `get_for_task`, feedback, cross-agent learning |
| E | Decision Intelligence | First-class decisions and `memlore why` |
| F | Engineering Drift | Intent vs implementation; PR checks |
| G | Developer Experience | PR intelligence, archaeology, knowledge graph UI |
| H | Integrations | GitHub-first forge adapters (features live in C/F/G) |
| I | Knowledge Health | Freshness, gaps, verification queues |
| J | Analytics | Health dashboard and outcome metrics |

**ID stability:** Completed platform tickets **F101–F114** are never reused.
Colliding product ideas from the 2026-09-01 roadmap were remapped:

| Roadmap name | Tracker ID | Frozen ID it would have collided with |
|--------------|------------|----------------------------------------|
| Cross-Agent Learning | **F115** | F101 Go skeleton |
| Session Outcome Capture | **F116** | F102 Go domain |
| Repository Learning Profile | **F117** | F103 Go persistence |
| Knowledge Coverage Analysis | **F118** | F110 invalidate/supersede |
| Missing Knowledge Detection | **F119** | F111 OIDC/RBAC |
| Knowledge Gap Review Queue | **F122** | F112 temporal/conflict filter |

Implementation history for DONE work lives in `specs/` and git. This file
tracks what to build next.

---

## Product flywheel

The first coherent loop — not every feature at once:

```text
             Git / PR / ADR
                  |
                  v
          Automatic Capture
                  |
                  v
            MemLore Graph
                  |
          authority + evidence
                  |
                  v
         get_for_task / why
                  |
                  v
             Coding Agent
                  |
                  v
                 PR
                  |
                  v
             Drift Check
                  |
                  v
          Human Correction
                  |
                  v
          MemLore Learns
                  |
                  +----------> next agent
```

```text
capture → trust → retrieve → use → detect drift → human correction → learn
```

### Recommended build sequence

1. **F020** — Repository Intelligence Profile
2. **F021** — Agent Context Bootstrap / `get_for_task` (extends F007)
3. **F030** — Git Commit Ingestion
4. **F031** — Pull Request Ingestion
5. **F032** — ADR Auto-Ingestion
6. **F035** — Suggested Lore Review Queue
7. **F040** — First-Class Decision Model
8. **F044** — `memlore why`
9. **F050** — Architecture Drift Detection
10. **F054** — GitHub PR Check Integration
11. **F100** — Agent Correction Capture (once drift/review is stable)

### Signature demo

One command should eventually explain the product:

```bash
$ memlore why src/payments/refund.go
```

```text
Why does this code use the transactional outbox?

ADR-023 · AUTHORITATIVE

Introduced after duplicate PaymentRefunded events
caused INC-481.

Decision:
All payment-domain events must be persisted atomically
with business state and published asynchronously.

Evidence:
  ADR-023
  PR #1842
  INC-481

Current implementation:
  ✓ Compliant

Last verified:
  12 days ago
```

That is the differentiator: rationale, provenance, authority, temporal
validity, evidence, and implementation awareness.

---

## Feature Summary

### Epic A — Memory Foundation (DONE)

| ID | Feature | Status | Spec | Notes |
|----|---------|--------|------|-------|
| F001 | Scoped human-authored lore (REST) | DONE | `001-scoped-lore-entry` | Go REST (Python core removed) |
| F002 | MCP lore tools | DONE | `002-mcp-lore-tools` | remember/get/verify/explain/search |
| F004 | Transactional outbox + graph sync | DONE | `010-transactional-outbox` | Delivered as F107 |
| F005 | Graph knowledge service | DONE | `009-graph-service` | `graph-service/` Graphiti boundary |
| F006 | Semantic search + graph retrieval | DONE | `011-…` + `019-…` | F108 + fuller retrieval |
| F007 | Context compiler + `get_for_task` v1 | DONE | `012-context-compiler` | F109 + F112 stages; F021 extends |
| F008 | Supersession + invalidation | DONE | `013-supersede-invalidate` | Delivered as F110 |
| F009 | Conflict detection | DONE | `014-conflict-filtering` | Delivered as F112 |

### Epic B — Trust & Governance (DONE)

| ID | Feature | Status | Spec | Notes |
|----|---------|--------|------|-------|
| F003 | Authority factor model + evaluation | DONE | `016-authority-factors` | Ephemeral factors + trust bands |
| F010 | Auth (OIDC) + team/project scopes | DONE | `015-…` + `018-…` | F111 authn+RBAC; F114 membership |
| F114 | Membership-scoped authorization | DONE | `018-membership-authz` | Goose 00004 |

### Epic C — Knowledge Acquisition

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F030 | Git commit ingestion | P0 | DONE | `specs/030-git-commit-ingest/`; observational, not canonical |
| F031 | Pull request ingestion | P0 | DONE | `specs/031-pull-request-ingest/`; observational, not canonical |
| F032 | ADR auto-ingestion | P0 | PLANNED | Accepted ADRs → high source authority |
| F033 | Documentation ingestion | P1 | PLANNED | Architecture/runbooks/standards only |
| F034 | Agent session knowledge extraction | P1 | PLANNED | Labeled `agent_observation` / `agent_inference` |
| F035 | Suggested Lore review queue | P0 | PLANNED | CLI + REST; nothing auto-becomes canonical |

### Epic D — Agent Intelligence

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F020 | Repository intelligence profile | P0 | DONE | `specs/020-repo-intelligence-profile/` |
| F021 | Agent context bootstrap / `get_for_task` | P0 | DONE | `specs/021-agent-context-bootstrap/`; extends F007 |
| F022 | Context packet profiles | P0 | PLANNED | coding / review / debug / architecture / incident / onboarding |
| F023 | Token-budgeted agent briefing | P0 | PLANNED | Extends F007 budgeting; priority ladder |
| F060 | Context usage feedback | P0 | PLANNED | Retrieval signal only; not authority |
| F061 | Retrieval quality metrics | P1 | PLANNED | Hit rate, unused, latency, conflicts |
| F062 | Context usefulness score | P1 | PLANNED | Must not override authority |
| F063 | Ranking feedback loop | P1 | PLANNED | Tune ranking by workflow profile |
| F064 | Retrieval experiment framework | P2 | PLANNED | Offline / A-B evaluation |
| F100 | Agent correction capture | P0 | PLANNED | After F050/F054/F035 are stable |
| F115 | Cross-agent learning | P1 | PLANNED | Formerly roadmap F101 |
| F116 | Session outcome capture | P1 | PLANNED | Formerly roadmap F102 |
| F117 | Repository learning profile | P1 | PLANNED | Formerly roadmap F103 |

### Epic E — Decision Intelligence

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F040 | First-class decision model | P0 | PLANNED | Dedicated domain, not generic lore only |
| F041 | Decision timeline | P1 | PLANNED | Previous truth remains discoverable |
| F042 | Alternative tracking | P1 | PLANNED | Why X, why not Y |
| F043 | Decision impact graph | P1 | PLANNED | Affected services, events, components |
| F044 | `memlore why` | P0 | PLANNED | Signature developer workflow |

### Epic F — Engineering Drift

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F050 | Architecture drift detection | P0 | PLANNED | Intent vs observed implementation |
| F051 | Convention violation detection | P1 | PLANNED | Machine-readable conventions |
| F052 | Drift severity scoring | P1 | PLANNED | CRITICAL…INFORMATIONAL |
| F053 | Drift review workflow | P1 | PLANNED | CLI + REST; resolution becomes lore |
| F054 | GitHub PR check integration | P0 | PLANNED | First forge adapter (Epic H) |

### Epic G — Developer Experience

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F070 | PR context retrieval | P1 | PLANNED | Context for changed files |
| F071 | Changed-file lore matching | P1 | PLANNED | Decisions, owners, risks |
| F072 | PR decision summary | P1 | PLANNED | Candidate lore from the PR |
| F073 | Historical risk context | P1 | PLANNED | Incidents, regressions, failed approaches |
| F074 | GitHub review bot | P1 | PLANNED | One concise summary; not noisy |
| F090 | File explanation | P1 | PLANNED | `memlore explain <path>` |
| F091 | `why-line` | P1 | PLANNED | Blame = who; MemLore = why |
| F092 | Historical change narrative | P2 | PLANNED | Component timeline |
| F120 | Engineering knowledge graph UI | P2 | PLANNED | First product web UI |
| F121 | Component relationship explorer | P2 | PLANNED | Graph navigation |

### Epic H — Integrations

GitHub is the first forge. Feature writeups live with their product epic.

| ID | Feature | Home epic | Priority | Status |
|----|---------|-----------|----------|--------|
| F031 | Pull request ingestion | C | P0 | DONE |
| F054 | GitHub PR check | F | P0 | PLANNED |
| F074 | GitHub review bot | G | P1 | PLANNED |

Other forges (GitLab, Bitbucket, etc.) are out of scope until the GitHub
flywheel works.

### Epic I — Knowledge Health

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F080 | Freshness scoring | P1 | PLANNED | Age, evidence, supersession, contradiction |
| F081 | Stale lore detection | P1 | PLANNED | Surface for human review |
| F082 | Evidence revalidation | P1 | PLANNED | File/ADR/PR still supports the claim |
| F083 | Knowledge expiry policies | P2 | PLANNED | Task context, workarounds, mitigations |
| F084 | Periodic verification queue | P1 | PLANNED | CLI + REST; authority × usage × staleness × impact |
| F118 | Knowledge coverage analysis | P2 | PLANNED | Formerly roadmap F110 |
| F119 | Missing knowledge detection | P2 | PLANNED | Formerly roadmap F111 |
| F122 | Knowledge gap review queue | P2 | PLANNED | Formerly roadmap F112 |

### Epic J — Analytics

| ID | Feature | Priority | Status | Notes |
|----|---------|----------|--------|-------|
| F130 | MemLore health dashboard | P2 | PLANNED | After F120 or as CLI/REST metrics first |
| F131 | Agent context effectiveness | P2 | PLANNED | Outcome metrics, not vanity retrieval counts |

### Completed platform delivery (frozen IDs)

| ID | Feature | Status | Spec |
|----|---------|--------|------|
| F101 | Go project skeleton + tooling | DONE | `003-go-core-skeleton` |
| F102 | Go domain primitives | DONE | `004-go-domain-primitives` |
| F103 | Go PostgreSQL persistence | DONE | `005-go-postgres-persistence` |
| F104 | Migrate lore CRUD/verify REST to Go | DONE | `006-go-rest-lore-crud` |
| F105 | Migrate MCP lore tools to Go | DONE | `007-go-mcp-lore-tools` |
| F106a | Go governance hardening + Python cutover | DONE | `008-go-governance-hardening` |
| F106 | Extract graph-service + contracts | DONE | `009-graph-service` |
| F107 | Transactional outbox + graph sync worker | DONE | `010-transactional-outbox` |
| F108 | Graph retrieval orchestration | DONE | `011-graph-retrieval-orchestration` |
| F109 | Context compiler + get_for_task v1 | DONE | `012-context-compiler` |
| F110 | Invalidate + supersede lifecycle | DONE | `013-supersede-invalidate` |
| F111 | OIDC + RBAC | DONE | `015-oidc-rbac` |
| F112 | Temporal filter + conflict detection | DONE | `014-conflict-filtering` |
| F113 | Retire legacy Python core | DONE | `017-retire-python-core` |

---

## EPIC A — Memory Foundation

**Status**: DONE (v0.8.0)

Governance-plane lore, dual-plane search, context compilation, lifecycle, and
the Go strangler. Specs remain the source of acceptance history.

| Concern | How it shipped |
|---------|----------------|
| Scoped lore CRUD, verify, audit | F001 / F104 |
| MCP `memlore.*` | F002 / F105 |
| Graph isolation | F005 / F106 |
| Outbox sync | F004 / F107 |
| Semantic + graph retrieval | F006 / F108 / `019-semantic-graph-retrieval` |
| Token-budgeted compile + `get_for_task` | F007 / F109 |
| Supersede / invalidate | F008 / F110 |
| Temporal filter + conflicts | F009 / F112 |
| Python core removed | F113 |

**F007 remains DONE.** F021–F023 are product-grade enhancements of that
compiler, not a reopening of F007.

---

## EPIC B — Trust & Governance

**Status**: DONE

| Concern | How it shipped |
|---------|----------------|
| Explainable authority factors + trust bands | F003 / `016-authority-factors` |
| Optional OIDC + reader/writer/admin | F010 / F111 / `015-oidc-rbac` |
| Team/project membership ACL | F010 / F114 / `018-membership-authz` |

New acquisition and agent-origin features MUST keep agent observations
distinguishable from human-verified knowledge (constitution V). F035 exists so
extraction cannot silently become canonical.

---

## EPIC C — Knowledge Acquisition

Automatically captured knowledge is **candidate lore**. It does not become
authoritative until a human accepts it (F035) or an existing high-authority
source type applies (accepted ADRs in F032).

### F030 — Git Commit Ingestion

**Status**: DONE  
**Priority**: P0  
**Depends on**: F004/F107 (outbox), F005/F106 (graph ingest)

**Goal**: Ingest Git commit history and extract candidate engineering knowledge
(rationale, migration context, bug explanations, component relationships,
technical constraints).

**Product value**: Recovers *why* from history that never became an ADR.

**Surfaces**: CLI `memlore ingest git` / `memlore ingest status`; REST trigger
and run/candidate list; existing `memlore worker` publishes outbox. No silent
canonical writes. No new MCP tool.

**Acceptance criteria**:

- [x] Commits from a configured local git directory can be ingested with author, SHA, timestamp, message, and changed-path metadata
- [x] Extracted candidates are stored as observational evidence, not `canonical` / human-verified
- [x] Each candidate preserves an evidence link to the commit SHA (`evidence.type=commit`)
- [x] Re-ingest of the same SHA is idempotent
- [x] Failed ingest retries without duplicating accepted lore

**Spec**: `specs/030-git-commit-ingest/`

### F031 — Pull Request Ingestion

**Status**: DONE  
**Priority**: P0  
**Depends on**: F030 (observational ingest producer), Epic H GitHub-first  
**Home also**: Epic H  
**Spec**: `specs/031-pull-request-ingest/`

**Goal**: Ingest PR title, description, linked issues, changed files, review
discussion, merged state, author, and timestamps. PRs are high-quality sources
for why code changed.

**Product value**: Review discussion is often the real decision record.

**Surfaces**: CLI `memlore ingest pr` / `memlore ingest status --kind pr`; REST
`POST /v1/ingest/pr`, `GET /v1/ingest/pr-runs`, candidates with
`evidence_type=pr`. Existing `memlore worker` publishes outbox. No silent
canonical writes. No new MCP tool.

**Acceptance criteria**:

- [x] Merged PRs from a configured GitHub repository can be ingested
- [x] Candidates preserve evidence links to the original PR (and review comments used)
- [x] Git-derived and PR-derived candidates remain observational until F035 accept
- [x] Linked issues/tickets are stored as evidence refs when present
- [x] Unmerged PRs are skipped (not treated as landed implementation)

**Next step**: Specify **F032** (ADR ingest) or **F035** (suggested-lore review queue).

### F032 — ADR Auto-Ingestion

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F001 (lore model), F003 (authority), F008 (supersession)

**Goal**: Discover and ingest ADRs from configured paths (`docs/adr/`, `adr/`,
`architecture/decisions/`, and repo-configured extras). Extract decision,
status, context, alternatives, consequences, supersession, and affected
components.

**Product value**: Turns the team’s existing decision corpus into governed lore
without copy-paste.

**Acceptance criteria** (draft):

- [ ] Configured paths are scanned; new/changed ADRs produce candidates
- [ ] Accepted ADRs receive high source authority (not unverified agent inference)
- [ ] Supersession relationships in ADR metadata map to lore supersession where possible
- [ ] Original ADR path remains the evidence source
- [ ] Human review (F035) still applies if extraction is uncertain; accepted-file ingest may be auto-trusted only when specify says so

**Next step**: Specify after or with F030/F031; needed before F040/F044 shine.

### F033 — Documentation Ingestion

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F030–F032 ingest pipeline

**Goal**: Ingest selected engineering documentation (architecture docs,
runbooks, coding standards, service docs, migration notes) without becoming a
general document search engine.

**Product value**: Trusted engineering docs become evidence-backed context.

**Acceptance criteria** (draft):

- [ ] Only configured doc classes/paths are eligible
- [ ] Original document remains the evidence source
- [ ] Candidates are not canonical until review or an explicit trusted-source policy
- [ ] Arbitrary wiki/dump ingestion is out of scope

**Next step**: Specify after P0 acquisition works.

### F034 — Agent Session Knowledge Extraction

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F035, F100; F002 remember already exists

**Goal**: Allow agent sessions to produce candidate lore (conventions,
unexpected behavior, debugging discoveries, workarounds, architectural
mismatches).

**Product value**: Coding sessions stop being a knowledge dead-end.

**Acceptance criteria** (draft):

- [ ] Extracted items are labeled `agent_observation` or `agent_inference` until verified
- [ ] They cannot receive `canonical` trust from extraction alone (F003 invariant)
- [ ] F035 review is the promotion path
- [ ] Prompt-injection / untrusted-agent content is treated as untrusted input

**Next step**: Specify after F035.

### F035 — Suggested Lore Review Queue

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F030–F032 (producers); F003 (authority); F010 (who may accept)

**Goal**: Automatically extracted knowledge must not silently become
authoritative. Humans Accept / Edit / Reject candidates with evidence and
confidence visible.

**Product value**: Safe path from extraction to trusted engineering knowledge.

**Surfaces**: **CLI + REST** (P0). No product web UI. MCP may list/read
candidates for agents; mutating accept/reject is a human (or authorized
human-equivalent) action.

Example (CLI/REST payload, not a GUI):

```text
Suggested Lore
Decision: Payment events use transactional outbox.
Reason: Prevent duplicate Kafka publishing.
Evidence: PR #1842, PAY-381
Confidence: 0.84
[Accept] [Edit] [Reject]
```

**Acceptance criteria** (draft):

- [ ] Candidates appear in a review queue with statement, reason, evidence, confidence, source type
- [ ] Accept creates/updates governed lore with human verification provenance
- [ ] Edit then accept records the human-authored statement, not the raw extract as-if-human
- [ ] Reject records a negative decision so the same extract is not re-prompted forever
- [ ] Queue is scoped by membership (F114)
- [ ] Nothing in the queue is treated as canonical until accept

**Next step**: Specify as soon as the first ingest producer (F030/F031/F032)
can create candidates — same flywheel slice if possible.

---

## EPIC D — Agent Intelligence

### F020 — Repository Intelligence Profile

**Status**: DONE  
**Priority**: P0  
**Spec**: `specs/020-repo-intelligence-profile/`  
**Depends on**: F006, F007 (retrieval + compile). Ingest (F030–F032) enriches later.

**Goal**: Create a compact intelligence profile for a repository: the
engineering context a human or coding agent needs before exploring dozens of
files.

Where available, the profile includes: architecture summary, important
architectural decisions, major technologies, coding conventions, repository
ownership, known gotchas, active migrations, architectural hotspots,
operational risks, recent important changes, related services and
dependencies.

**Product value**: Useful repository overview without archaeology.

**Surfaces**: REST `POST /v1/repository-profile`, MCP `memlore.repo_profile`,
CLI `memlore profile --repository`. Token-bounded; compiled from existing
lore/graph, not a second knowledge store.

**Acceptance criteria**:

- [x] Given a repository scope the caller is authorized to read, MemLore returns a structured profile
- [x] Missing sections are omitted — not hallucinated
- [x] Decisions cite evidence when present
- [x] Profile generation respects F114 membership and F007-style token limits
- [x] Output is usable by humans (CLI) and agents (MCP/REST)

**Next step**: F021 agent context bootstrap is DONE.

### F021 — Agent Context Bootstrap / `get_for_task`

**Status**: DONE  
**Priority**: P0  
**Spec**: `specs/021-agent-context-bootstrap/`  
**Depends on**: F007 (v1 compiler DONE), F020 (profile as an input/section)
**Extends**: F007

**Goal**: Make `memlore.get_for_task` the primary agent entry point with richer
inputs and a compiled packet of *useful* sections, not a bag of similar text.

Inputs may include: repository, branch, task/ticket, changed files, working
files, query, token budget, agent identity.

Output packet sections (omitted when empty):

```text
Relevant Architecture
Applicable Decisions
Coding Conventions
Task Context
Known Gotchas
Conflicts
Evidence / Sources
```

Observed implementation drift and potentially stale knowledge are omitted in
this slice (F050 / `include_stale` not added). Conflicts stay on the existing
`conflicts` array. F007 `items[]` remains.

**Product value**: Immediate reduction in context-discovery cost and tokens.

**Surfaces**: REST `POST /v1/context/compile` (additive fields), MCP
`memlore.get_for_task` (same; tool count stays 10), CLI `memlore context`.

**Acceptance criteria**:

- [x] `get_for_task` / compile accept the richer input set (unspecified fields optional)
- [x] Packet exposes the section types above when data exists
- [x] Empty sections are omitted rather than padded
- [x] F007 ranking, temporal filter, conflicts, and authority evaluation still apply
- [x] REST and MCP stay in parity
- [x] Agent identity is recorded for later F060 feedback, not used as authority

**Next step**: Specify **F022** (packet profiles) or **F032** (ADR ingest) /
**F035** (suggested-lore review queue) per flywheel sequence.

### F022 — Context Packet Profiles

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F021

**Goal**: Different packet styles by workflow: `coding`, `code_review`,
`debugging`, `architecture`, `incident`, `onboarding`. Profiles influence
ranking and token allocation.

```bash
memlore context --profile debugging
```

**Product value**: The same repo should not brief a debugger like an architect.

**Acceptance criteria** (draft):

- [ ] A documented set of profiles is selectable on compile / CLI / MCP
- [ ] Unknown profile is a validation error
- [ ] Profiles change ranking/token allocation in a testable way
- [ ] Default profile is `coding` unless specify says otherwise

**Next step**: Specify after F021 (deferred from the F021 slice; F021 ships a
single compile packet without a `profile` field).

### F023 — Token-Budgeted Agent Briefing

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F007 (budgeting exists), F021–F022
**Extends**: F007 token budgeting

**Goal**: Generate packets under an explicit token budget. Maximize useful
context per token, not retrieval volume.

Priority ladder:

1. High-authority architecture/policy
2. Task-specific context
3. Relevant repository conventions
4. Current implementation observations
5. Conflicts and warnings
6. Supporting evidence

**Acceptance criteria** (draft):

- [ ] Caller-supplied token budget is honored (existing F007 behavior retained)
- [ ] Drop order follows the ladder when over budget (testable fixtures)
- [ ] Conflicts/warnings are not silently dropped ahead of low-value evidence
- [ ] Budget and dropped-section counts are visible on the packet (specify exact fields)

**Next step**: Specify with F021/F022 as compiler hardening.

### F060 — Context Usage Feedback

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F021

**Goal**: Coding agents report whether retrieved context was USED, NOT_USED, or
IRRELEVANT. Retrieval feedback only — never authority truth.

**Acceptance criteria** (draft):

- [ ] REST/MCP accepts per-item feedback against a compile/get_for_task result
- [ ] Feedback cannot raise trust band or mark lore verified
- [ ] Feedback is auditable (who/agent, packet id, item id, signal)
- [ ] Missing feedback is allowed; quality metrics degrade gracefully

**Next step**: Specify after F021 is in use (dogfood), before F063.

### F061 — Retrieval Quality Metrics

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F060

**Goal**: Measure context hit rate, unused context, irrelevant context, packet
size, retrieval latency, conflicts surfaced, source coverage.

**Next step**: Specify after F060; expose via observability (constitution VIII)
and later F130.

### F062 — Context Usefulness Score

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F060, F003

**Goal**: A usefulness signal from downstream usage. Popularity MUST NOT
override authority.

**Next step**: Specify with F063; document the non-override invariant as an AC.

### F063 — Ranking Feedback Loop

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F062, F022

**Goal**: Improve ranking with feedback: boost repeatedly useful task-specific
sources, demote repeatedly irrelevant memories, detect too-generic context,
tune by workflow profile.

**Next step**: Specify after F062; keep ranking explainable (F003).

### F064 — Retrieval Experiment Framework

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F061–F063

**Goal**: Offline or A/B evaluation of ranking formulas, token budgets, query
expansion, graph depth, and authority weighting.

**Next step**: Defer until the flywheel produces enough labeled feedback.

### F100 — Agent Correction Capture

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F035, F050 or F054 (a correction channel exists)

**Goal**: Capture explicit human corrections of agent behavior as candidate
lore (e.g. reviewer: “No. Use transactional outbox.”) with evidence (PR review)
and `human_verified` only after review/accept.

**Product value**: One agent’s mistake becomes the next agent’s constraint.

**Acceptance criteria** (draft):

- [ ] A correction can be recorded with the rejected action, the rule, and evidence
- [ ] Corrections enter F035 (or a dedicated correction queue with the same accept/edit/reject semantics)
- [ ] Promoted rules are available to other authorized agents in the same scopes (F115)
- [ ] Agent-originated restatements remain labeled until human accept

**Next step**: Specify immediately after the first drift/review workflow is
stable (after F050/F054).

### F115 — Cross-Agent Learning

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F100, F114  
**Roadmap ID**: F101 (remapped)

**Goal**: Knowledge learned in one coding-agent workflow is available to other
authorized agents in the same membership scopes.

**Acceptance criteria** (draft):

- [ ] Lore accepted from a Claude (etc.) correction is retrievable by Codex (etc.) under the same authz
- [ ] No cross-tenant leak (F114)
- [ ] Agent identity is provenance, not a visibility wall inside the team

**Next step**: Specify with or right after F100.

### F116 — Session Outcome Capture

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F034, F035  
**Roadmap ID**: F102 (remapped)

**Goal**: Optionally capture discovered facts, decisions, failed approaches,
successful resolution, and remaining uncertainty at session end. Only useful
information is promoted.

**Next step**: Specify after F034/F035.

### F117 — Repository Learning Profile

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F020, F100, F115  
**Roadmap ID**: F103 (remapped)

**Goal**: Repository-specific lessons: common mistakes, recurring review
feedback, dangerous areas, preferred implementation patterns.

**Next step**: Specify after F020 and F100 have data to summarize.

---

## EPIC E — Decision Intelligence

### F040 — First-Class Decision Model

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F001 (lore), F008 (supersession), F032 (ADR ingest strongly preferred)

**Goal**: Treat engineering decisions as a dedicated domain concept, not only
generic memories.

A decision supports: question/problem, decision, rationale, alternatives,
consequences, owner, date, affected components, evidence, superseded decision,
current validity.

**Product value**: Agents can answer “why Kafka?” as a decision, not a snippet.

**Acceptance criteria** (draft):

- [ ] Decisions can be created, retrieved, and superseded without deleting history
- [ ] Required fields from the model above are represented (optional fields allowed)
- [ ] Decisions participate in compile / `get_for_task` as first-class sections
- [ ] Evidence and owner are preserved
- [ ] REST + CLI + MCP parity for read; write via REST/CLI (and MCP remember-shaped API if specify agrees)

**Next step**: Specify after F032 or in parallel if manual decision entry is the
MVP and ADR ingest lands immediately after.

### F041 — Decision Timeline

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F040

**Goal**: Show how a decision evolved (adopted → proposed migration → selected
→ compatibility removed). Previous truth stays historically discoverable.

**Next step**: Specify after F040; reuse F008 history + explain.

### F042 — Alternative Tracking

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F040

**Goal**: Capture alternatives considered so agents can answer why not SQS /
whether DynamoDB was considered / which tradeoff won.

**Next step**: Specify as F040 fields if that keeps the model smaller; split
only if query/UX needs it.

### F043 — Decision Impact Graph

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F040, F005

**Goal**: Show what a decision affects (services, events, publishers,
components) for impact analysis before changing architecture.

**Next step**: Specify after F040; graph-service relationships, not a new DB.

### F044 — `memlore why`

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F040, F007/F021

**Goal**: Signature developer command: `memlore why "Kafka"` or
`memlore why src/payments/refund.go`. Explain the decision, alternatives,
introduction date, current validity, and evidence.

**Product value**: This is the demo that should sell the product.

**Surfaces**: CLI first; REST + MCP equivalents required for agents.

**Acceptance criteria** (draft):

- [ ] Natural-language topic and path/file inputs are supported
- [ ] Output includes decision, reason, alternatives, introduced-at, still-valid, evidence
- [ ] Stale/superseded decisions are labeled, not presented as current
- [ ] Unauthorized scopes return not-found/forbidden without leaking existence beyond existing authz policy
- [ ] Implementation-awareness (compliant vs drift) may be “unknown” until F050

**Next step**: Specify after F040; can stub “still valid / compliant” until F050.

---

## EPIC F — Engineering Drift

### F050 — Architecture Drift Detection

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F040 (intent), F006 (retrieval); benefits from F030/F031 (code observations)

**Goal**: Detect differences between authoritative engineering intent and
observed implementation.

Categories: architecture violation, coding convention violation, dependency
violation, deprecated technology use, missing security control, operational
policy mismatch.

**Product value**: MemLore notices when the code stopped matching the decision.

**Acceptance criteria** (draft):

- [ ] Given an authoritative rule and an observation, the system can emit `IMPLEMENTATION_DRIFT` (or equivalent) with both sides preserved
- [ ] Neither side is silently dropped (F009 invariant)
- [ ] Drift is scoped and authorized (F114)
- [ ] Confidence of the observation is visible
- [ ] CLI + REST can list open drift for a repository/PR

**Next step**: Specify after F040; observations may start conservative (file/PR
signals before deep static analysis).

### F051 — Convention Violation Detection

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F050, trusted convention lore

**Goal**: Treat coding conventions as machine-readable expectations and detect
likely violations in changed code.

**Next step**: Specify after F050; start with conventions that are already
structured, not free-text only.

### F052 — Drift Severity Scoring

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F050, F003

**Goal**: Score CRITICAL / HIGH / MEDIUM / LOW / INFORMATIONAL from authority
of the rule, code scope, production/security risk, and observation confidence.

**Next step**: Specify after F050; keep the score explainable.

### F053 — Drift Review Workflow

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F050, F035-style queue

**Goal**: Not every apparent violation is wrong. Confirm drift, accept
exception, update architecture, mark false positive, or create follow-up.
Preserve the resolution as lore.

**Surfaces**: CLI + REST (same P0 queue policy).

**Next step**: Specify after F050; reuse F035 mechanics where possible.

### F054 — GitHub PR Check Integration

**Status**: PLANNED  
**Priority**: P0  
**Depends on**: F050 (meaningful check body); F031 useful for PR metadata
**Home also**: Epic H

**Goal**: Expose MemLore as a GitHub PR check so drift and decision violations
show up next to tests and lint.

**Product value**: Everyday developer workflow, not a separate console.

**Acceptance criteria** (draft):

- [ ] A GitHub check run is created for configured repositories
- [ ] Findings cite the violated lore/ADR and a code location when known
- [ ] Absence of findings is a passing check, not silence
- [ ] Check is not noisy: actionable findings only (threshold via F052 later)
- [ ] Authz: only configured installation + MemLore membership

**Next step**: Specify after F050 can produce at least one deterministic finding type.

---

## EPIC G — Developer Experience

### F070 — PR Context Retrieval

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F021, F031

**Goal**: For a pull request, retrieve context relevant to the changed files
(decisions, prior PRs, incidents, conventions).

**Next step**: Specify after F021 and F031.

### F071 — Changed-File Lore Matching

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F070

**Goal**: Given changed files, identify related decisions, conventions,
incidents, owners, known risks, and dependencies.

**Next step**: Specify with F070.

### F072 — PR Decision Summary

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F031, F040, F035

**Goal**: Concise summary of decisions introduced or modified by a PR; later
candidate lore.

**Next step**: Specify after F040 and F031.

### F073 — Historical Risk Context

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F070

**Goal**: Surface incidents, regressions, and previous failed approaches
relevant to changed code.

**Next step**: Specify with F070/F071.

### F074 — GitHub Review Bot

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F054, F070
**Home also**: Epic H

**Goal**: One concise PR comment or check summary: relevant decisions, drift,
stale context, related incidents, missing evidence. Avoid noisy comments.

**Next step**: Specify after F054; default to check-run only if comments are
too noisy.

### F090 — File Explanation

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F044, F031

**Goal**: `memlore explain src/refund/retry.go` — why the file exists, introducing
PR, related incident, relevant ADR.

**Next step**: Specify after F044; CLI + REST + MCP.

### F091 — `why-line`

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F090, F030

**Goal**: `memlore why-line src/refund/retry.go:143` — Git blame explains who
changed it; MemLore explains why.

**Next step**: Specify after F090.

### F092 — Historical Change Narrative

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F041, F090

**Goal**: Timeline around a component: original design → incident → workaround
→ ADR → refactor.

**Next step**: Defer until decision + archaeology data exist.

### F120 — Engineering Knowledge Graph UI

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F005, F040, F043

**Goal**: First product web UI over the Graphiti knowledge model: click an
entity to see decisions, owners, incidents, PRs, dependencies, conventions,
risks, history.

**Note**: P0 review queues explicitly do **not** wait on this UI.

**Next step**: Separate specify when the API flywheel is proven.

### F121 — Component Relationship Explorer

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F120 or a CLI/REST graph browse MVP

**Goal**: Navigate relationships (USES, DEPENDS_ON, PUBLISHES, SUPERSEDES,
CONTRADICTS, EVIDENCED_BY, …).

**Next step**: Specify with F120.

---

## EPIC H — Integrations

**Strategy**: GitHub first. Do not start GitLab/Bitbucket until F031 + F054
work on GitHub.

Integration features are specified in their home epics (C, F, G). This epic
exists so forge work is not scattered without an owner.

**Out of scope until the GitHub flywheel is done**: other forges, generic
“all git hosts” abstractions, and a marketplace listing.

---

## EPIC I — Knowledge Health

### F080 — Freshness Scoring

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F003, F008, F009

**Goal**: Estimate whether knowledge remains current using age, recent code
changes, evidence freshness, validity dates, contradicting observations, and
supersession.

**Next step**: Specify after F050 observations exist; can start with
age + supersession + explicit validity only.

### F081 — Stale Lore Detection

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F080

**Goal**: Surface likely-stale knowledge for human review (e.g. “services use
Java 17” vs repo reality).

**Next step**: Specify with F080; queue via F084.

### F082 — Evidence Revalidation

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F080, F030–F032

**Goal**: Periodically check whether evidence still exists and still supports
the claim (file removed, ADR superseded, PR reverted, docs rewritten).

**Next step**: Specify after ingest can resolve evidence URIs/SHAs.

### F083 — Knowledge Expiry Policies

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F008, F080

**Goal**: Explicit or policy-based expiration for task context, migration
notes, temporary workarounds, incident mitigations.

**Next step**: Defer; invalidation/supersession may be enough at first.

### F084 — Periodic Verification Queue

**Status**: PLANNED  
**Priority**: P1  
**Depends on**: F081, F035 mechanics

**Goal**: High-impact stale knowledge for human verification, prioritized by
authority × usage × staleness × impact.

**Surfaces**: CLI + REST.

**Next step**: Specify after F081; reuse F035 queue UX.

### F118 — Knowledge Coverage Analysis

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F020, F040  
**Roadmap ID**: F110 (remapped)

**Goal**: Measure whether important components have supporting knowledge
(architecture, decision, operational, ownership coverage).

**Next step**: Defer until F020/F040 exist to define “covered.”

### F119 — Missing Knowledge Detection

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F118  
**Roadmap ID**: F111 (remapped)

**Goal**: Detect important gaps (no retry policy, no runbook, no owner, no ADR
for a critical component).

**Next step**: Specify with F118.

### F122 — Knowledge Gap Review Queue

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F119, F035  
**Roadmap ID**: F112 (remapped)

**Goal**: Teams prioritize missing documentation/context by criticality and
agent usage.

**Surfaces**: CLI + REST (web UI only if F120 exists).

**Next step**: Specify after F119.

---

## EPIC J — Analytics

### F130 — MemLore Health Dashboard

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F061, F080, F118, F050

**Goal**: Coverage, verified %, stale %, unresolved conflicts, drift, agent
context hit rate, and monthly promotion/supersession counts.

**Surfaces**: REST metrics first; visual dashboard may wait for F120.

**Next step**: Specify after flywheel metrics exist; do not block P0 on a UI.

### F131 — Agent Context Effectiveness

**Status**: PLANNED  
**Priority**: P2  
**Depends on**: F060, F054, F100

**Goal**: Measure whether MemLore context improves engineering outcomes (PR
rework, reviewer corrections, repeated agent mistakes, tokens before
implementation, task success). Long-term business case, not a v0.9 launch
requirement.

**Next step**: Defer; requires production usage and careful causal claims.

---

## Development Ledger Notes

### 2026-09-01 — F031 pull request ingestion

- Merged GitHub PRs only; observational `repository_observation` lore
- Evidence type `pr` (`owner/repo#N`); optional `url` for used review comments and linked issues
- Additive `pr_ingest_*` tables (not `git_ingest_shas`); PR-number idempotency
- CLI `memlore ingest pr` / `ingest status --kind pr`; REST `/v1/ingest/pr`
- MCP unchanged (10 tools); compile ranking unchanged
- Spec: `specs/031-pull-request-ingest/`; F031 marked DONE

### 2026-09-01 — F030 git commit ingestion

- Local git directory ingest; observational `repository_observation` lore
- Evidence type `commit` (full SHA); skip noisy merges/chores; SHA idempotency
- CLI `memlore ingest git` / `memlore ingest status`; REST `/v1/ingest/git`
- MCP unchanged (10 tools); compile ranking unchanged
- Spec: `specs/030-git-commit-ingest/`; F030 marked DONE

### 2026-09-01 — F021 agent context bootstrap

- Extends F007 compile with named packet sections and optional files/ticket/agent_id
- REST `POST /v1/context/compile` additive JSON; MCP `memlore.get_for_task` (tool count 10)
- CLI `memlore context --task … --repository …`
- Spec: `specs/021-agent-context-bootstrap/`; F021 marked DONE

### 2026-09-01 — F020 repository intelligence profile

- Compile-on-read profile from current lore + graph; cue-based sections
- REST `POST /v1/repository-profile`, MCP `memlore.repo_profile`, CLI `memlore profile`
- Spec: `specs/020-repo-intelligence-profile/`; F020 marked DONE

### 2026-09-01 — Product epic roadmap

- Foundation (F001–F010, F101–F114, F003, F113) treated as DONE v0.8.0
- Tracker reorganized around epics A–J and the capture → trust → retrieve →
  use → drift → learn flywheel
- Next specify: **F020** Repository Intelligence Profile
- ID remaps: F115–F119, F122 (see Clarifications)
- P0 human queues: CLI + REST; web UI is F120 (P2)

### 2026-09-01 — F006 fuller semantic search + graph retrieval

- Query-relevant governance; scope-less membership-aware search
- Prefer governance + `graph_receipt` collapse
- Contract: `specs/019-semantic-graph-retrieval/`; F006 marked DONE

### Immediate recommended tasks

1. Specify **F032** (ADR ingest) or **F035** (suggested-lore review queue)
2. Or specify **F022/F023** compiler profiles/budget if staying on agent briefing
3. Then remaining Epic C ingest after F032/F035 as sequenced
4. Dogfood OIDC-on with HMAC or IdP JWKS (ops, not a product epic)
5. Optional: Postgres `pg_trgm` / FTS upgrade for governance relevance at scale

---

## Related Documents

- [Constitution](../../.specify/memory/constitution.md)
- [Target architecture](../architecture/target-architecture.md)
- [Architecture overview](../architecture/overview.md)
- [ADRs](../adr/README.md)
- [MIGRATION_DISCOVERY.md](MIGRATION_DISCOVERY.md) (historical Go strangler)
- [migration-inventory.md](migration-inventory.md)
