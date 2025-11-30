"""
Control database connection and migrations module.

The control database stores tenant metadata and cross-tenant data (users).
Each tenant has its own isolated database for tenant-specific data.
"""

import logging
import os
import ssl
from pathlib import Path
from typing import Optional

import asyncpg

from app.config import Config
from app.db.database import Database

logger = logging.getLogger(__name__)


async def connect_control_db(cfg: Config) -> Database:
    """
    Create a connection pool to the control database.
    
    The control database stores:
    - Tenant metadata
    - Tenant database mappings
    - Users (cross-tenant)
    - Tenant-User memberships
    """
    try:
        # Map SSL mode to asyncpg ssl parameter
        ssl_context = None
        if cfg.ssl_mode == "require":
            ssl_context = "require"
        elif cfg.ssl_mode == "prefer":
            ssl_context = "prefer"
        elif cfg.ssl_mode == "verify-ca" or cfg.ssl_mode == "verify-full":
            ssl_context = ssl.create_default_context()
        # "disable" is the default (ssl_context = None)

        pool = await asyncpg.create_pool(
            host=cfg.host,
            port=cfg.port,
            user=cfg.user,
            password=cfg.password,
            database=cfg.control_db_name,
            min_size=1,
            max_size=10,
            ssl=ssl_context,
        )
        
        # Test the connection
        async with pool.acquire() as conn:
            await conn.execute("SELECT 1")
        
        logger.info(f"Connected to control database: {cfg.control_db_name}")
        return Database(pool)
    except Exception as e:
        raise Exception(f"Failed to connect to control database: {e}") from e


async def run_control_migrations(db: Database) -> None:
    """Apply all control database migrations."""
    async with db.pool.acquire() as conn:
        # Create migrations tracking table
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS schema_migrations (
                version TEXT PRIMARY KEY,
                applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            )
        """)

        # Get already applied migrations
        rows = await conn.fetch("SELECT version FROM schema_migrations")
        applied = {row["version"] for row in rows}

        # Read and apply migrations
        migrations_dir = Path(__file__).parent / "control_migrations"
        if not migrations_dir.exists():
            logger.warning(f"Control migrations directory not found: {migrations_dir}")
            return
        
        up_files = sorted([f for f in os.listdir(migrations_dir) if f.endswith(".up.sql")])

        for filename in up_files:
            version = filename.replace(".up.sql", "")
            if version in applied:
                logger.info(f"Control migration {version} already applied, skipping")
                continue

            logger.info(f"Applying control migration {version}")
            content = (migrations_dir / filename).read_text()
            
            # Execute the migration in a transaction
            async with conn.transaction():
                await conn.execute(content)
                await conn.execute(
                    "INSERT INTO schema_migrations (version) VALUES ($1)",
                    version
                )
        
        logger.info("Control database migrations completed")


async def ensure_control_database_exists(cfg: Config) -> None:
    """
    Ensure the control database exists. Creates it if it doesn't exist.
    
    This requires connecting to the default 'postgres' database first.
    """
    try:
        # Connect to default postgres database to create control database
        ssl_context = None
        if cfg.ssl_mode == "require":
            ssl_context = "require"
        elif cfg.ssl_mode == "prefer":
            ssl_context = "prefer"
        elif cfg.ssl_mode == "verify-ca" or cfg.ssl_mode == "verify-full":
            ssl_context = ssl.create_default_context()
        
        # Connect to default postgres database
        conn = await asyncpg.connect(
            host=cfg.host,
            port=cfg.port,
            user=cfg.user,
            password=cfg.password,
            database="postgres",  # Connect to default database
            ssl=ssl_context,
        )
        
        try:
            # Check if control database exists
            db_exists = await conn.fetchval(
                "SELECT 1 FROM pg_database WHERE datname = $1",
                cfg.control_db_name
            )
            
            if not db_exists:
                logger.info(f"Creating control database: {cfg.control_db_name}")
                # Terminate any existing connections to the database (if it's being dropped/recreated)
                # Note: We need to use string formatting here since database names can't be parameterized
                await conn.execute(
                    f"""
                    SELECT pg_terminate_backend(pid) 
                    FROM pg_stat_activity 
                    WHERE datname = '{cfg.control_db_name}' AND pid <> pg_backend_pid()
                    """
                )
                # Create the database (must use string formatting, not parameterized query)
                await conn.execute(f'CREATE DATABASE "{cfg.control_db_name}"')
                logger.info(f"Control database created: {cfg.control_db_name}")
            else:
                logger.info(f"Control database already exists: {cfg.control_db_name}")
        finally:
            await conn.close()
    except asyncpg.exceptions.DuplicateDatabaseError:
        # Database already exists, that's fine
        logger.info(f"Control database already exists: {cfg.control_db_name}")
    except Exception as e:
        raise Exception(f"Failed to ensure control database exists: {e}") from e

