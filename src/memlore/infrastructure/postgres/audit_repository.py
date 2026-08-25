from __future__ import annotations

from sqlalchemy import select
from sqlalchemy.orm import Session

from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.enums import AuditAction
from memlore.infrastructure.postgres.models import AuditRecordRow


class SqlAlchemyAuditRepository:
    def __init__(self, session: Session) -> None:
        self._session = session

    def add(self, record: AuditRecord) -> None:
        self._session.add(
            AuditRecordRow(
                id=record.id,
                target_id=record.target_id,
                action=record.action.value,
                actor_id=record.actor_id,
                created_at=record.created_at,
            )
        )

    def list_by_target(self, target_id: str) -> list[AuditRecord]:
        stmt = (
            select(AuditRecordRow)
            .where(AuditRecordRow.target_id == target_id)
            .order_by(AuditRecordRow.created_at.asc(), AuditRecordRow.id.asc())
        )
        return [
            AuditRecord(
                id=row.id,
                target_id=row.target_id,
                action=AuditAction(row.action),
                actor_id=row.actor_id,
                created_at=row.created_at,
            )
            for row in self._session.scalars(stmt)
        ]
