# Provenance

Provenance answers: who supplied this knowledge, from what source, with what
evidence, and when.

MemLore MUST retain provenance for important knowledge and must not silently
promote agent inference to human-authoritative truth.

For the first slice, create and verify writes append-only audit records
(`create`, `verify`) queryable by lore entry id. Re-verify is a no-op that does
not append a second verify audit.
