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

    async def create(self, node_type_id: str, data: str) -> Node:
        """Create a new node."""
        if not node_type_id:
            raise ValueError("node_type_id is required")

        # Validate that the node type exists (repository is already scoped to tenant database)
        node_type = await self.node_type_repo.get_by_id(node_type_id)

        node = Node(
            tenant_id="",  # Not stored in tenant database
            node_type_id=node_type_id,
            data=data,
        )
        return await self.repo.create(node)

    async def get_by_id(self, id: str) -> Node:
        """Retrieve a node by ID."""
        if not id:
            raise ValueError("id is required")
        return await self.repo.get_by_id(id)

    async def update(self, id: str, data: str) -> Node:
        """Update an existing node."""
        if not id:
            raise ValueError("id is required")

        node = await self.repo.get_by_id(id)

        if data:
            node.data = data

        return await self.repo.update(node)

    async def delete(self, id: str) -> None:
        """Delete a node."""
        if not id:
            raise ValueError("id is required")
        await self.repo.delete(id)

    async def list(
        self,
        node_type_id: Optional[str],
        page_size: int,
        page_token: str
    ) -> Tuple[List[Node], ListResult]:
        """Retrieve nodes with pagination and optional filtering."""
        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(node_type_id, opts)
