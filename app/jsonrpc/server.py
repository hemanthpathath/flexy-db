"""
JSON-RPC server implementation using FastAPI.
"""

import json
import logging
from fastapi import APIRouter, Request, Response, status
from jsonrpcserver import async_dispatch

logger = logging.getLogger(__name__)

router = APIRouter()


@router.post("/jsonrpc")
async def handle_jsonrpc(request: Request) -> Response:
    """Handle JSON-RPC requests."""
    try:
        body = await request.body()
        body_str = body.decode('utf-8')
        response = await async_dispatch(body_str)
        
        if response is None:
            # Notification (no response needed)
            return Response(status_code=status.HTTP_204_NO_CONTENT)
        
        return Response(
            content=response,
            media_type="application/json",
        )
    except json.JSONDecodeError:
        error_response = {
            "jsonrpc": "2.0",
            "error": {"code": -32700, "message": "Parse error"},
            "id": None,
        }
        return Response(
            content=json.dumps(error_response),
            media_type="application/json",
            status_code=status.HTTP_400_BAD_REQUEST,
        )
    except Exception as e:
        logger.exception("Error handling JSON-RPC request")
        error_response = {
            "jsonrpc": "2.0",
            "error": {"code": -32603, "message": str(e)},
            "id": None,
        }
        return Response(
            content=json.dumps(error_response),
            media_type="application/json",
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        )


@router.get("/openrpc.json")
async def get_openrpc_spec() -> Response:
    """
    Get OpenRPC specification for the JSON-RPC API.
    
    Returns the complete OpenRPC specification document that describes
    all available JSON-RPC methods, their parameters, return types, and errors.
    This spec can be used with OpenRPC tooling for interactive documentation,
    code generation, and API validation.
    
    See: https://spec.open-rpc.org/
    """
    try:
        from app.jsonrpc.openrpc import get_openrpc_spec_json
        spec_json = get_openrpc_spec_json()
        return Response(
            content=spec_json,
            media_type="application/json",
        )
    except Exception as e:
        logger.exception("Error generating OpenRPC spec")
        error_response = {
            "error": "Failed to generate OpenRPC specification",
            "message": str(e)
        }
        return Response(
            content=json.dumps(error_response),
            media_type="application/json",
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        )
