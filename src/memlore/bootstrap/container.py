from __future__ import annotations

from collections.abc import Iterator
from contextlib import contextmanager
from dataclasses import dataclass
from typing import Protocol, cast

from sqlalchemy.orm import Session, sessionmaker

from memlore.application.commands.create_lore import CreateLoreHandler
from memlore.application.commands.verify_lore import VerifyLoreHandler
from memlore.application.ports.clock import Clock
from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.application.queries.get_lore import GetLoreHandler
from memlore.application.queries.list_audits import ListAuditsHandler
from memlore.application.queries.list_lore_by_scope import ListLoreByScopeHandler
from memlore.bootstrap.settings import Settings, get_settings
from memlore.infrastructure.clock import SystemClock
from memlore.infrastructure.postgres.session import (
    create_db_engine,
    create_session_factory,
)
from memlore.infrastructure.postgres.unit_of_work import SqlAlchemyUnitOfWork


class UnitOfWorkFactory(Protocol):
    @contextmanager
    def __call__(self) -> Iterator[UnitOfWork]: ...


@dataclass
class AppContainer:
    settings: Settings
    clock: Clock
    _uow_factory: UnitOfWorkFactory
    session_factory: sessionmaker[Session] | None = None

    @contextmanager
    def unit_of_work(self) -> Iterator[UnitOfWork]:
        with self._uow_factory() as uow:
            yield uow

    def create_lore_handler(self, uow: UnitOfWork) -> CreateLoreHandler:
        return CreateLoreHandler(uow=uow, clock=self.clock)

    def verify_lore_handler(self, uow: UnitOfWork) -> VerifyLoreHandler:
        return VerifyLoreHandler(uow=uow, clock=self.clock)

    def get_lore_handler(self, uow: UnitOfWork) -> GetLoreHandler:
        return GetLoreHandler(uow=uow)

    def list_lore_handler(self, uow: UnitOfWork) -> ListLoreByScopeHandler:
        return ListLoreByScopeHandler(uow=uow)

    def list_audits_handler(self, uow: UnitOfWork) -> ListAuditsHandler:
        return ListAuditsHandler(uow=uow)


@dataclass
class _SqlUowFactory:
    session_factory: sessionmaker[Session]

    @contextmanager
    def __call__(self) -> Iterator[UnitOfWork]:
        session = self.session_factory()
        uow = cast(UnitOfWork, SqlAlchemyUnitOfWork(session=session))
        try:
            yield uow
        except Exception:
            uow.rollback()
            raise
        finally:
            session.close()


@dataclass
class _MemoryUowFactory:
    """Shared in-memory UoW so HTTP requests see the same store."""

    uow: UnitOfWork

    @contextmanager
    def __call__(self) -> Iterator[UnitOfWork]:
        yield self.uow


def build_container(
    settings: Settings | None = None,
    session_factory: sessionmaker[Session] | None = None,
    clock: Clock | None = None,
    uow_factory: UnitOfWorkFactory | None = None,
) -> AppContainer:
    resolved = settings or get_settings()
    factory = session_factory
    if uow_factory is None:
        if factory is None:
            engine = create_db_engine(resolved.memlore_postgres_dsn)
            factory = create_session_factory(engine)
        uow_factory = _SqlUowFactory(session_factory=factory)
    return AppContainer(
        settings=resolved,
        clock=clock or SystemClock(),
        _uow_factory=uow_factory,
        session_factory=factory,
    )


def build_memory_container(
    uow: UnitOfWork,
    clock: Clock | None = None,
) -> AppContainer:
    return AppContainer(
        settings=get_settings(),
        clock=clock or SystemClock(),
        _uow_factory=_MemoryUowFactory(uow=uow),
    )
