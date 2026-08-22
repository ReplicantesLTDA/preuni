"""Typed-error HTTP responses (matches contracts/error_codes.md)."""

from __future__ import annotations

from fastapi.responses import JSONResponse


def typed_error_response(
    *,
    status_code: int,
    error_code: str,
    message: str,
    details: dict | None = None,
) -> JSONResponse:
    body: dict[str, object] = {"error_code": error_code, "message": message}
    if details:
        body["details"] = details
    return JSONResponse(status_code=status_code, content=body)
