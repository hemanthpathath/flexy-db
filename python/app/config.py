"""
Database configuration module.
"""

import os
from dataclasses import dataclass


@dataclass
class Config:
    """Database configuration."""
    host: str = "localhost"
    port: int = 5432
    user: str = "postgres"
    password: str = "postgres"
    db_name: str = "dbaas"
    ssl_mode: str = "disable"

    def connection_string(self) -> str:
        """Return PostgreSQL connection string."""
        ssl_param = "sslmode=" + self.ssl_mode
        return f"postgresql://{self.user}:{self.password}@{self.host}:{self.port}/{self.db_name}?{ssl_param}"

    def dsn(self) -> str:
        """Return DSN for asyncpg connection."""
        return f"postgresql://{self.user}:{self.password}@{self.host}:{self.port}/{self.db_name}"


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
        db_name=os.getenv("DB_NAME", "dbaas"),
        ssl_mode=os.getenv("DB_SSL_MODE", "disable"),
    )
