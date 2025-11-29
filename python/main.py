#!/usr/bin/env python3
"""
flex-db Python Backend - Main Entry Point

A Database-as-a-Service (DBaaS) implemented in Python with JSON-RPC.
"""

import asyncio
import logging
import os

from aiohttp import web
from dotenv import load_dotenv

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
from app.jsonrpc import register_methods, create_app

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)


async def main():
    """Main entry point."""
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
        db = await connect(cfg)
        logger.info("Connected to database successfully")
    except Exception as e:
        logger.error(f"Failed to connect to database: {e}")
        sys.exit(1)

    # Run migrations
    logger.info("Running database migrations...")
    try:
        await run_migrations(db)
        logger.info("Migrations completed successfully")
    except Exception as e:
        logger.error(f"Failed to run migrations: {e}")
        await db.close()
        sys.exit(1)

    # Initialize repositories
    tenant_repo = TenantRepository(db)
    user_repo = UserRepository(db)
    nodetype_repo = NodeTypeRepository(db)
    node_repo = NodeRepository(db)
    relationship_repo = RelationshipRepository(db)

    # Initialize services
    tenant_svc = TenantService(tenant_repo)
    user_svc = UserService(user_repo)
    nodetype_svc = NodeTypeService(nodetype_repo)
    node_svc = NodeService(node_repo, nodetype_repo)
    relationship_svc = RelationshipService(relationship_repo, node_repo)

    # Register JSON-RPC methods
    register_methods(tenant_svc, user_svc, nodetype_svc, node_svc, relationship_svc)

    # Create and configure the web application
    app = create_app()

    # Get server configuration
    host = os.getenv("JSONRPC_HOST", "0.0.0.0")
    port = int(os.getenv("JSONRPC_PORT", "5000"))

    # Set up graceful shutdown
    async def shutdown_handler():
        logger.info("Shutting down...")
        await db.close()

    # Start the server
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)

    try:
        await site.start()
        logger.info(f"Starting JSON-RPC server on {host}:{port}...")
        logger.info(f"JSON-RPC endpoint: http://{host}:{port}/jsonrpc")
        logger.info(f"Health check endpoint: http://{host}:{port}/health")

        # Keep the server running until interrupted
        while True:
            await asyncio.sleep(3600)
    except asyncio.CancelledError:
        pass
    finally:
        await shutdown_handler()
        await runner.cleanup()


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("Received shutdown signal")
