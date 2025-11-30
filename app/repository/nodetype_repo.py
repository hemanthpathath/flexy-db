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
            INSERT INTO node_types (id, name, description, schema, created_at, updated_at)
            VALUES ($1, $2, $3, $4::jsonb, $5, $6)
            RETURNING id, name, description, COALESCE(schema::text, ''), created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node_type.id, node_type.name, node_type.description,
                schema_value,
                node_type.created_at, node_type.updated_at
            )

        return self._row_to_node_type(row)

    async def get_by_id(self, id: str) -> NodeType:
        """Retrieve a node type by ID."""
        query = """
            SELECT id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
            FROM node_types 
            WHERE id = $1
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id)

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
            SET name = $2, description = $3, schema = $4::jsonb, updated_at = $5
            WHERE id = $1
            RETURNING id, name, description, COALESCE(schema::text, ''), created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                node_type.id, node_type.name, node_type.description,
                schema_value,
                node_type.updated_at
            )

        if not row:
            raise NotFoundError(f"node_type not found: {node_type.id}")

        return self._row_to_node_type(row)

    async def delete(self, id: str) -> None:
        """Delete a node type by ID."""
        query = "DELETE FROM node_types WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id)

        if result == "DELETE 0":
            raise NotFoundError(f"node_type not found: {id}")

    async def list(self, opts: ListOptions) -> Tuple[List[NodeType], ListResult]:
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
                "SELECT COUNT(*) FROM node_types"
            )

            query = """
                SELECT id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
                FROM node_types 
                ORDER BY created_at DESC 
                LIMIT $1 OFFSET $2
            """
            rows = await conn.fetch(query, page_size, offset)

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
            tenant_id="",  # Not stored in tenant database (each tenant has own DB)
            name=row[1],
            description=row[2] or "",
            schema=row[3] or "",
            created_at=row[4],
            updated_at=row[5],
        )
