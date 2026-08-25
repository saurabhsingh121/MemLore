from __future__ import annotations

from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.domain.exceptions import NotFoundError
from memlore.domain.models.lore_entry import LoreEntry


class GetLoreHandler:
    def __init__(self, uow: UnitOfWork) -> None:
        self._uow = uow

    def handle(self, entry_id: str) -> LoreEntry:
        entry = self._uow.lore_entries.get(entry_id)
        if entry is None:
            raise NotFoundError(f"lore entry {entry_id} not found")
        return entry
