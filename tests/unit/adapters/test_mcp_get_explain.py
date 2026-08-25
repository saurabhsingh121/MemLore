from __future__ import annotations

from datetime import UTC, datetime

import pytest
from mcp.server.mcpserver.exceptions import ToolError
from tests.support.fakes import InMemoryUnitOfWork

from memlore.adapters.mcp.tools import explain_lore, get_lore, remember_lore
from memlore.adapters.rest.schemas import ScopeIn
from memlore.bootstrap.container import build_memory_container
from memlore.domain.models.enums import ScopeKind
from memlore.infrastructure.clock import FixedClock


def _container():
    uow = InMemoryUnitOfWork()
    return uow, build_memory_container(
        uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )


def test_get_and_explain_unknown_id() -> None:
    _, container = _container()
    missing = "00000000-0000-0000-0000-000000000000"
    with pytest.raises(ToolError, match="not_found:"):
        get_lore(container, id=missing)
    with pytest.raises(ToolError, match="not_found:"):
        explain_lore(container, id=missing)


def test_explain_includes_entry_fields_and_chronological_audits() -> None:
    _, container = _container()
    created = remember_lore(
        container,
        statement="Prefer outbox",
        scope=ScopeIn(kind=ScopeKind.REPOSITORY, key="r1"),
        actor_id="alice",
    )
    explained = explain_lore(container, id=created.id)
    assert explained.statement == created.statement
    assert explained.origin == created.origin
    assert "summary" not in explained.model_dump()
    assert [a.action.value for a in explained.audits] == ["create"]
    got = get_lore(container, id=created.id)
    assert got.id == created.id
    assert got.created_by == "alice"
