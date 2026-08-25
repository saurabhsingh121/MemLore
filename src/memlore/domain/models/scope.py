from __future__ import annotations

from dataclasses import dataclass

from memlore.domain.exceptions import ValidationError
from memlore.domain.models.enums import ScopeKind

MAX_SCOPE_KEY_LENGTH = 512


@dataclass(frozen=True, slots=True)
class Scope:
    kind: ScopeKind
    key: str

    def __post_init__(self) -> None:
        key = self.key.strip()
        if not key:
            raise ValidationError("scope key must be non-empty")
        if len(key) > MAX_SCOPE_KEY_LENGTH:
            raise ValidationError(
                f"scope key must be at most {MAX_SCOPE_KEY_LENGTH} characters"
            )
        object.__setattr__(self, "key", key)
