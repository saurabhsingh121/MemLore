from __future__ import annotations

from fastapi import Request
from fastapi.responses import JSONResponse

from memlore.domain.exceptions import MemloreError, NotFoundError, ValidationError


def error_body(
    code: str, message: str, details: list[object] | None = None
) -> dict[str, object]:
    return {"error": {"code": code, "message": message, "details": details or []}}


async def memlore_error_handler(_: Request, exc: Exception) -> JSONResponse:
    if isinstance(exc, ValidationError):
        return JSONResponse(
            status_code=400,
            content=error_body("validation_error", str(exc)),
        )
    if isinstance(exc, NotFoundError):
        return JSONResponse(
            status_code=404,
            content=error_body("not_found", str(exc)),
        )
    if isinstance(exc, MemloreError):
        return JSONResponse(
            status_code=500,
            content=error_body("internal_error", "unexpected error"),
        )
    return JSONResponse(
        status_code=500,
        content=error_body("internal_error", "unexpected error"),
    )
