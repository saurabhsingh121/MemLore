from __future__ import annotations

import os
import subprocess
import sys
import uuid

import pytest
from sqlalchemy import create_engine, text


@pytest.mark.integration
def test_alembic_upgrade_creates_governance_tables(
    require_postgres: None, database_url: str
) -> None:
    admin_url = database_url.rsplit("/", 1)[0] + "/postgres"
    db_name = "memlore_alembic_" + uuid.uuid4().hex[:8]
    admin = create_engine(admin_url, isolation_level="AUTOCOMMIT")
    with admin.connect() as conn:
        conn.execute(text(f'CREATE DATABASE "{db_name}"'))
    test_url = database_url.rsplit("/", 1)[0] + f"/{db_name}"

    env = os.environ.copy()
    env["MEMLORE_POSTGRES_DSN"] = test_url
    result = subprocess.run(
        [sys.executable, "-m", "alembic", "upgrade", "head"],
        cwd=os.path.dirname(os.path.dirname(os.path.dirname(__file__))),
        env=env,
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr + result.stdout

    engine = create_engine(test_url)
    with engine.connect() as conn:
        for table in ("lore_entries", "audit_records"):
            exists = conn.execute(
                text(
                    "SELECT EXISTS (SELECT 1 FROM information_schema.tables "
                    "WHERE table_schema = 'public' AND table_name = :name)"
                ),
                {"name": table},
            ).scalar_one()
            assert exists, f"{table} missing after alembic upgrade"

    with admin.connect() as conn:
        conn.execute(text(f'DROP DATABASE "{db_name}" WITH (FORCE)'))
    admin.dispose()
    engine.dispose()
