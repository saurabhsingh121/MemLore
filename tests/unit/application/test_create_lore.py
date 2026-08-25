from __future__ import annotations

from datetime import UTC, datetime

import pytest

from memlore.application.commands.create_lore import (
    CreateLoreCommand,
    CreateLoreHandler,
)
from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import AuditAction, ScopeKind
from memlore.domain.models.lore_entry import MAX_STATEMENT_LENGTH
from memlore.domain.models.scope import Scope
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def test_create_lore_persists_entry_and_create_audit() -> None:
    uow = InMemoryUnitOfWork()
    clock = FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    handler = CreateLoreHandler(uow=uow, clock=clock)

    entry = handler.handle(
        CreateLoreCommand(
            statement="Use the outbox",
            scope=Scope(kind=ScopeKind.REPOSITORY, key="r1"),
            actor_id="alice",
        )
    )

    assert uow.committed is True
    assert uow.lore_entries.get(entry.id) is not None
    audits = uow.audits.list_by_target(entry.id)
    assert len(audits) == 1
    assert audits[0].action == AuditAction.CREATE
    assert audits[0].actor_id == "alice"


def test_create_lore_allows_duplicate_statement_in_same_scope() -> None:
    uow = InMemoryUnitOfWork()
    handler = CreateLoreHandler(
        uow=uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )
    cmd = CreateLoreCommand(
        statement="Same",
        scope=Scope(kind=ScopeKind.TEAM, key="t1"),
        actor_id="alice",
    )
    a = handler.handle(cmd)
    b = handler.handle(cmd)
    assert a.id != b.id


def test_create_lore_rejects_oversized_statement() -> None:
    uow = InMemoryUnitOfWork()
    handler = CreateLoreHandler(
        uow=uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )
    with pytest.raises(ValidationError):
        handler.handle(
            CreateLoreCommand(
                statement="x" * (MAX_STATEMENT_LENGTH + 1),
                scope=Scope(kind=ScopeKind.TEAM, key="t1"),
                actor_id="alice",
            )
        )
    assert uow.lore_entries.list_by_scope(Scope(kind=ScopeKind.TEAM, key="t1")) == []
