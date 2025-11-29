"""
Tenant database manager for database-per-tenant architecture.

Manages connection pools to tenant-specific databases and handles
database creation and migrations on-demand.
"""

import logging
import os
import ssl
from pathlib import Path
from typing import Dict, Optional

import asyncpg

from app.config import Config
from app.db.database import Database
from app.db.control_database import connect_control_db

logger = logging.getLogger(__name__)


class TenantDatabaseManager:
    """
    Manages tenant database connections and routing.
    
    Features:
    - Lazy initialization of tenant databases
    - Connection pool caching per tenant
    - Automatic database creation and migration
    - Integration with control database for tenant metadata
    """

    def __init__(self, cfg: Config, control_db: Optional[Database] = None):
        """
        Initialize tenant database manager.
        
        Args:
            cfg: Database configuration
            control_db: Optional control database connection. If not provided,
                       will create its own connection when needed.
        """
        self.cfg = cfg
        self.control_db = control_db
        self._tenant_pools: Dict[str, Database] = {}  # tenant_id -> Database pool
        self._pool_lock = None  # Will use asyncio.Lock if needed for thread safety

    async def get_tenant_db(self, tenant_id: str) -> Database:
        """
        Get or create connection pool for a tenant database.
        
        This method:
        1. Checks cache for existing pool
        2. Looks up tenant database name from control DB
        3. Creates tenant database if it doesn't exist
        4. Runs migrations if needed
        5. Creates and caches connection pool
        
        Args:
            tenant_id: UUID of the tenant
            
        Returns:
            Database connection pool for the tenant
        """
        # Check cache first
        if tenant_id in self._tenant_pools:
            return self._tenant_pools[tenant_id]

        # Get control database connection (use provided or create new)
        control_db = self.control_db
        if not control_db:
            control_db = await connect_control_db(self.cfg)

        # Look up tenant database name from control DB
        async with control_db.pool.acquire() as conn:
            # First, get tenant slug to determine database name
            tenant_row = await conn.fetchrow(
                "SELECT slug FROM tenants WHERE id = $1",
                tenant_id
            )
            if not tenant_row:
                raise ValueError(f"Tenant not found: {tenant_id}")

            slug = tenant_row["slug"]
            db_name = self.cfg.tenant_db_name(slug)

            # Check if database mapping exists
            db_mapping = await conn.fetchrow(
                "SELECT database_name, status FROM tenant_databases WHERE tenant_id = $1",
                tenant_id
            )

            if db_mapping and db_mapping["status"] == "active":
                # Database mapping exists, connect to it
                db_name = db_mapping["database_name"]
            else:
                # Need to create tenant database
                await self._create_tenant_database(tenant_id, slug, db_name, control_db, conn)

        # Connect to tenant database and run migrations
        tenant_db = await self._connect_tenant_database(db_name)
        await self._run_tenant_migrations(tenant_id, tenant_db)

        # Cache the pool
        self._tenant_pools[tenant_id] = tenant_db
        logger.info(f"Cached connection pool for tenant {tenant_id} (database: {db_name})")

        return tenant_db

    async def create_tenant_database(
        self,
        tenant_id: str,
        slug: str,
        control_db: Optional[Database] = None
    ) -> Database:
        """
        Create a new tenant database explicitly.
        
        This is typically called when creating a new tenant.
        The database will be created, migrations run, and the connection
        pool cached.
        
        Args:
            tenant_id: UUID of the tenant
            slug: Tenant slug (used for database naming)
            control_db: Optional control database connection
            
        Returns:
            Database connection pool for the new tenant
        """
        db_name = self.cfg.tenant_db_name(slug)

        # Use provided control DB or get cached one
        control_db = control_db or self.control_db
        if not control_db:
            control_db = await connect_control_db(self.cfg)

        async with control_db.pool.acquire() as conn:
            await self._create_tenant_database(tenant_id, slug, db_name, control_db, conn)

        # Connect to tenant database and run migrations
        tenant_db = await self._connect_tenant_database(db_name)
        await self._run_tenant_migrations(tenant_id, tenant_db)

        # Cache the pool
        self._tenant_pools[tenant_id] = tenant_db
        logger.info(f"Created and cached tenant database: {db_name} for tenant {tenant_id}")

        return tenant_db

    async def _create_tenant_database(
        self,
        tenant_id: str,
        slug: str,
        db_name: str,
        control_db: Database,
        control_conn
    ) -> None:
        """Internal method to create tenant database and record mapping."""
        try:
            # Connect to postgres database to create new database
            ssl_context = None
            if self.cfg.ssl_mode == "require":
                ssl_context = "require"
            elif self.cfg.ssl_mode == "prefer":
                ssl_context = "prefer"
            elif self.cfg.ssl_mode == "verify-ca" or self.cfg.ssl_mode == "verify-full":
                ssl_context = ssl.create_default_context()

            admin_conn = await asyncpg.connect(
                host=self.cfg.host,
                port=self.cfg.port,
                user=self.cfg.user,
                password=self.cfg.password,
                database="postgres",  # Connect to default database
                ssl=ssl_context,
            )

            try:
                # Check if database already exists
                db_exists = await admin_conn.fetchval(
                    "SELECT 1 FROM pg_database WHERE datname = $1",
                    db_name
                )

                if not db_exists:
                    logger.info(f"Creating tenant database: {db_name}")
                    # Create the database
                    await admin_conn.execute(f'CREATE DATABASE "{db_name}"')
                    logger.info(f"Tenant database created: {db_name}")
                else:
                    logger.info(f"Tenant database already exists: {db_name}")
            finally:
                await admin_conn.close()

            # Record database mapping in control database
            await control_conn.execute(
                """
                INSERT INTO tenant_databases (tenant_id, database_name)
                VALUES ($1, $2)
                ON CONFLICT (tenant_id) DO UPDATE
                SET database_name = EXCLUDED.database_name,
                    status = 'active'
                """,
                tenant_id,
                db_name
            )
            logger.info(f"Recorded tenant database mapping: {tenant_id} -> {db_name}")

        except asyncpg.exceptions.DuplicateDatabaseError:
            # Database already exists, that's fine
            logger.info(f"Tenant database already exists: {db_name}")
            # Still record the mapping
            await control_conn.execute(
                """
                INSERT INTO tenant_databases (tenant_id, database_name)
                VALUES ($1, $2)
                ON CONFLICT (tenant_id) DO UPDATE
                SET database_name = EXCLUDED.database_name,
                    status = 'active'
                """,
                tenant_id,
                db_name
            )
        except Exception as e:
            logger.error(f"Failed to create tenant database {db_name}: {e}")
            raise

    async def _connect_tenant_database(self, db_name: str) -> Database:
        """Connect to a tenant database and return Database wrapper."""
        try:
            # Map SSL mode to asyncpg ssl parameter
            ssl_context = None
            if self.cfg.ssl_mode == "require":
                ssl_context = "require"
            elif self.cfg.ssl_mode == "prefer":
                ssl_context = "prefer"
            elif self.cfg.ssl_mode == "verify-ca" or self.cfg.ssl_mode == "verify-full":
                ssl_context = ssl.create_default_context()

            pool = await asyncpg.create_pool(
                host=self.cfg.host,
                port=self.cfg.port,
                user=self.cfg.user,
                password=self.cfg.password,
                database=db_name,
                min_size=1,
                max_size=10,
                ssl=ssl_context,
            )

            # Test the connection
            async with pool.acquire() as conn:
                await conn.execute("SELECT 1")

            return Database(pool)
        except Exception as e:
            raise Exception(f"Failed to connect to tenant database {db_name}: {e}") from e

    async def _run_tenant_migrations(self, tenant_id: str, tenant_db: Database) -> None:
        """
        Run migrations on a tenant database.
        
        Tracks which migrations have been applied to which tenant database
        in the control database's tenant_migrations table.
        """
        # Get control database connection
        control_db = self.control_db
        if not control_db:
            control_db = await connect_control_db(self.cfg)

        # Get migrations already applied to this tenant (single query)
        async with control_db.pool.acquire() as control_conn:
            applied_rows = await control_conn.fetch(
                "SELECT version FROM tenant_migrations WHERE tenant_id = $1",
                tenant_id
            )
            applied = {row["version"] for row in applied_rows}

        async with tenant_db.pool.acquire() as conn:
            # Create migrations tracking table in tenant database
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS schema_migrations (
                    version TEXT PRIMARY KEY,
                    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
                )
            """)

            # Read and apply migrations
            migrations_dir = Path(__file__).parent / "tenant_migrations"
            if not migrations_dir.exists():
                logger.warning(f"Tenant migrations directory not found: {migrations_dir}")
                return

            up_files = sorted([f for f in os.listdir(migrations_dir) if f.endswith(".up.sql")])

            # Collect migrations to record in control DB
            new_migrations = []

            for filename in up_files:
                version = filename.replace(".up.sql", "")
                if version in applied:
                    logger.debug(f"Tenant {tenant_id} migration {version} already applied, skipping")
                    continue

                logger.info(f"Applying tenant migration {version} to tenant {tenant_id}")
                content = (migrations_dir / filename).read_text()

                # Execute the migration in a transaction
                async with conn.transaction():
                    await conn.execute(content)
                    # Record in tenant database
                    await conn.execute(
                        "INSERT INTO schema_migrations (version) VALUES ($1)",
                        version
                    )
                    new_migrations.append(version)

            # Record new migrations in control database (batch insert)
            if new_migrations:
                async with control_db.pool.acquire() as control_conn:
                    for version in new_migrations:
                        await control_conn.execute(
                            "INSERT INTO tenant_migrations (tenant_id, version) VALUES ($1, $2) ON CONFLICT DO NOTHING",
                            tenant_id,
                            version
                        )

            logger.info(f"Tenant migrations completed for tenant {tenant_id}")

    async def close_all_pools(self) -> None:
        """Close all cached tenant database connection pools."""
        logger.info(f"Closing {len(self._tenant_pools)} tenant database pools")
        for tenant_id, db in self._tenant_pools.items():
            try:
                await db.close()
                logger.debug(f"Closed pool for tenant {tenant_id}")
            except Exception as e:
                logger.error(f"Error closing pool for tenant {tenant_id}: {e}")
        self._tenant_pools.clear()

    async def evict_tenant_pool(self, tenant_id: str) -> None:
        """Evict a specific tenant's connection pool from cache."""
        if tenant_id in self._tenant_pools:
            try:
                await self._tenant_pools[tenant_id].close()
            except Exception as e:
                logger.error(f"Error closing pool for tenant {tenant_id}: {e}")
            del self._tenant_pools[tenant_id]
            logger.info(f"Evicted pool for tenant {tenant_id}")

