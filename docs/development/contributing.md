# Contributing

1. Read the [constitution](../../.specify/memory/constitution.md).
2. Pick work from the [feature roadmap](FEATURE_DEVELOPMENT.md) (next: F032 or F035; F022 is next Epic D).
3. Use Spec Kit for material features.
4. Follow TDD for behavioral changes.
5. Keep PRs small and independently verifiable.
6. Update docs/ADRs in the same change when behavior or architecture shifts.
7. Ensure `go test ./...` and `go vet ./...` pass before review.
   Graph-service changes also need `uv run ruff` / `mypy` / `pytest` in
   `graph-service/`.
