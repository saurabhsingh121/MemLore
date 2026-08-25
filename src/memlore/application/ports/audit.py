from __future__ import annotations

from typing import Protocol

from memlore.domain.models.audit_record import AuditRecord


class AuditRepository(Protocol):
    def add(self, record: AuditRecord) -> None: ...

    def list_by_target(self, target_id: str) -> list[AuditRecord]: ...
