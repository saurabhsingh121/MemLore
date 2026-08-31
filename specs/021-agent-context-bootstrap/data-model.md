# Data Model: Agent Context Bootstrap (F021)

No new persistence. On-read types only. Extends F007 ContextPacket.

## PacketSectionID

Display order for non-empty packet sections:

1. `architecture` — Relevant Architecture (F020 id)
2. `decisions` — Applicable Decisions (F020 id)
3. `conventions` — Coding Conventions (F020 id)
4. `task_context` — Task Context (new)
5. `gotchas` — Known Gotchas (F020 id)

F020 ids `migrations`, `ownership`, `operational_risks`, `hotspots`,
`related_services`, `technologies`, `recent_changes` are **not** packet
section keys. Items classified as those ids go to `task_context` when
task-relevant; otherwise they count as unclassified for packet sections.

Conflicts are **not** a classified section; they remain `conflicts[]`.
Drift/stale section ids are not emitted.

## Task relevance

An item is task-relevant when its statement or any evidence value contains
(case-insensitive) any of: task, query, ticket, a changed-file path, a
working-file path — after trimming; empty signals ignored.

When the only signal is `task` (no extra query/ticket/files), leftover items
that are not briefing-classified (`architecture`/`decisions`/`conventions`/`gotchas`)
and that came from the retrieval set still qualify for `task_context` if they
match the task string; if they do not match, they stay unclassified (in
`items` if budgeted, not in sections).

## CompileContextQuery (extended)

| Field | Required | Notes |
|-------|----------|-------|
| task | yes | Non-empty |
| scope | yes | `{ kind, key }` as today |
| query | no | Defaults to task |
| token_budget | no | ≤ 0 → 4096 |
| branch | no | Echo only; no filter |
| ticket | no | Appended to search text |
| changed_files | no | Path strings; appended to search text |
| working_files | no | Path strings; appended to search text |
| agent_id | no | Echo only; never ranking |

## ContextPacket (extended)

Existing: `task`, `query`, `scope`, `items`, `meta`, `warnings`, `conflicts`.

Additive:

| Field | Meaning |
|-------|---------|
| branch, ticket, changed_files, working_files, agent_id | Echo when provided |
| sections | `{ id, items[] }` — only non-empty packet sections |
| sources | unique `{ type, value }` from included item evidence; omit if none |
| meta.unclassified_count | budgeted items not placed in any packet section |

## Validation

- Empty optional strings/lists ≡ omitted
- `agent_id` MUST NOT change `authority_score` or `authority_factors`
- No new tables or packet ids persisted
