from __future__ import annotations

from typing import Protocol

from memlore.application.ports.audit import AuditRepository
from memlore.application.ports.lore import LoreRepository


class UnitOfWork(Protocol):
    lore_entries: LoreRepository
    audits: AuditRepository

    def commit(self) -> None: ...

    def rollback(self) -> None: ...
