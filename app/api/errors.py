"""
Error handling utilities for REST API.
"""

from fastapi import HTTPException
from app.repository.errors import NotFoundError


def handle_service_error(err: Exception) -> HTTPException:
    """Convert service exception to HTTP exception."""
    if isinstance(err, NotFoundError):
        return HTTPException(status_code=404, detail=str(err))
    elif isinstance(err, ValueError):
        return HTTPException(status_code=400, detail=str(err))
    else:
        return HTTPException(status_code=500, detail=str(err))

