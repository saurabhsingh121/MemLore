from __future__ import annotations

from typing import Literal

from pydantic import BaseModel, Field


class HealthStatus(BaseModel):
    """Liveness/readiness payload for operators and local smoke checks."""

    status: Literal["ok"] = "ok"
    service: str = Field(default="memlore")
    version: str
