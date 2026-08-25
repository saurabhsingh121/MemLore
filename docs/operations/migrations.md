# Migrations

PostgreSQL schema changes MUST use Alembic migrations. Migrations must be
reviewable, deterministic, and safe. Prefer verifying upgrade and downgrade
paths.

Alembic is not wired in bootstrap; it will be introduced with the first
governance-plane persistence feature.
