from __future__ import annotations

from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.domain.exceptions import NotFoundError
from memlore.domain.models.audit_record import AuditRecord


class ListAuditsHandler:
    def __init__(self, uow: UnitOfWork) -> None:
        self._uow = uow

    def handle(self, entry_id: str) -> list[AuditRecord]:
        if self._uow.lore_entries.get(entry_id) is None:
            raise NotFoundError(f"lore entry {entry_id} not found")
        return self._uow.audits.list_by_target(entry_id)
