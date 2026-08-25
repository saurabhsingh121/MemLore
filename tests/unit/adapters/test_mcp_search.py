from __future__ import annotations

from datetime import UTC, datetime

import pytest
from mcp.server.mcpserver.exceptions import ToolError
from tests.support.fakes import InMemoryUnitOfWork

from memlore.adapters.mcp.tools import remember_lore, search_lore
from memlore.adapters.rest.schemas import ScopeIn
from memlore.bootstrap.container import build_memory_container
from memlore.domain.models.enums import ScopeKind
from memlore.infrastructure.clock import FixedClock


def _container():
    uow = InMemoryUnitOfWork()
    return uow, build_memory_container(
        uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )


def test_search_filters_exact_scope_and_empty_list() -> None:
    _, container = _container()
    remember_lore(
        container,
        statement="A",
        scope=ScopeIn(kind=ScopeKind.REPOSITORY, key="alpha"),
        actor_id="alice",
    )
    remember_lore(
        container,
        statement="B",
        scope=ScopeIn(kind=ScopeKind.REPOSITORY, key="beta"),
        actor_id="alice",
    )
    found = search_lore(
        container, scope=ScopeIn(kind=ScopeKind.REPOSITORY, key="alpha")
    )
    assert [item.statement for item in found.items] == ["A"]
    empty = search_lore(container, scope=ScopeIn(kind=ScopeKind.TEAM, key="nobody"))
    assert empty.items == []


def test_search_incomplete_scope_is_validation_error() -> None:
    _, container = _container()
    with pytest.raises(ToolError, match="validation_error:"):
        search_lore(container, scope=None)
