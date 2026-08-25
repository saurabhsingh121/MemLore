from __future__ import annotations

from dataclasses import dataclass

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import EvidenceType

MAX_EVIDENCE_VALUE_LENGTH = 2048


@dataclass(frozen=True, slots=True)
class EvidenceReference:
    type: EvidenceType
    value: str

    def __post_init__(self) -> None:
        value = self.value.strip()
        if not value:
            raise ValidationError("evidence value must be non-empty")
        if len(value) > MAX_EVIDENCE_VALUE_LENGTH:
            raise ValidationError(
                f"evidence value must be at most {MAX_EVIDENCE_VALUE_LENGTH} characters"
            )
        object.__setattr__(self, "value", value)
