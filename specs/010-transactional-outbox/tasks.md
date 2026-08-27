# Tasks: Transactional Outbox (F107)

- [x] T001 Spec + migration `00002_outbox_events.sql`
- [x] T002 Domain `OutboxEvent` + ports + postgres/memory repos
- [x] T003 `CreateLore` emits outbox in same transaction
- [x] T004 Worker processor + `memlore worker` command
- [x] T005 Unit tests (create + worker + idempotency)
- [x] T006 Postgres integration test (skip if unavailable)
- [x] T007 Docs + FEATURE_DEVELOPMENT (F004/F107 DONE)
