"""
Tenant service implementation.
"""

from typing import List, Tuple, Optional

from app.repository import Tenant, TenantRepository, ListOptions, ListResult
from app.db.tenant_db_manager import TenantDatabaseManager


class TenantService:
    """Tenant business logic service."""

    def __init__(self, repo: TenantRepository, tenant_db_manager: Optional[TenantDatabaseManager] = None):
        self.repo = repo
        self.tenant_db_manager = tenant_db_manager

    async def create(self, slug: str, name: str) -> Tenant:
        """Create a new tenant and its associated tenant database."""
        if not slug:
            raise ValueError("slug is required")
        if not name:
            raise ValueError("name is required")

        # Create tenant record in control database
        tenant = Tenant(slug=slug, name=name)
        tenant = await self.repo.create(tenant)

        # Create tenant database and run migrations
        if self.tenant_db_manager:
            await self.tenant_db_manager.create_tenant_database(
                tenant_id=tenant.id,
                slug=tenant.slug
            )

        return tenant

    async def get_by_id(self, id: str) -> Tenant:
        """Retrieve a tenant by ID."""
        if not id:
            raise ValueError("id is required")
        return await self.repo.get_by_id(id)

    async def update(self, id: str, slug: str, name: str, status: str) -> Tenant:
        """Update an existing tenant."""
        if not id:
            raise ValueError("id is required")

        tenant = await self.repo.get_by_id(id)

        if slug:
            tenant.slug = slug
        if name:
            tenant.name = name
        if status:
            tenant.status = status

        return await self.repo.update(tenant)

    async def delete(self, id: str) -> None:
        """Delete a tenant."""
        if not id:
            raise ValueError("id is required")
        await self.repo.delete(id)

    async def list(self, page_size: int, page_token: str) -> Tuple[List[Tenant], ListResult]:
        """Retrieve tenants with pagination."""
        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(opts)
