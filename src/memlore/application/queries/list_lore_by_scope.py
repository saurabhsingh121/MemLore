from __future__ import annotations

from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope


class ListLoreByScopeHandler:
    def __init__(self, uow: UnitOfWork) -> None:
        self._uow = uow

    def handle(self, scope: Scope) -> list[LoreEntry]:
        return self._uow.lore_entries.list_by_scope(scope)
