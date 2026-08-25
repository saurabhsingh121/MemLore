from __future__ import annotations

from dataclasses import dataclass

from memlore.application.ports.clock import Clock
from memlore.application.ports.unit_of_work import UnitOfWork
from memlore.domain.exceptions import ValidationError
from memlore.domain.models.audit_record import AuditRecord
from memlore.domain.models.enums import AuditAction, KnowledgeOrigin
from memlore.domain.models.evidence import EvidenceReference
from memlore.domain.models.lore_entry import LoreEntry
from memlore.domain.models.scope import Scope


@dataclass(frozen=True, slots=True)
class CreateLoreCommand:
    statement: str
    scope: Scope
    actor_id: str
    evidence: tuple[EvidenceReference, ...] = ()


class CreateLoreHandler:
    def __init__(self, uow: UnitOfWork, clock: Clock) -> None:
        self._uow = uow
        self._clock = clock

    def handle(self, command: CreateLoreCommand) -> LoreEntry:
        actor = command.actor_id.strip()
        if not actor:
            raise ValidationError("actor must be non-empty")

        now = self._clock.now()
        entry = LoreEntry(
            statement=command.statement,
            scope=command.scope,
            created_by=actor,
            origin=KnowledgeOrigin.HUMAN_AUTHORED,
            evidence=list(command.evidence),
            created_at=now,
            updated_at=now,
        )
        audit = AuditRecord(
            target_id=entry.id,
            action=AuditAction.CREATE,
            actor_id=actor,
            created_at=now,
        )
        self._uow.lore_entries.add(entry)
        self._uow.audits.add(audit)
        self._uow.commit()
        return entry
