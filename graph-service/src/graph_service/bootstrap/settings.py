from __future__ import annotations

import os

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_prefix="MEMLORE_",
        extra="ignore",
        populate_by_name=True,
    )

    neo4j_uri: str = "bolt://localhost:7687"
    neo4j_user: str = "neo4j"
    neo4j_password: str = "memlore-dev-password"
    graph_service_host: str = "127.0.0.1"
    graph_service_port: int = 8090
    openai_api_key: str | None = Field(default=None, validation_alias="OPENAI_API_KEY")


def load_settings() -> Settings:
    settings = Settings()
    if settings.openai_api_key is None:
        settings.openai_api_key = os.environ.get("OPENAI_API_KEY")
    return settings
