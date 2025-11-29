"""
JSON-RPC module initialization.
"""

from app.jsonrpc.handlers import register_methods
from app.jsonrpc.server import router as jsonrpc_router

__all__ = ["register_methods", "jsonrpc_router"]
