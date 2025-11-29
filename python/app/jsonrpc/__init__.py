"""
JSON-RPC module initialization.
"""

from app.jsonrpc.handlers import register_methods
from app.jsonrpc.server import create_app

__all__ = ["register_methods", "create_app"]
