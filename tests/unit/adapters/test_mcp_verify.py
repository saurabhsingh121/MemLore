from __future__ import annotations

from datetime import UTC, datetime

import pytest
from mcp.server.mcpserver.exceptions import ToolError
from tests.support.fakes import InMemoryUnitOfWork

from memlore.adapters.mcp.tools import remember_lore, verify_lore
from memlore.adapters.rest.schemas import ScopeIn
from memlore.application.queries.list_audits import ListAuditsHandler
from memlore.bootstrap.container import build_memory_container
from memlore.domain.models.enums import AuditAction, KnowledgeOrigin, ScopeKind
from memlore.infrastructure.clock import FixedClock


def _container():
    uow = InMemoryUnitOfWork()
    return uow, build_memory_container(
        uow, clock=FixedClock(datetime(2026, 8, 25, tzinfo=UTC))
    )


def test_verify_blank_actor_and_unknown_id() -> None:
    _, container = _container()
    created = remember_lore(
        container,
        statement="Rule",
        scope=ScopeIn(kind=ScopeKind.TEAM, key="t1"),
        actor_id="alice",
    )
    with pytest.raises(ToolError, match="validation_error:"):
        verify_lore(container, id=created.id, actor_id="")
    with pytest.raises(ToolError, match="not_found:"):
        verify_lore(
            container,
            id="00000000-0000-0000-0000-000000000000",
            actor_id="alice",
        )


def test_verify_is_idempotent_and_preserves_origin() -> None:
    _uow, container = _container()
    created = remember_lore(
        container,
        statement="Rule",
        scope=ScopeIn(kind=ScopeKind.TEAM, key="t1"),
        actor_id="alice",
    )
    first = verify_lore(container, id=created.id, actor_id="alice")
    second = verify_lore(container, id=created.id, actor_id="bob")
    assert first.verification_status.value == "verified"
    assert first.origin == KnowledgeOrigin.HUMAN_AUTHORED
    assert second.verified_by == "alice"
    with container.unit_of_work() as inner:
        audits = ListAuditsHandler(inner).handle(created.id)
    actions = [a.action for a in audits]
    assert actions.count(AuditAction.CREATE) == 1
    assert actions.count(AuditAction.VERIFY) == 1
