"""
User service implementation.
"""

from typing import List, Tuple

from app.repository import User, TenantUser, UserRepository, ListOptions, ListResult


class UserService:
    """User business logic service."""

    def __init__(self, repo: UserRepository):
        self.repo = repo

    async def create(self, email: str, display_name: str) -> User:
        """Create a new user."""
        if not email:
            raise ValueError("email is required")
        if not display_name:
            raise ValueError("display_name is required")

        user = User(email=email, display_name=display_name)
        return await self.repo.create(user)

    async def get_by_id(self, id: str) -> User:
        """Retrieve a user by ID."""
        if not id:
            raise ValueError("id is required")
        return await self.repo.get_by_id(id)

    async def update(self, id: str, email: str, display_name: str) -> User:
        """Update an existing user."""
        if not id:
            raise ValueError("id is required")

        user = await self.repo.get_by_id(id)

        if email:
            user.email = email
        if display_name:
            user.display_name = display_name

        return await self.repo.update(user)

    async def delete(self, id: str) -> None:
        """Delete a user."""
        if not id:
            raise ValueError("id is required")
        await self.repo.delete(id)

    async def list(self, page_size: int, page_token: str) -> Tuple[List[User], ListResult]:
        """Retrieve users with pagination."""
        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(opts)

    async def add_to_tenant(self, tenant_id: str, user_id: str, role: str) -> TenantUser:
        """Add a user to a tenant."""
        if not tenant_id:
            raise ValueError("tenant_id is required")
        if not user_id:
            raise ValueError("user_id is required")

        tenant_user = TenantUser(tenant_id=tenant_id, user_id=user_id, role=role)
        return await self.repo.add_to_tenant(tenant_user)

    async def remove_from_tenant(self, tenant_id: str, user_id: str) -> None:
        """Remove a user from a tenant."""
        if not tenant_id:
            raise ValueError("tenant_id is required")
        if not user_id:
            raise ValueError("user_id is required")
        await self.repo.remove_from_tenant(tenant_id, user_id)

    async def list_tenant_users(self, tenant_id: str, page_size: int, page_token: str) -> Tuple[List[TenantUser], ListResult]:
        """List users in a tenant."""
        if not tenant_id:
            raise ValueError("tenant_id is required")

        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list_tenant_users(tenant_id, opts)
