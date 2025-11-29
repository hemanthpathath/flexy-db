"""
Relationship service implementation.
"""

from typing import List, Optional, Tuple

from app.repository import Relationship, RelationshipRepository, NodeRepository, ListOptions, ListResult


class RelationshipService:
    """Relationship business logic service."""

    def __init__(self, repo: RelationshipRepository, node_repo: NodeRepository):
        self.repo = repo
        self.node_repo = node_repo

    async def create(
        self,
        source_node_id: str,
        target_node_id: str,
        rel_type: str,
        data: str
    ) -> Relationship:
        """Create a new relationship."""
        if not source_node_id:
            raise ValueError("source_node_id is required")
        if not target_node_id:
            raise ValueError("target_node_id is required")
        if not rel_type:
            raise ValueError("relationship_type is required")

        # Validate that the source node exists (repository is already scoped to tenant database)
        source_node = await self.node_repo.get_by_id(source_node_id)

        # Validate that the target node exists (repository is already scoped to tenant database)
        target_node = await self.node_repo.get_by_id(target_node_id)

        rel = Relationship(
            tenant_id="",  # Not stored in tenant database
            source_node_id=source_node_id,
            target_node_id=target_node_id,
            relationship_type=rel_type,
            data=data,
        )
        return await self.repo.create(rel)

    async def get_by_id(self, id: str) -> Relationship:
        """Retrieve a relationship by ID."""
        if not id:
            raise ValueError("id is required")
        return await self.repo.get_by_id(id)

    async def update(self, id: str, rel_type: str, data: str) -> Relationship:
        """Update an existing relationship."""
        if not id:
            raise ValueError("id is required")

        rel = await self.repo.get_by_id(id)

        if rel_type:
            rel.relationship_type = rel_type
        if data:
            rel.data = data

        return await self.repo.update(rel)

    async def delete(self, id: str) -> None:
        """Delete a relationship."""
        if not id:
            raise ValueError("id is required")
        await self.repo.delete(id)

    async def list(
        self,
        source_node_id: Optional[str],
        target_node_id: Optional[str],
        rel_type: Optional[str],
        page_size: int,
        page_token: str
    ) -> Tuple[List[Relationship], ListResult]:
        """Retrieve relationships with pagination and optional filtering."""
        opts = ListOptions(page_size=page_size, page_token=page_token)
        return await self.repo.list(source_node_id, target_node_id, rel_type, opts)
