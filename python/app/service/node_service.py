"""
Node service implementation.
"""

from typing import List, Optional, Tuple

from app.repository import Node, NodeRepository, NodeTypeRepository, ListOptions, ListResult


class NodeService:
    """Node business logic service."""

    def __init__(self, repo: NodeRepository, node_type_repo: NodeTypeRepository):
        self.repo = repo
        self.node_type_repo = node_type_repo

    async def create(self, tenant_id: str, node_type_id: str, data: str) -> Node:
        """Create a new node."""
        if not tenant_id:
            raise ValueError("tenant_id is required")
        if not node_type_id:
            raise ValueError("node_type_id is required")

        # Validate that the node type belongs to the same tenant
        node_type = await self.node_type_repo.get_by_id(tenant_id, node_type_id)
        if node_type.tenant_id != tenant_id:
            raise ValueError("invalid node_type_id: node type does not belong to this tenant")

        node = Node(
            tenant_id=tenant_id,
            node_type_id=node_type_id,
            data=data,
        )
        return await self.repo.create(node)

    async def get_by_id(self, tenant_id: str, id: str) -> Node:
        """Retrieve a node by ID."""
        if not id:
            raise ValueError("id is required")
        if not tenant_id:
            raise ValueError("tenant_id is required")
        return await self.repo.get_by_id(tenant_id, id)

    async def update(self, tenant_id: str, id: str, data: str) -> Node:
        """Update an existing node."""
        if not id:
            raise ValueError("id is required")
        if not tenant_id:
            raise ValueError("tenant_id is required")

        node = await self.repo.get_by_id(tenant_id, id)

        if data:
            node.data = data

        return await self.repo.update(node)

    async def delete(self, tenant_id: str, id: str) -> None:
        """Delete a node."""
        if not id:
            raise ValueError("id is required")
        if not tenant_id:
            raise ValueError("tenant_id is required")
        await self.repo.delete(tenant_id, id)

    async def list(
        self,
        tenant_id: str,
        node_type_id: Optional[str],
        page_size: int,
        page_token: str
    ) -> Tuple[List[Node], ListResult]:
        """Retrieve nodes with pagination and optional filtering."""
        if not tenant_id:
            raise ValueError("tenant_id is required")

        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(tenant_id, node_type_id, opts)
