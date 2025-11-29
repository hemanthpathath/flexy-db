"""
Node repository implementation.
"""

import json
import uuid
from datetime import datetime
from typing import List, Optional, Tuple

import asyncpg

from app.db.database import Database
from app.repository.models import Node, ListOptions, ListResult
from app.repository.errors import NotFoundError


class NodeRepository:
    """PostgreSQL node repository."""

    def __init__(self, db: Database):
        self.db = db

    async def create(self, node: Node) -> Node:
        """Create a new node."""
        node.id = str(uuid.uuid4())
        node.created_at = datetime.now()
        node.updated_at = datetime.now()

        if not node.data:
            node.data = "{}"

        query = """
            INSERT INTO nodes (id, tenant_id, node_type_id, data, created_at, updated_at)
            VALUES ($1, $2, $3, $4::jsonb, $5, $6)
            RETURNING id, tenant_id, node_type_id, data::text, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node.id, node.tenant_id, node.node_type_id, node.data,
                node.created_at, node.updated_at
            )

        return self._row_to_node(row)

    async def get_by_id(self, tenant_id: str, id: str) -> Node:
        """Retrieve a node by ID and tenant ID."""
        query = """
            SELECT id, tenant_id, node_type_id, data::text, created_at, updated_at 
            FROM nodes 
            WHERE id = $1 AND tenant_id = $2
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id, tenant_id)

        if not row:
            raise NotFoundError(f"node not found: {id}")

        return self._row_to_node(row)

    async def update(self, node: Node) -> Node:
        """Update an existing node."""
        node.updated_at = datetime.now()

        if not node.data:
            node.data = "{}"

        query = """
            UPDATE nodes 
            SET data = $3::jsonb, updated_at = $4
            WHERE id = $1 AND tenant_id = $2
            RETURNING id, tenant_id, node_type_id, data::text, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node.id, node.tenant_id, node.data, node.updated_at
            )

        if not row:
            raise NotFoundError(f"node not found: {node.id}")

        return self._row_to_node(row)

    async def delete(self, tenant_id: str, id: str) -> None:
        """Delete a node by ID and tenant ID."""
        query = "DELETE FROM nodes WHERE id = $1 AND tenant_id = $2"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id, tenant_id)

        if result == "DELETE 0":
            raise NotFoundError(f"node not found: {id}")

    async def list(self, tenant_id: str, node_type_id: Optional[str], opts: ListOptions) -> Tuple[List[Node], ListResult]:
        """Retrieve nodes with pagination and optional filtering."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        async with self.db.pool.acquire() as conn:
            # Build count query
            if node_type_id:
                total_count = await conn.fetchval(
                    "SELECT COUNT(*) FROM nodes WHERE tenant_id = $1 AND node_type_id = $2",
                    tenant_id, node_type_id
                )
                query = """
                    SELECT id, tenant_id, node_type_id, data::text, created_at, updated_at 
                    FROM nodes 
                    WHERE tenant_id = $1 AND node_type_id = $2
                    ORDER BY created_at DESC 
                    LIMIT $3 OFFSET $4
                """
                rows = await conn.fetch(query, tenant_id, node_type_id, page_size, offset)
            else:
                total_count = await conn.fetchval(
                    "SELECT COUNT(*) FROM nodes WHERE tenant_id = $1",
                    tenant_id
                )
                query = """
                    SELECT id, tenant_id, node_type_id, data::text, created_at, updated_at 
                    FROM nodes 
                    WHERE tenant_id = $1
                    ORDER BY created_at DESC 
                    LIMIT $2 OFFSET $3
                """
                rows = await conn.fetch(query, tenant_id, page_size, offset)

        nodes = [self._row_to_node(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(nodes)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return nodes, result

    def _row_to_node(self, row: asyncpg.Record) -> Node:
        """Convert a database row to a Node object."""
        return Node(
            id=str(row[0]),
            tenant_id=str(row[1]),
            node_type_id=str(row[2]),
            data=row[3] or "{}",
            created_at=row[4],
            updated_at=row[5],
        )
