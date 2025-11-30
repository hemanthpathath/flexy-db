"""
Tenant repository implementation.
"""

import uuid
from datetime import datetime
from typing import List, Tuple

import asyncpg

from app.db.database import Database
from app.repository.models import Tenant, ListOptions, ListResult
from app.repository.errors import NotFoundError


class TenantRepository:
    """PostgreSQL tenant repository."""

    def __init__(self, db: Database):
        self.db = db

    async def create(self, tenant: Tenant) -> Tenant:
        """Create a new tenant."""
        tenant.id = str(uuid.uuid4())
        tenant.created_at = datetime.now()
        tenant.updated_at = datetime.now()
        if not tenant.status:
            tenant.status = "active"

        query = """
            INSERT INTO tenants (id, slug, name, status, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING id, slug, name, status, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                tenant.id, tenant.slug, tenant.name, tenant.status,
                tenant.created_at, tenant.updated_at
            )

        return self._row_to_tenant(row)

    async def get_by_id(self, id: str) -> Tenant:
        """Retrieve a tenant by ID."""
        query = "SELECT id, slug, name, status, created_at, updated_at FROM tenants WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id)

        if not row:
            raise NotFoundError(f"tenant not found: {id}")

        return self._row_to_tenant(row)

    async def update(self, tenant: Tenant) -> Tenant:
        """Update an existing tenant."""
        tenant.updated_at = datetime.now()

        query = """
            UPDATE tenants 
            SET slug = $2, name = $3, status = $4, updated_at = $5
            WHERE id = $1
            RETURNING id, slug, name, status, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                tenant.id, tenant.slug, tenant.name, tenant.status, tenant.updated_at
            )

        if not row:
            raise NotFoundError(f"tenant not found: {tenant.id}")

        return self._row_to_tenant(row)

    async def delete(self, id: str) -> None:
        """Delete a tenant by ID."""
        query = "DELETE FROM tenants WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id)

        # Check if any row was affected
        if result == "DELETE 0":
            raise NotFoundError(f"tenant not found: {id}")

    async def list(self, opts: ListOptions) -> Tuple[List[Tenant], ListResult]:
        """Retrieve tenants with pagination."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        async with self.db.pool.acquire() as conn:
            # Get total count
            total_count = await conn.fetchval("SELECT COUNT(*) FROM tenants")

            # Get tenants
            query = """
                SELECT id, slug, name, status, created_at, updated_at 
                FROM tenants 
                ORDER BY created_at DESC 
                LIMIT $1 OFFSET $2
            """
            rows = await conn.fetch(query, page_size, offset)

        tenants = [self._row_to_tenant(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(tenants)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return tenants, result

    def _row_to_tenant(self, row: asyncpg.Record) -> Tenant:
        """Convert a database row to a Tenant object."""
        return Tenant(
            id=str(row["id"]),
            slug=row["slug"],
            name=row["name"],
            status=row["status"],
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )
