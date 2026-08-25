from __future__ import annotations

from typing import Annotated

from fastapi import Header

from memlore.domain.exceptions import ValidationError


def require_actor(
    x_memlore_actor: Annotated[str | None, Header(alias="X-Memlore-Actor")] = None,
) -> str:
    if x_memlore_actor is None or not x_memlore_actor.strip():
        raise ValidationError("X-Memlore-Actor header is required")
    return x_memlore_actor.strip()
