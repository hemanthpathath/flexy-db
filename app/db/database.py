"""
Database connection and migrations module.
"""

import asyncio
import logging
import os
import ssl
from pathlib import Path
from typing import Optional

import asyncpg

from app.config import Config

logger = logging.getLogger(__name__)


class Database:
    """Database connection pool wrapper."""

    def __init__(self, pool: asyncpg.Pool):
        self.pool = pool

    async def close(self):
        """Close the database connection pool."""
        await self.pool.close()


async def connect(cfg: Config) -> Database:
    """Create a new database connection pool."""
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
            database=cfg.db_name,
            min_size=1,
            max_size=10,
            ssl=ssl_context,
        )
        # Test the connection
        async with pool.acquire() as conn:
            await conn.execute("SELECT 1")
        
        return Database(pool)
    except Exception as e:
        raise Exception(f"Failed to connect to database: {e}") from e


async def run_migrations(db: Database) -> None:
    """Apply all SQL migrations."""
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
        migrations_dir = Path(__file__).parent / "migrations"
        up_files = sorted([f for f in os.listdir(migrations_dir) if f.endswith(".up.sql")])

        for filename in up_files:
            version = filename.replace(".up.sql", "")
            if version in applied:
                logger.info(f"Migration {version} already applied, skipping")
                continue

            logger.info(f"Applying migration {version}")
            content = (migrations_dir / filename).read_text()
            
            # Execute the migration in a transaction
            async with conn.transaction():
                await conn.execute(content)
                await conn.execute(
                    "INSERT INTO schema_migrations (version) VALUES ($1)",
                    version
                )
