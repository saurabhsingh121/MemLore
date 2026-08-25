"""create lore_entries and audit_records

Revision ID: 0001_lore_audit
Revises:
Create Date: 2026-08-25

"""

from __future__ import annotations

from collections.abc import Sequence

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects import postgresql

revision: str = "0001_lore_audit"
down_revision: str | None = None
branch_labels: str | Sequence[str] | None = None
depends_on: str | Sequence[str] | None = None


def upgrade() -> None:
    op.create_table(
        "lore_entries",
        sa.Column("id", sa.String(length=36), primary_key=True),
        sa.Column("statement", sa.Text(), nullable=False),
        sa.Column("scope_kind", sa.String(length=64), nullable=False),
        sa.Column("scope_key", sa.String(length=512), nullable=False),
        sa.Column("origin", sa.String(length=64), nullable=False),
        sa.Column("verification_status", sa.String(length=32), nullable=False),
        sa.Column("evidence", postgresql.JSONB(astext_type=sa.Text()), nullable=False),
        sa.Column("created_by", sa.String(length=256), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
        sa.Column("verified_by", sa.String(length=256), nullable=True),
        sa.Column("verified_at", sa.DateTime(timezone=True), nullable=True),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False),
    )
    op.create_index(
        "ix_lore_entries_scope_created",
        "lore_entries",
        ["scope_kind", "scope_key", "created_at"],
    )
    op.create_table(
        "audit_records",
        sa.Column("id", sa.String(length=36), primary_key=True),
        sa.Column("target_id", sa.String(length=36), nullable=False),
        sa.Column("action", sa.String(length=32), nullable=False),
        sa.Column("actor_id", sa.String(length=256), nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False),
    )
    op.create_index("ix_audit_records_target_id", "audit_records", ["target_id"])
    op.create_index(
        "ix_audit_records_target_created",
        "audit_records",
        ["target_id", "created_at", "id"],
    )


def downgrade() -> None:
    op.drop_index("ix_audit_records_target_created", table_name="audit_records")
    op.drop_index("ix_audit_records_target_id", table_name="audit_records")
    op.drop_table("audit_records")
    op.drop_index("ix_lore_entries_scope_created", table_name="lore_entries")
    op.drop_table("lore_entries")
