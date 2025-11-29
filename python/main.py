#!/usr/bin/env python3
"""
flex-db Python Backend - Main Entry Point

A Database-as-a-Service (DBaaS) implemented in Python with JSON-RPC and REST API.
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
from app.db import connect, run_migrations
from app.repository import (
    TenantRepository,
    UserRepository,
    NodeTypeRepository,
    NodeRepository,
    RelationshipRepository,
)
from app.service import (
    TenantService,
    UserService,
    NodeTypeService,
    NodeService,
    RelationshipService,
)
from app.jsonrpc import register_methods, jsonrpc_router
from app.api.routers import (
    tenants,
    users,
    node_types,
    nodes,
    relationships,
)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

# Global database instance
_db = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Lifespan context manager for FastAPI app."""
    global _db
    
    # Startup
    logger.info("Starting up...")
    
    # Load environment variables from .env.local if it exists
    env_file = os.path.join(os.path.dirname(os.path.dirname(__file__)), ".env.local")
    if os.path.exists(env_file):
        load_dotenv(env_file)
        logger.info(f"Loaded environment from {env_file}")

    # Load configuration from environment variables
    cfg = config_from_env()

    # Connect to database
    logger.info("Connecting to database...")
    try:
        _db = await connect(cfg)
        logger.info("Connected to database successfully")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        sys.exit(1)

    # Run migrations
    logger.info("Running database migrations...")
    try:
        await run_migrations(_db)
        logger.info("Migrations completed successfully")
    except Exception as e:
        logger.error(f"Failed to run migrations: {e}")
        await _db.close()
        sys.exit(1)

    # Initialize repositories
    tenant_repo = TenantRepository(_db)
    user_repo = UserRepository(_db)
    nodetype_repo = NodeTypeRepository(_db)
    node_repo = NodeRepository(_db)
    relationship_repo = RelationshipRepository(_db)

    # Initialize services
    tenant_svc = TenantService(tenant_repo)
    user_svc = UserService(user_repo)
    nodetype_svc = NodeTypeService(nodetype_repo)
    node_svc = NodeService(node_repo, nodetype_repo)
    relationship_svc = RelationshipService(relationship_repo, node_repo)

    # Register JSON-RPC methods
    register_methods(tenant_svc, user_svc, nodetype_svc, node_svc, relationship_svc)
    
    # Register services with REST routers
    tenants.set_tenant_service(tenant_svc)
    users.set_user_service(user_svc)
    node_types.set_nodetype_service(nodetype_svc)
    nodes.set_node_service(node_svc)
    relationships.set_relationship_service(relationship_svc)

    logger.info("Services initialized successfully")
    
    yield
    
    # Shutdown
    logger.info("Shutting down...")
    if _db:
        await _db.close()
    logger.info("Shutdown complete")


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    # Get API metadata from environment
    api_title = os.getenv("API_TITLE", "flex-db API")
    api_description = os.getenv("API_DESCRIPTION", "Database-as-a-Service with JSON-RPC and REST API")
    api_version = os.getenv("API_VERSION", "1.0.0")
    
    app = FastAPI(
        title=api_title,
        description=api_description,
        version=api_version,
        docs_url="/docs",  # Swagger UI
        redoc_url="/redoc",  # ReDoc
        openapi_url="/openapi.json",  # OpenAPI schema
        lifespan=lifespan,
        openapi_tags=[
            {"name": "Tenants", "description": "Tenant management operations"},
            {"name": "Users", "description": "User management operations"},
            {"name": "Tenant Users", "description": "Tenant-user membership operations"},
            {"name": "Node Types", "description": "Node type management operations"},
            {"name": "Nodes", "description": "Node management operations"},
            {"name": "Relationships", "description": "Relationship management operations"},
            {"name": "JSON-RPC", "description": "JSON-RPC 2.0 protocol endpoint"},
            {"name": "Health", "description": "Health check endpoint"},
        ],
    )
    
    # Add CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    # Register JSON-RPC router
    app.include_router(jsonrpc_router, tags=["JSON-RPC"])
    
    # Register REST API routers
    app.include_router(tenants.router)
    app.include_router(users.router)
    app.include_router(users.tenant_users_router)
    app.include_router(node_types.router)
    app.include_router(nodes.router)
    app.include_router(relationships.router)
    
    # Health check endpoint
    @app.get("/health", tags=["Health"])
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
    logger.info(f"REST API docs: http://{host}:{port}/docs")
    logger.info(f"Health check: http://{host}:{port}/health")
    
    uvicorn.run(
        "main:app",
        host=host,
        port=port,
        reload=os.getenv("RELOAD", "false").lower() == "true",
    )
