from __future__ import annotations

from dataclasses import dataclass

from memlore.application.ports.clock import Clock
from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.domain.exceptions import NotFoundError, ValidationError
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.services.verification import apply_verification


@dataclass(frozen=True, slots=True)
class VerifyLoreCommand:
    entry_id: str
    actor_id: str


class VerifyLoreHandler:
    def __init__(self, uow: UnitOfWork, clock: Clock) -> None:
        self._uow = uow
        self._clock = clock

    def handle(self, command: VerifyLoreCommand) -> LoreEntry:
        actor = command.actor_id.strip()
        if not actor:
            raise ValidationError("actor must be non-empty")

        entry = self._uow.lore_entries.get(command.entry_id)
        if entry is None:
            raise NotFoundError(f"lore entry {command.entry_id} not found")

        updated, audit = apply_verification(
            entry, actor_id=actor, now=self._clock.now()
        )
        self._uow.lore_entries.save(updated)
        if audit is not None:
            self._uow.audits.add(audit)
        self._uow.commit()
        return updated
