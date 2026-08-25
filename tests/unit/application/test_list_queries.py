from __future__ import annotations

from datetime import UTC, datetime

import pytest

from memlore.application.commands.create_lore import (
    CreateLoreCommand,
    CreateLoreHandler,
)
from memlore.application.queries.get_lore import GetLoreHandler
from memlore.application.queries.list_audits import ListAuditsHandler
from memlore.application.queries.list_lore_by_scope import ListLoreByScopeHandler
from memlore.domain.exceptions import NotFoundError
from memlore.domain.models.enums import ScopeKind
from memlore.domain.models.scope import Scope
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def test_list_by_scope_filters_exact_kind_key() -> None:
    uow = InMemoryUnitOfWork()
    handler = CreateLoreHandler(
        uow=uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )
    handler.handle(
        CreateLoreCommand(
            statement="A",
            scope=Scope(kind=ScopeKind.REPOSITORY, key="r1"),
            actor_id="alice",
        )
    )
    handler.handle(
        CreateLoreCommand(
            statement="B",
            scope=Scope(kind=ScopeKind.TEAM, key="r1"),
            actor_id="alice",
        )
    )
    items = ListLoreByScopeHandler(uow).handle(
        Scope(kind=ScopeKind.REPOSITORY, key="r1")
    )
    assert len(items) == 1
    assert items[0].statement == "A"


def test_list_audits_requires_existing_entry() -> None:
    uow = InMemoryUnitOfWork()
    with pytest.raises(NotFoundError):
        ListAuditsHandler(uow).handle("missing")


def test_get_missing_raises() -> None:
    with pytest.raises(NotFoundError):
        GetLoreHandler(InMemoryUnitOfWork()).handle("missing")
