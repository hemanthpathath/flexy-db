"""
JSON-RPC error handling.
"""

from app.repository.errors import NotFoundError


def map_error(err: Exception) -> dict:
    """Map domain errors to JSON-RPC error responses."""
    if isinstance(err, NotFoundError):
        return {"code": -32001, "message": str(err)}
    if isinstance(err, ValueError):
        return {"code": -32602, "message": str(err)}
    return {"code": -32603, "message": str(err)}
