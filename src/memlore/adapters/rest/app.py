from __future__ import annotations

from fastapi import FastAPI

from memlore import __version__
from memlore.domain.models.health import HealthStatus


def create_app() -> FastAPI:
    """Build the REST application (composition root for HTTP adapters)."""

    app = FastAPI(
        title="MemLore",
        description="Shared engineering memory for humans and AI coding agents",
        version=__version__,
    )

    @app.get("/health", response_model=HealthStatus)
    def health() -> HealthStatus:
        return HealthStatus(version=__version__)

    return app
