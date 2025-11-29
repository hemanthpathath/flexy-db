#!/usr/bin/env python3
"""
flex-db Python Backend - Main Entry Point

A Database-as-a-Service (DBaaS) implemented in Python with JSON-RPC API.
"""

import logging
import os
import sys

from contextlib import asynccontextmanager
from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn

from app.config import config_from_env
from app.db import (
    connect_control_db,
    run_control_migrations,
    ensure_control_database_exists,
    TenantDatabaseManager,
)
from app.repository import (
    TenantRepository,
    UserRepository,
)
from app.service import (
    TenantService,
    UserService,
)
from app.jsonrpc import register_methods, jsonrpc_router
from app.api.dependencies import set_tenant_db_manager

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

# Global database instances
_control_db = None
_tenant_db_manager = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for FastAPI app."""
    global _control_db, _tenant_db_manager
    
    # Startup
    logger.info("Starting up...")
    
    # Load environment variables from .env.local if it exists
    env_file = os.path.join(os.path.dirname(os.path.dirname(__file__)), ".env.local")
    if os.path.exists(env_file):
        load_dotenv(env_file)
        logger.info(f"Loaded environment from {env_file}")

    # Load configuration from environment variables
    cfg = config_from_env()

    # Ensure control database exists
    logger.info("Ensuring control database exists...")
    try:
        await ensure_control_database_exists(cfg)
    except Exception as e:
        logger.error(f"Failed to ensure control database exists: {e}")
        sys.exit(1)

    # Connect to control database
    logger.info("Connecting to control database...")
    try:
        _control_db = await connect_control_db(cfg)
        logger.info("Connected to control database successfully")
    except Exception as e:
        logger.error(f"Failed to connect to control database: {e}")
        sys.exit(1)

    # Run control database migrations
    logger.info("Running control database migrations...")
    try:
        await run_control_migrations(_control_db)
        logger.info("Control database migrations completed successfully")
    except Exception as e:
        logger.error(f"Failed to run control database migrations: {e}")
        await _control_db.close()
        sys.exit(1)

    # Initialize tenant database manager
    logger.info("Initializing tenant database manager...")
    try:
        _tenant_db_manager = TenantDatabaseManager(cfg, _control_db)
        set_tenant_db_manager(_tenant_db_manager)
        logger.info("Tenant database manager initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize tenant database manager: {e}")
        await _control_db.close()
        sys.exit(1)

    # Initialize control database repositories
    tenant_repo = TenantRepository(_control_db)
    user_repo = UserRepository(_control_db)

    # Initialize control database services (tenant and user services work with control DB)
    tenant_svc = TenantService(tenant_repo, _tenant_db_manager)
    user_svc = UserService(user_repo)

    # Register JSON-RPC methods (tenant-scoped services are resolved per-request)
    register_methods(tenant_svc, user_svc)

    logger.info("Services initialized successfully")
    
    yield
    
    # Shutdown
    logger.info("Shutting down...")
    if _tenant_db_manager:
        await _tenant_db_manager.close_all_pools()
    if _control_db:
        await _control_db.close()
    logger.info("Shutdown complete")


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(
        title="flex-db API",
        description="Database-as-a-Service with JSON-RPC 2.0 API",
        version="1.0.0",
        docs_url=None,  # Disable Swagger UI (we use OpenRPC instead)
        redoc_url=None,  # Disable ReDoc (we use OpenRPC instead)
        openapi_url=None,  # Disable OpenAPI schema (we use OpenRPC instead)
        lifespan=lifespan,
    )
    
    # Add CORS middleware
    # Note: allow_origins=["*"] with allow_credentials=True is insecure for production
    # In production, specify exact origins: allow_origins=["https://yourdomain.com"]
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # TODO: Restrict to specific origins in production
        allow_credentials=False,  # Set to False when using allow_origins=["*"]
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Register JSON-RPC router
    app.include_router(jsonrpc_router)
    
    # Health check endpoint
    @app.get("/health")
    async def health_check():
        """Health check endpoint."""
        return {"status": "ok"}
    
    return app


# Create the application instance
app = create_app()


if __name__ == "__main__":
    # Get server configuration
    host = os.getenv("JSONRPC_HOST", "0.0.0.0")
    port = int(os.getenv("JSONRPC_PORT", "5000"))
    
    logger.info(f"Starting flex-db server on {host}:{port}...")
    logger.info(f"JSON-RPC endpoint: http://{host}:{port}/jsonrpc")
    logger.info(f"OpenRPC spec: http://{host}:{port}/openrpc.json")
    logger.info(f"Health check: http://{host}:{port}/health")
    
    uvicorn.run(
        "main:app",
        host=host,
        port=port,
        reload=os.getenv("RELOAD", "false").lower() == "true",
    )
