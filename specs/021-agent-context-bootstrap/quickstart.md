# Quickstart: Agent Context Bootstrap (F021)

1. Create current lore in a repository: architecture wording, ADR-backed
   decision, a convention, a gotcha, and a task-specific outbox statement.
2. `POST /v1/context/compile` with only `{ task, scope }` — expect `200` and
   `items` populated (v1 body still works).
3. Compile again with `changed_files` / `ticket` matching the outbox
   statement; expect `sections` to include `task_context` and briefing ids
   when those items classified; empty ids omitted.
4. `memlore.get_for_task` with the same inputs; JSON matches REST (same
   item ids and section membership). MCP tool list still has 10 tools.
5. `memlore context --task "…" --repository <key>` prints section headings.
6. Repeat compile with a different `agent_id`; authority scores unchanged.
7. Supersede one entry; predecessor must not appear in `items` or sections.
8. `go test ./...` and `go vet ./...` green.
