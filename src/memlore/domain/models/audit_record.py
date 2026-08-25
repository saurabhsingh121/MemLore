from __future__ import annotations

from dataclasses import dataclass, field
from datetime import UTC, datetime
from uuid import uuid4

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import AuditAction


@dataclass(frozen=True, slots=True)
class AuditRecord:
    target_id: str
    action: AuditAction
    actor_id: str
    id: str = field(default_factory=lambda: str(uuid4()))
    created_at: datetime = field(default_factory=lambda: datetime.now(UTC))

    def __post_init__(self) -> None:
        actor = self.actor_id.strip()
        target = self.target_id.strip()
        if not actor:
            raise ValidationError("actor_id must be non-empty")
        if not target:
            raise ValidationError("target_id must be non-empty")
        object.__setattr__(self, "actor_id", actor)
        object.__setattr__(self, "target_id", target)
