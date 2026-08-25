from __future__ import annotations

from datetime import UTC, datetime

from memlore.application.commands.create_lore import (
    CreateLoreCommand,
    CreateLoreHandler,
)
from memlore.application.commands.verify_lore import (
    VerifyLoreCommand,
    VerifyLoreHandler,
)
from memlore.application.queries.get_lore import GetLoreHandler
from memlore.domain.models.enums import (
    AuditAction,
    KnowledgeOrigin,
    ScopeKind,
    VerificationStatus,
)
from memlore.domain.models.scope import Scope
from memlore.infrastructure.clock import FixedClock
from tests.support.fakes import InMemoryUnitOfWork


def _seed(uow: InMemoryUnitOfWork) -> str:
    entry = CreateLoreHandler(
        uow=uow, clock=FixedClock(datetime(2026, 8, 25, 12, 0, tzinfo=UTC))
    ).handle(
        CreateLoreCommand(
            statement="Rule",
            scope=Scope(kind=ScopeKind.REPOSITORY, key="r1"),
            actor_id="alice",
        )
    )
    return entry.id


def test_get_lore_returns_entry() -> None:
    uow = InMemoryUnitOfWork()
    entry_id = _seed(uow)
    got = GetLoreHandler(uow).handle(entry_id)
    assert got.id == entry_id
    assert got.origin == KnowledgeOrigin.HUMAN_AUTHORED


def test_verify_self_and_idempotent() -> None:
    uow = InMemoryUnitOfWork()
    entry_id = _seed(uow)
    verify = VerifyLoreHandler(
        uow=uow, clock=FixedClock(datetime(2026, 8, 25, 13, 0, tzinfo=UTC))
    )
    first = verify.handle(VerifyLoreCommand(entry_id=entry_id, actor_id="alice"))
    assert first.verification_status == VerificationStatus.VERIFIED
    assert first.verified_by == "alice"
    assert first.origin == KnowledgeOrigin.HUMAN_AUTHORED

    second = verify.handle(VerifyLoreCommand(entry_id=entry_id, actor_id="bob"))
    assert second.verified_by == "alice"
    audits = uow.audits.list_by_target(entry_id)
    assert [a.action for a in audits] == [AuditAction.CREATE, AuditAction.VERIFY]
