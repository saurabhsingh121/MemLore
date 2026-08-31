# Quickstart: Repository Intelligence Profile (F020)

1. Create current lore in a repository scope (ADR evidence for a decision,
   architecture wording for architecture).
2. `POST /v1/repository-profile` with that scope; expect those sections only.
3. `memlore.repo_profile` with the same scope; JSON matches REST.
4. `memlore profile --repository <key>` prints section headings.
5. Supersede one entry; profile sections must not include the predecessor.
6. `go test ./...` green.
