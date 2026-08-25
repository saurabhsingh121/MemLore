from __future__ import annotations

import os

import pytest


def pytest_configure(config: pytest.Config) -> None:
    config.addinivalue_line(
        "markers",
        "integration: tests that require PostgreSQL "
        "(docker compose up -d postgres or MEMLORE_TEST_DATABASE_URL)",
    )


@pytest.fixture(scope="session")
def database_url() -> str:
    return os.environ.get(
        "MEMLORE_TEST_DATABASE_URL",
        "postgresql+psycopg://memlore:memlore@localhost:15432/memlore",
    )


@pytest.fixture(scope="session")
def postgres_available(database_url: str) -> bool:
    try:
        from sqlalchemy import create_engine, text

        engine = create_engine(database_url)
        with engine.connect() as conn:
            conn.execute(text("SELECT 1"))
        return True
    except Exception:
        return False


@pytest.fixture
def require_postgres(postgres_available: bool) -> None:
    if not postgres_available:
        pytest.skip(
            "PostgreSQL unavailable. Run `docker compose up -d postgres` "
            "or set MEMLORE_TEST_DATABASE_URL."
        )
