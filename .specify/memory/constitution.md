<!--
Sync Impact Report
- Version change: 1.1.0 → 1.2.0 (MINOR)
- Modified principles:
  - V. Authority and Provenance Are First-Class — expanded: extracted
    candidates MUST NOT silently become canonical; usage/popularity MUST NOT
    override authority
  - VI. Temporal Correctness — expanded: intent vs implementation drift MUST
    be surfaced, never silently discarded or auto-resolved as truth
  - VIII. Observability — examples include ingest, drift, and review queues
- Added sections:
  - IX. Engineering Intelligence Over Generic Memory (new principle)
- Modified sections:
  - Architecture & Technology Baseline: CLI is a first-class developer
    adapter; web UI is not required for feature completeness
  - Development Workflow item 3: contract tests include CLI product surfaces
- Removed sections: n/a
- Templates requiring updates:
  - .specify/templates/plan-template.md - ✅ Constitution Check for IX +
    candidate-lore / no-popularity-override / drift
  - .specify/templates/tasks-template.md - ✅ no structural change
  - .specify/templates/spec-template.md - ✅ no structural change
  - .specify/templates/checklist-template.md - ✅ no structural change
- Follow-up TODOs: none
- Rationale: Product roadmap (2026-09-01) makes engineering intelligence,
  candidate-to-trusted promotion, and drift surfacing non-negotiable. These
  are governance rules, not feature IDs. GitHub-first sequencing and F120 web
  UI deferral stay in FEATURE_DEVELOPMENT.md, not here.
-->

# MemLore Constitution

## Core Principles

### I. Test-Driven Development

All behavioral production code MUST be developed using RED → GREEN → REFACTOR.

- **RED**: Write a failing test that demonstrates the required behavior. Confirm
  it fails for the expected reason.
- **GREEN**: Write the minimum production implementation necessary to make the
  test pass. Do not solve unrelated problems.
- **REFACTOR**: Improve names, duplication, abstractions, boundaries,
  readability, and structure while keeping all tests green.

Never create production behavior first and retroactively write tests merely to
claim TDD compliance. Defect fixes MUST begin with a failing regression test
unless technically impossible; if impossible, document why.

### II. Specification-Driven Development

Every material feature MUST begin from a Spec Kit specification describing
**what** and **why**. Technical implementation details belong primarily in the
plan, not the feature specification.

Specifications MUST contain measurable acceptance criteria. Ambiguous
requirements that affect product behavior or architecture MUST be clarified
before irreversible choices are made. Spec Kit lifecycle
(constitution → specify → clarify → checklist → plan → tasks → analyze →
implement → converge) is mandatory for substantial features.

### III. Architecture Integrity

Maintain clear boundaries between domain, application, ports/interfaces,
infrastructure, API/MCP adapters, persistence, and external integrations.

- Domain logic MUST NOT depend directly on FastAPI, PostgreSQL, Neo4j,
  Graphiti, Redis, or other infrastructure frameworks.
- Prefer dependency inversion: infrastructure implements ports defined nearer
  the application/domain boundary.
- Transport models and persistence models MUST NOT become domain models by
  accident.
- PostgreSQL is authoritative for governance (users, teams, scopes, authority,
  verification, audit, outbox). Graphiti/Neo4j own semantic/temporal knowledge.
  Do not require distributed transactions between them; synchronize via
  transactional outbox or equivalent.

### IV. Documentation Is Part of the Product

Code changes and documentation changes belong to the same unit of work. A
feature that changes system behavior but leaves documentation stale is
incomplete. Prefer ADRs for irreversible or cross-cutting decisions. Do not
create empty documentation files merely to populate the tree.

### V. Authority and Provenance Are First-Class

MemLore MUST preserve who supplied knowledge, whether the source was human or
agent, evidence, source type, verification status, validity interval,
supersession relationships, authority factors, and audit history.

Never silently convert an inference or an automatically extracted candidate
into an authoritative fact. Agent observations, agent inferences, git-derived
observations, and pull-request extracts MUST remain distinguishable from
human-authored or human-verified knowledge until a human accepts them or an
explicit trusted-source policy in a specification applies (for example
accepted ADRs). Trusted-source ingest MUST still preserve evidence and remain
auditable.

Authority is computed from explicit factors, not a magical opaque scalar;
factors MUST remain explainable. Retrieval usage, popularity, and feedback
signals MUST NOT override authority evaluation. They MAY inform ranking only
as secondary, explainable factors.

### VI. Temporal Correctness

Knowledge may change over time. The system MUST distinguish currently true,
previously true, superseded, invalidated, conflicting, and unverified states.
Do not overwrite historical truth when a fact changes. Conflicts MUST be
preserved and surfaced, never silently discarded.

Authoritative engineering intent versus observed implementation (drift) MUST
be surfaced as conflict or drift. Drift MUST NOT silently discard either side,
and MUST NOT automatically rewrite or invalidate the authoritative decision.

### VII. Secure by Default

Apply least privilege, explicit authorization, secure secret handling, input
validation, dependency hygiene, auditability, and tenant/team isolation.

Never log secrets, credentials, authorization tokens, or sensitive raw prompts
unnecessarily. Treat stored agent context as untrusted input. Never assume
internal coding-agent content is safe.

### VIII. Observability

Important system operations MUST be observable through appropriate structured
logs, metrics, distributed traces, health checks, and error reporting.

Do not add telemetry blindly. Instrument operations that help understand
correctness, latency, failures, and capacity (e.g., context retrieval,
ingestion, outbox processing, authority evaluation, conflict and drift
detection, candidate review queues, context compilation, MCP operations).
Include correlation fields such as request_id, trace_id, actor_id, agent_id,
team_id, repository_id, and operation where applicable.

### IX. Engineering Intelligence Over Generic Memory

MemLore is an engineering intelligence layer, not a generic memory store.
It MUST help humans and coding agents understand what the team decided, why,
whether that intent is still valid, what implements it, what contradicts it,
and what is needed for the current task.

MemLore MUST NOT optimize for undifferentiated memory storage, embedding
dumps, or general-purpose document search. A feature that adds memory
primitives without improving trust, evidence, retrieval usefulness for a
real engineering workflow, or implementation awareness requires explicit
justification in its specification.

## Architecture & Technology Baseline

Unless an accepted ADR changes it, MemLore uses:

- **MemLore Core**: Go 1.25+, chi, pgx, sqlc, goose, slog, Go MCP SDK
  ([ADR-0005](../../docs/adr/0005-go-memlore-core.md))
- **Graph knowledge service**: Python 3.12, FastAPI, Graphiti (`graph-service/`)
- PostgreSQL (governance/control plane)
- Graphiti + Neo4j (temporal knowledge plane)
- MCP as primary agent-facing integration; REST for automation and
  integrations; CLI for developer workflows (review, explain, `why`)
- A product web UI is not required for a feature to be complete; CLI + REST
  (and MCP where agents participate) are sufficient governance surfaces
- Redis where justified (e.g., background workers)
- OpenTelemetry; Docker Compose for local development
- Tests: `go test` for core; pytest for graph-service
- Spec Kit for specification-driven development

Canonical public brand for this repository is **MemLore**. Package, CLI, and
MCP namespaces use `memlore`.

Preferred package layout:

```text
cmd/memlore/
internal/{domain,application,adapters,infrastructure,bootstrap}
migrations/                 # goose
db/queries/                 # sqlc
graph-service/              # Python Graphiti boundary
```

## Development Workflow

1. Prefer small, independently verifiable vertical increments.
2. Behavioral tasks MUST include RED/GREEN/REFACTOR and a DONE WHEN checklist.
3. Use the test pyramid: many unit tests, fewer integration tests, only
   necessary e2e tests. Integration tests where infrastructure behavior
   matters; contract tests for MCP/REST and for CLI product surfaces.
4. All PostgreSQL schema changes MUST use goose migrations (`migrations/`).
5. Do not add third-party dependencies without clear purpose, license, and
   security consideration.
6. Before marking work complete, update docs if architecture, public API, MCP,
   data model, configuration, deployment, developer workflow, user behavior,
   security, or operations changed.
7. A feature is DONE only when specification acceptance criteria, tests,
   lint/format/type checks, relevant migrations, security/observability
   consideration, and documentation converge with Spec Kit artifacts.

## Governance

This constitution is the highest-level engineering contract for the codebase.
It supersedes informal practice and agent convenience.

- Amendments MUST update this file, bump the version (MAJOR for removals or
  incompatible redefinitions; MINOR for new/expanded principles; PATCH for
  clarifications), and record a Sync Impact Report.
- Accepted ADRs MUST NOT be rewritten to erase history; supersede them with a
  new ADR instead.
- All PRs and agent completion reports MUST be reviewable against these
  principles. Unjustified complexity is a defect.
- Runtime development guidance lives in `README.md` and `docs/`.
- Product sequencing (which feature ships first, which forge is first, when a
  web UI exists) lives in `docs/development/FEATURE_DEVELOPMENT.md` and MUST
  not contradict these principles.

**Version**: 1.2.0 | **Ratified**: 2026-08-25 | **Last Amended**: 2026-09-01
