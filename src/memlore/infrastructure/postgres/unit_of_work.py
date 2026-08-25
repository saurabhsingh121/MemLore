from __future__ import annotations

from dataclasses import dataclass

from sqlalchemy.orm import Session

from memlore.infrastructure.postgres.audit_repository import SqlAlchemyAuditRepository
from memlore.infrastructure.postgres.lore_repository import SqlAlchemyLoreRepository


@dataclass
class SqlAlchemyUnitOfWork:
    session: Session

    def __post_init__(self) -> None:
        self.lore_entries = SqlAlchemyLoreRepository(self.session)
        self.audits = SqlAlchemyAuditRepository(self.session)

    def commit(self) -> None:
        self.session.commit()

    def rollback(self) -> None:
        self.session.rollback()
