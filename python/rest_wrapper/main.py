#!/usr/bin/env python3
"""
flex-db REST Wrapper - FastAPI Application

A REST API facade for the flex-db JSON-RPC backend.
"""

import logging
import os

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from rest_wrapper.config import config_from_env
from rest_wrapper.client import init_client
from rest_wrapper.routers import tenants, users, node_types, nodes, relationships

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    # Load environment variables
    env_file = os.path.join(os.path.dirname(os.path.dirname(__file__)), ".env.local")
    if os.path.exists(env_file):
        load_dotenv(env_file)
        logger.info(f"Loaded environment from {env_file}")
    
    # Load configuration
    config = config_from_env()
    
    # Initialize JSON-RPC client
    init_client(config)
    logger.info(f"Initialized JSON-RPC client with URL: {config.jsonrpc_url}")
    
    # Create FastAPI app with OpenAPI documentation
    app = FastAPI(
        title=config.title,
        description=config.description,
        version=config.version,
        docs_url="/docs",  # Swagger UI
        redoc_url="/redoc",  # ReDoc
        openapi_url="/openapi.json",  # OpenAPI schema
        openapi_tags=[
            {"name": "Tenants", "description": "Tenant management operations"},
            {"name": "Users", "description": "User management operations"},
            {"name": "Tenant Users", "description": "Tenant-user membership operations"},
            {"name": "Node Types", "description": "Node type management operations"},
            {"name": "Nodes", "description": "Node management operations"},
            {"name": "Relationships", "description": "Relationship management operations"},
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
    
    # Register routers
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
    import uvicorn
    
    config = config_from_env()
    logger.info(f"Starting REST wrapper on {config.host}:{config.port}")
    logger.info(f"Swagger UI: http://{config.host}:{config.port}/docs")
    logger.info(f"ReDoc: http://{config.host}:{config.port}/redoc")
    logger.info(f"OpenAPI schema: http://{config.host}:{config.port}/openapi.json")
    
    uvicorn.run(
        "rest_wrapper.main:app",
        host=config.host,
        port=config.port,
        reload=True,
    )
