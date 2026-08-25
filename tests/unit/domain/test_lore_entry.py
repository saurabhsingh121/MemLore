from __future__ import annotations

import pytest

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import (
    KnowledgeOrigin,
    ScopeKind,
    VerificationStatus,
)
from memlore.domain.models.lore_entry import MAX_STATEMENT_LENGTH, LoreEntry
from memlore.domain.models.scope import Scope


def test_lore_entry_defaults_to_unverified_human_authored() -> None:
    entry = LoreEntry(
        statement="Use the outbox",
        scope=Scope(kind=ScopeKind.REPOSITORY, key="r1"),
        created_by="alice",
    )
    assert entry.origin == KnowledgeOrigin.HUMAN_AUTHORED
    assert entry.verification_status == VerificationStatus.UNVERIFIED
    assert entry.verified_by is None
    assert entry.id


def test_lore_entry_rejects_oversized_statement() -> None:
    with pytest.raises(ValidationError):
        LoreEntry(
            statement="x" * (MAX_STATEMENT_LENGTH + 1),
            scope=Scope(kind=ScopeKind.TEAM, key="t1"),
            created_by="alice",
        )
