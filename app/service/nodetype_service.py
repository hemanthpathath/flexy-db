"""
NodeType service implementation.
"""

from typing import List, Tuple

from app.repository import NodeType, NodeTypeRepository, ListOptions, ListResult


class NodeTypeService:
    """NodeType business logic service."""

    def __init__(self, repo: NodeTypeRepository):
        self.repo = repo

    async def create(self, name: str, description: str, schema: str) -> NodeType:
        """Create a new node type."""
        if not name:
            raise ValueError("name is required")

        node_type = NodeType(
            tenant_id="",  # Not stored in tenant database
            name=name,
            description=description,
            schema=schema,
        )
        return await self.repo.create(node_type)

    async def get_by_id(self, id: str) -> NodeType:
        """Retrieve a node type by ID."""
        if not id:
            raise ValueError("id is required")
        return await self.repo.get_by_id(id)

    async def update(self, id: str, name: str, description: str, schema: str) -> NodeType:
        """Update an existing node type."""
        if not id:
            raise ValueError("id is required")

        node_type = await self.repo.get_by_id(id)

        if name:
            node_type.name = name
        if description:
            node_type.description = description
        if schema:
            node_type.schema = schema

        return await self.repo.update(node_type)

    async def delete(self, id: str) -> None:
        """Delete a node type."""
        if not id:
            raise ValueError("id is required")
        await self.repo.delete(id)

    async def list(self, page_size: int, page_token: str) -> Tuple[List[NodeType], ListResult]:
        """Retrieve node types with pagination."""
        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(opts)
