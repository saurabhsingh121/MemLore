from __future__ import annotations

import logging
from typing import Any


def get_logger(name: str = "memlore") -> logging.Logger:
    logger = logging.getLogger(name)
    if not logger.handlers:
        handler = logging.StreamHandler()
        handler.setFormatter(
            logging.Formatter("%(asctime)s %(levelname)s %(name)s %(message)s")
        )
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
    return logger


def log_operation(
    logger: logging.Logger,
    *,
    operation: str,
    actor_id: str | None = None,
    lore_entry_id: str | None = None,
    **extra: Any,
) -> None:
    payload = {
        "operation": operation,
        "actor_id": actor_id,
        "lore_entry_id": lore_entry_id,
        **extra,
    }
    logger.info("%s", payload)
