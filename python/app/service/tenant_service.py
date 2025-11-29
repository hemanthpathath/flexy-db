"""
Tenant service implementation.
"""

from typing import List, Tuple

from app.repository import Tenant, TenantRepository, ListOptions, ListResult


class TenantService:
    """Tenant business logic service."""

    def __init__(self, repo: TenantRepository):
        self.repo = repo

    async def create(self, slug: str, name: str) -> Tenant:
        """Create a new tenant."""
        if not slug:
            raise ValueError("slug is required")
        if not name:
            raise ValueError("name is required")

        tenant = Tenant(slug=slug, name=name)
        return await self.repo.create(tenant)

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
