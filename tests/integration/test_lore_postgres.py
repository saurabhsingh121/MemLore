from __future__ import annotations

import pytest
from sqlalchemy import text

from memlore.application.commands.create_lore import (
    CreateLoreCommand,
    CreateLoreHandler,
)
from memlore.application.commands.verify_lore import (
    VerifyLoreCommand,
    VerifyLoreHandler,
)
from memlore.application.queries.list_audits import ListAuditsHandler
from memlore.application.queries.list_lore_by_scope import ListLoreByScopeHandler
from memlore.domain.models.enums import AuditAction, ScopeKind, VerificationStatus
from memlore.domain.models.scope import Scope
from memlore.infrastructure.clock import SystemClock
from memlore.infrastructure.postgres.models import Base
from memlore.infrastructure.postgres.session import (
    create_db_engine,
    create_session_factory,
)
from memlore.infrastructure.postgres.unit_of_work import SqlAlchemyUnitOfWork


@pytest.fixture
def pg_uow(require_postgres: None, database_url: str):
    engine = create_db_engine(database_url)
    Base.metadata.drop_all(engine)
    Base.metadata.create_all(engine)
    factory = create_session_factory(engine)
    session = factory()
    uow = SqlAlchemyUnitOfWork(session=session)
    yield uow
    session.close()
    Base.metadata.drop_all(engine)
    engine.dispose()


@pytest.mark.integration
def test_postgres_create_verify_list_audits(pg_uow: SqlAlchemyUnitOfWork) -> None:
    clock = SystemClock()
    create = CreateLoreHandler(uow=pg_uow, clock=clock)
    entry = create.handle(
        CreateLoreCommand(
            statement="Outbox required",
            scope=Scope(kind=ScopeKind.REPOSITORY, key="github.com/acme/app"),
            actor_id="alice",
        )
    )
    verify = VerifyLoreHandler(uow=pg_uow, clock=clock)
    verified = verify.handle(VerifyLoreCommand(entry_id=entry.id, actor_id="alice"))
    assert verified.verification_status == VerificationStatus.VERIFIED

    audits = ListAuditsHandler(pg_uow).handle(entry.id)
    assert [a.action for a in audits] == [AuditAction.CREATE, AuditAction.VERIFY]

    listed = ListLoreByScopeHandler(pg_uow).handle(
        Scope(kind=ScopeKind.REPOSITORY, key="github.com/acme/app")
    )
    assert len(listed) == 1
    assert listed[0].id == entry.id


@pytest.mark.integration
def test_postgres_connectivity_fixture(
    require_postgres: None, database_url: str
) -> None:
    engine = create_db_engine(database_url)
    with engine.connect() as conn:
        assert conn.execute(text("SELECT 1")).scalar() == 1
    engine.dispose()
