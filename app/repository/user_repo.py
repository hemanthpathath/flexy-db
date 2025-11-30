"""
User repository implementation.
"""

import uuid
from datetime import datetime
from typing import List, Tuple

import asyncpg

from app.db.database import Database
from app.repository.models import User, TenantUser, ListOptions, ListResult
from app.repository.errors import NotFoundError


class UserRepository:
    """PostgreSQL user repository."""

    def __init__(self, db: Database):
        self.db = db

    async def create(self, user: User) -> User:
        """Create a new user."""
        user.id = str(uuid.uuid4())
        user.created_at = datetime.now()
        user.updated_at = datetime.now()

        query = """
            INSERT INTO users (id, email, display_name, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING id, email, display_name, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                user.id, user.email, user.display_name, user.created_at, user.updated_at
            )

        return self._row_to_user(row)

    async def get_by_id(self, id: str) -> User:
        """Retrieve a user by ID."""
        query = "SELECT id, email, display_name, created_at, updated_at FROM users WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id)

        if not row:
            raise NotFoundError(f"user not found: {id}")

        return self._row_to_user(row)

    async def update(self, user: User) -> User:
        """Update an existing user."""
        user.updated_at = datetime.now()

        query = """
            UPDATE users 
            SET email = $2, display_name = $3, updated_at = $4
            WHERE id = $1
            RETURNING id, email, display_name, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                user.id, user.email, user.display_name, user.updated_at
            )

        if not row:
            raise NotFoundError(f"user not found: {user.id}")

        return self._row_to_user(row)

    async def delete(self, id: str) -> None:
        """Delete a user by ID."""
        query = "DELETE FROM users WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id)

        if result == "DELETE 0":
            raise NotFoundError(f"user not found: {id}")

    async def list(self, opts: ListOptions) -> Tuple[List[User], ListResult]:
        """Retrieve users with pagination."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        async with self.db.pool.acquire() as conn:
            total_count = await conn.fetchval("SELECT COUNT(*) FROM users")

            query = """
                SELECT id, email, display_name, created_at, updated_at 
                FROM users 
                ORDER BY created_at DESC 
                LIMIT $1 OFFSET $2
            """
            rows = await conn.fetch(query, page_size, offset)

        users = [self._row_to_user(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(users)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return users, result

    async def add_to_tenant(self, tenant_user: TenantUser) -> TenantUser:
        """Add a user to a tenant."""
        if not tenant_user.role:
            tenant_user.role = "member"
        if not tenant_user.status:
            tenant_user.status = "active"

        query = """
            INSERT INTO tenant_users (tenant_id, user_id, role, status)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = $3, status = $4
            RETURNING tenant_id, user_id, role, status
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                tenant_user.tenant_id, tenant_user.user_id, tenant_user.role, tenant_user.status
            )

        return self._row_to_tenant_user(row)

    async def remove_from_tenant(self, tenant_id: str, user_id: str) -> None:
        """Remove a user from a tenant."""
        query = "DELETE FROM tenant_users WHERE tenant_id = $1 AND user_id = $2"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, tenant_id, user_id)

        if result == "DELETE 0":
            raise NotFoundError(f"tenant_user not found: tenant_id={tenant_id}, user_id={user_id}")

    async def list_tenant_users(self, tenant_id: str, opts: ListOptions) -> Tuple[List[TenantUser], ListResult]:
        """List users in a tenant."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        async with self.db.pool.acquire() as conn:
            total_count = await conn.fetchval(
                "SELECT COUNT(*) FROM tenant_users WHERE tenant_id = $1",
                tenant_id
            )

            query = """
                SELECT tenant_id, user_id, role, status 
                FROM tenant_users 
                WHERE tenant_id = $1
                ORDER BY user_id
                LIMIT $2 OFFSET $3
            """
            rows = await conn.fetch(query, tenant_id, page_size, offset)

        tenant_users = [self._row_to_tenant_user(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(tenant_users)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return tenant_users, result

    def _row_to_user(self, row: asyncpg.Record) -> User:
        """Convert a database row to a User object."""
        return User(
            id=str(row["id"]),
            email=row["email"],
            display_name=row["display_name"],
            created_at=row["created_at"],
            updated_at=row["updated_at"],
        )

    def _row_to_tenant_user(self, row: asyncpg.Record) -> TenantUser:
        """Convert a database row to a TenantUser object."""
        return TenantUser(
            tenant_id=str(row["tenant_id"]),
            user_id=str(row["user_id"]),
            role=row["role"],
            status=row["status"],
        )
