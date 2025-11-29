"""
JSON-RPC server implementation using aiohttp.
"""

import json
import logging
from aiohttp import web
from jsonrpcserver import async_dispatch

logger = logging.getLogger(__name__)


async def handle_jsonrpc(request: web.Request) -> web.Response:
    """Handle JSON-RPC requests."""
    try:
        body = await request.text()
        response = await async_dispatch(body)
        
        if response is None:
            # Notification (no response needed)
            return web.Response(status=204)
        
        return web.Response(
            text=response,
            content_type="application/json",
        )
    except json.JSONDecodeError:
        error_response = {
            "jsonrpc": "2.0",
            "error": {"code": -32700, "message": "Parse error"},
            "id": None,
        }
        return web.Response(
            text=json.dumps(error_response),
            content_type="application/json",
        )
    except Exception as e:
        logger.exception("Error handling JSON-RPC request")
        error_response = {
            "jsonrpc": "2.0",
            "error": {"code": -32603, "message": str(e)},
            "id": None,
        }
        return web.Response(
            text=json.dumps(error_response),
            content_type="application/json",
        )


async def health_check(request: web.Request) -> web.Response:
    """Health check endpoint."""
    return web.Response(text="OK")


def create_app() -> web.Application:
    """Create and configure the aiohttp application."""
    app = web.Application()
    app.router.add_post("/jsonrpc", handle_jsonrpc)
    app.router.add_get("/health", health_check)
    return app
