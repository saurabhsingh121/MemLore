# Contributing

1. Read the [constitution](../../.specify/memory/constitution.md).
2. Use Spec Kit for material features.
3. Follow TDD for behavioral changes.
4. Keep PRs small and independently verifiable.
5. Update docs/ADRs in the same change when behavior or architecture shifts.
6. Ensure `go test ./...` and `go vet ./...` pass before review.
   Graph-service changes also need `uv run ruff` / `mypy` / `pytest` in
   `graph-service/`.
