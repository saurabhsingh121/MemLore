from __future__ import annotations

import os
from datetime import UTC, datetime

import pytest
from mcp.server.mcpserver.exceptions import ToolError
from tests.support.fakes import InMemoryUnitOfWork

from memlore.adapters.mcp.tools import remember_lore
from memlore.adapters.rest.schemas import EvidenceIn, ScopeIn
from memlore.bootstrap.container import build_memory_container
from memlore.domain.models.enums import KnowledgeOrigin, ScopeKind
from memlore.domain.models.scope import Scope
from memlore.infrastructure.clock import FixedClock


def _container(uow: InMemoryUnitOfWork | None = None):
    store = uow or InMemoryUnitOfWork()
    return store, build_memory_container(
        store, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )


def test_remember_requires_non_empty_actor_and_ignores_env(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("MEMLORE_ACTOR", "env-actor")
    uow, container = _container()
    with pytest.raises(ToolError, match="validation_error:"):
        remember_lore(
            container,
            statement="Use the outbox",
            scope=ScopeIn(kind=ScopeKind.REPOSITORY, key="r1"),
            actor_id="   ",
        )
    assert (
        uow.lore_entries.list_by_scope(Scope(kind=ScopeKind.REPOSITORY, key="r1")) == []
    )
    assert os.environ.get("MEMLORE_ACTOR") == "env-actor"


def test_remember_stores_human_authored_and_allows_duplicates() -> None:
    uow, container = _container()
    scope = ScopeIn(kind=ScopeKind.TEAM, key="payments")
    first = remember_lore(
        container,
        statement="Prefer explicit actor_id",
        scope=scope,
        actor_id="alice",
        evidence=[EvidenceIn(type="adr", value="0003-mcp")],
    )
    second = remember_lore(
        container,
        statement="Prefer explicit actor_id",
        scope=scope,
        actor_id="alice",
    )
    assert first.origin == KnowledgeOrigin.HUMAN_AUTHORED
    assert first.id != second.id
    assert uow.committed is True
    listed = uow.lore_entries.list_by_scope(Scope(kind=ScopeKind.TEAM, key="payments"))
    assert len(listed) == 2


def test_remember_invalid_statement_stores_nothing() -> None:
    uow, container = _container()
    with pytest.raises(ToolError, match="validation_error:"):
        remember_lore(
            container,
            statement="   ",
            scope=ScopeIn(kind=ScopeKind.TEAM, key="t1"),
            actor_id="alice",
        )
    assert uow.lore_entries._items == {}
