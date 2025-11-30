"""
Relationship repository implementation.
"""

import uuid
from datetime import datetime
from typing import List, Optional, Tuple

import asyncpg

from app.db.database import Database
from app.repository.models import Relationship, ListOptions, ListResult
from app.repository.errors import NotFoundError


class RelationshipRepository:
    """PostgreSQL relationship repository."""

    def __init__(self, db: Database):
        self.db = db

    async def create(self, rel: Relationship) -> Relationship:
        """Create a new relationship."""
        rel.id = str(uuid.uuid4())
        rel.created_at = datetime.now()
        rel.updated_at = datetime.now()

        if not rel.data:
            rel.data = "{}"

        query = """
            INSERT INTO relationships (id, source_node_id, target_node_id, relationship_type, data, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
            RETURNING id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                rel.id, rel.source_node_id, rel.target_node_id,
                rel.relationship_type, rel.data, rel.created_at, rel.updated_at
            )

        return self._row_to_relationship(row)

    async def get_by_id(self, id: str) -> Relationship:
        """Retrieve a relationship by ID."""
        query = """
            SELECT id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at 
            FROM relationships 
            WHERE id = $1
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(query, id)

        if not row:
            raise NotFoundError(f"relationship not found: {id}")

        return self._row_to_relationship(row)

    async def update(self, rel: Relationship) -> Relationship:
        """Update an existing relationship."""
        rel.updated_at = datetime.now()

        if not rel.data:
            rel.data = "{}"

        query = """
            UPDATE relationships 
            SET relationship_type = $2, data = $3::jsonb, updated_at = $4
            WHERE id = $1
            RETURNING id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at
        """

        async with self.db.pool.acquire() as conn:
            row = await conn.fetchrow(
                query,
                rel.id, rel.relationship_type, rel.data, rel.updated_at
            )

        if not row:
            raise NotFoundError(f"relationship not found: {rel.id}")

        return self._row_to_relationship(row)

    async def delete(self, id: str) -> None:
        """Delete a relationship by ID."""
        query = "DELETE FROM relationships WHERE id = $1"

        async with self.db.pool.acquire() as conn:
            result = await conn.execute(query, id)

        if result == "DELETE 0":
            raise NotFoundError(f"relationship not found: {id}")

    async def list(
        self,
        source_node_id: Optional[str],
        target_node_id: Optional[str],
        rel_type: Optional[str],
        opts: ListOptions
    ) -> Tuple[List[Relationship], ListResult]:
        """Retrieve relationships with pagination and optional filtering."""
        page_size = max(1, min(opts.page_size or 10, 100))
        offset = 0
        if opts.page_token:
            try:
                offset = int(opts.page_token)
            except ValueError:
                offset = 0

        # Build dynamic query with filters
        count_query = "SELECT COUNT(*) FROM relationships WHERE 1=1"
        list_query = """
            SELECT id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at 
            FROM relationships 
            WHERE 1=1
        """
        args = []
        arg_idx = 1

        if source_node_id:
            count_query += f" AND source_node_id = ${arg_idx}"
            list_query += f" AND source_node_id = ${arg_idx}"
            args.append(source_node_id)
            arg_idx += 1

        if target_node_id:
            count_query += f" AND target_node_id = ${arg_idx}"
            list_query += f" AND target_node_id = ${arg_idx}"
            args.append(target_node_id)
            arg_idx += 1

        if rel_type:
            count_query += f" AND relationship_type = ${arg_idx}"
            list_query += f" AND relationship_type = ${arg_idx}"
            args.append(rel_type)
            arg_idx += 1

        list_query += f" ORDER BY created_at DESC LIMIT ${arg_idx} OFFSET ${arg_idx + 1}"
        list_args = args + [page_size, offset]

        async with self.db.pool.acquire() as conn:
            total_count = await conn.fetchval(count_query, *args)
            rows = await conn.fetch(list_query, *list_args)

        relationships = [self._row_to_relationship(row) for row in rows]

        result = ListResult(total_count=total_count)
        next_offset = offset + len(relationships)
        if next_offset < total_count:
            result.next_page_token = str(next_offset)

        return relationships, result

    def _row_to_relationship(self, row: asyncpg.Record) -> Relationship:
        """Convert a database row to a Relationship object."""
        return Relationship(
            id=str(row[0]),
            tenant_id="",  # Not stored in tenant database (each tenant has own DB)
            source_node_id=str(row[1]),
            target_node_id=str(row[2]),
            relationship_type=row[3],
            data=row[4] or "{}",
            created_at=row[5],
            updated_at=row[6],
        )
