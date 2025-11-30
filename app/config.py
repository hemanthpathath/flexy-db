"""
Database configuration module.
"""

import os
from dataclasses import dataclass
from typing import Optional


@dataclass
class Config:
    """Database configuration."""
    host: str = "localhost"
    port: int = 5432
    user: str = "postgres"
    password: str = "postgres"
    # Control database for tenant metadata and cross-tenant data
    control_db_name: str = "dbaas_control"
    # Prefix for tenant database naming: {prefix}{tenant_slug}
    tenant_db_prefix: str = "dbaas_tenant_"
    # Legacy: kept for backward compatibility during migration
    db_name: str = "dbaas"
    ssl_mode: str = "disable"

    def connection_string(self, database: Optional[str] = None) -> str:
        """Return PostgreSQL connection string."""
        db = database or self.control_db_name
        ssl_param = "sslmode=" + self.ssl_mode
        return f"postgresql://{self.user}:{self.password}@{self.host}:{self.port}/{db}?{ssl_param}"

    def dsn(self, database: Optional[str] = None) -> str:
        """Return DSN for asyncpg connection."""
        db = database or self.control_db_name
        return f"postgresql://{self.user}:{self.password}@{self.host}:{self.port}/{db}"
    
    def tenant_db_name(self, tenant_slug: str) -> str:
        """Generate tenant database name from slug."""
        # Sanitize slug: lowercase, replace non-alphanumeric with underscore
        sanitized = "".join(c if c.isalnum() else "_" for c in tenant_slug.lower())
        return f"{self.tenant_db_prefix}{sanitized}"


def default_config() -> Config:
    """Return default database configuration."""
    return Config()


def config_from_env() -> Config:
    """Load configuration from environment variables."""
    return Config(
        host=os.getenv("DB_HOST", "localhost"),
        port=int(os.getenv("DB_PORT", "5432")),
        user=os.getenv("DB_USER", "postgres"),
        password=os.getenv("DB_PASSWORD", "postgres"),
        control_db_name=os.getenv("DB_CONTROL_NAME", "dbaas_control"),
        tenant_db_prefix=os.getenv("DB_TENANT_PREFIX", "dbaas_tenant_"),
        db_name=os.getenv("DB_NAME", "dbaas"),  # Legacy, kept for compatibility
        ssl_mode=os.getenv("DB_SSL_MODE", "disable"),
    )
