from __future__ import annotations

from datetime import UTC, datetime

from memlore.application.ports.clock import Clock


class SystemClock:
    def now(self) -> datetime:
        return datetime.now(UTC)


class FixedClock:
    """Test double implementing Clock."""

    def __init__(self, instant: datetime) -> None:
        self._instant = instant

    def now(self) -> datetime:
        return self._instant


def as_clock(clock: SystemClock | FixedClock) -> Clock:
    return clock
