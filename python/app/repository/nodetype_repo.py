"""
NodeType repository implementation.
"""

import uuid
from datetime import datetime
from typing import List, Tuple

import asyncpg

from app.db.database import Database
from app.repository.models import NodeType, ListOptions, ListResult
from app.repository.errors import NotFoundError


class NodeTypeRepository:
    """PostgreSQL node type repository."""

    def __init__(self, db: Database):
        self.db = db

    async def create(self, node_type: NodeType) -> NodeType:
        """Create a new node type."""
        node_type.id = str(uuid.uuid4())
        node_type.created_at = datetime.now()
        node_type.updated_at = datetime.now()

        # Parse schema to JSON or None - preserve empty/falsy JSON schemas like '{}' or '[]'
        schema_value = None
        if node_type.schema is not None and node_type.schema != "":
            schema_value = node_type.schema

        query = """
            INSERT INTO node_types (id, tenant_id, name, description, schema, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
            RETURNING id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node_type.id, node_type.tenant_id, node_type.name, node_type.description,
                schema_value,
                node_type.created_at, node_type.updated_at
            )

        return self._row_to_node_type(row)

    async def get_by_id(self, tenant_id: str, id: str) -> NodeType:
        """Retrieve a node type by ID and tenant ID."""
        query = """
            SELECT id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
            FROM node_types 
            WHERE id = $1 AND tenant_id = $2
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id, tenant_id)

        if not row:
            raise NotFoundError(f"node_type not found: {id}")

        return self._row_to_node_type(row)

    async def update(self, node_type: NodeType) -> NodeType:
        """Update an existing node type."""
        node_type.updated_at = datetime.now()

        # Preserve empty/falsy JSON schemas like '{}' or '[]'
        schema_value = None
        if node_type.schema is not None and node_type.schema != "":
            schema_value = node_type.schema

        query = """
            UPDATE node_types 
            SET name = $3, description = $4, schema = $5::jsonb, updated_at = $6
            WHERE id = $1 AND tenant_id = $2
            RETURNING id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node_type.id, node_type.tenant_id, node_type.name, node_type.description,
                schema_value,
                node_type.updated_at
            )

        if not row:
            raise NotFoundError(f"node_type not found: {node_type.id}")

        return self._row_to_node_type(row)

    async def delete(self, tenant_id: str, id: str) -> None:
        """Delete a node type by ID and tenant ID."""
        query = "DELETE FROM node_types WHERE id = $1 AND tenant_id = $2"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id, tenant_id)

        if result == "DELETE 0":
            raise NotFoundError(f"node_type not found: {id}")

    async def list(self, tenant_id: str, opts: ListOptions) -> Tuple[List[NodeType], ListResult]:
        """Retrieve node types with pagination."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        async with self.db.pool.acquire() as conn:
            total_count = await conn.fetchval(
                "SELECT COUNT(*) FROM node_types WHERE tenant_id = $1",
                tenant_id
            )

            query = """
                SELECT id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
                FROM node_types 
                WHERE tenant_id = $1
                ORDER BY created_at DESC 
                LIMIT $2 OFFSET $3
            """
            rows = await conn.fetch(query, tenant_id, page_size, offset)

        node_types = [self._row_to_node_type(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(node_types)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return node_types, result

    def _row_to_node_type(self, row: asyncpg.Record) -> NodeType:
        """Convert a database row to a NodeType object."""
        return NodeType(
            id=str(row[0]),
            tenant_id=str(row[1]),
            name=row[2],
            description=row[3] or "",
            schema=row[4] or "",
            created_at=row[5],
            updated_at=row[6],
        )
