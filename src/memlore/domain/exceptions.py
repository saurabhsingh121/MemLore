from __future__ import annotations


class MemloreError(Exception):
    """Base domain/application error."""


class ValidationError(MemloreError):
    """Input failed domain validation rules."""


class NotFoundError(MemloreError):
    """Requested entity does not exist."""
