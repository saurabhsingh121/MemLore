from __future__ import annotations

from fastapi import FastAPI

from memlore import __version__
from memlore.adapters.rest.errors import memlore_error_handler
from memlore.adapters.rest.routes_lore import get_container
from memlore.adapters.rest.routes_lore import router as lore_router
from memlore.bootstrap.container import AppContainer, build_container
from memlore.domain.exceptions import MemloreError, NotFoundError, ValidationError
from memlore.domain.models.health import HealthStatus


def create_app(container: AppContainer | None = None) -> FastAPI:
    """Build the REST application (composition root for HTTP adapters)."""

    app = FastAPI(
        title="MemLore",
        description="Shared engineering memory for humans and AI coding agents",
        version=__version__,
    )
    resolved = container or build_container()
    app.state.container = resolved

    def _container() -> AppContainer:
        state_container = app.state.container
        assert isinstance(state_container, AppContainer)
        return state_container

    app.dependency_overrides[get_container] = _container
    app.add_exception_handler(ValidationError, memlore_error_handler)
    app.add_exception_handler(NotFoundError, memlore_error_handler)
    app.add_exception_handler(MemloreError, memlore_error_handler)
    app.include_router(lore_router)

    @app.get("/health", response_model=HealthStatus)
    def health() -> HealthStatus:
        return HealthStatus(version=__version__)

    return app
