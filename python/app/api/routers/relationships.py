"""
Relationship REST API router.
"""

from fastapi import APIRouter, HTTPException, Query
from typing import Optional

from app.api.models import (
    RelationshipCreate,
    RelationshipUpdate,
    RelationshipResponse,
    RelationshipListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error

# Service instances will be set by main.py
_relationship_service = None


def set_relationship_service(service):
    """Set the relationship service instance."""
    global _relationship_service
    _relationship_service = service


router = APIRouter(prefix="/tenants/{tenant_id}/relationships", tags=["Relationships"])


@router.post(
    "",
    response_model=RelationshipResponse,
    status_code=201,
    summary="Create a relationship",
    description="Create a new relationship between two nodes within a tenant.",
    responses={
        201: {"description": "Relationship created successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Tenant or nodes not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def create_relationship(tenant_id: str, relationship: RelationshipCreate):
    """Create a new relationship."""
    try:
        if _relationship_service is None:
            raise RuntimeError("Relationship service not initialized")
        rel_obj = await _relationship_service.create(
            tenant_id,
            relationship.source_node_id,
            relationship.target_node_id,
            relationship.relationship_type,
            relationship.data or "{}"
        )
        return RelationshipResponse(relationship=rel_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "/{relationship_id}",
    response_model=RelationshipResponse,
    summary="Get a relationship",
    description="Get a relationship by its ID within a tenant.",
    responses={
        200: {"description": "Relationship found"},
        404: {"description": "Relationship or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def get_relationship(tenant_id: str, relationship_id: str):
    """Get a relationship by ID."""
    try:
        if _relationship_service is None:
            raise RuntimeError("Relationship service not initialized")
        rel_obj = await _relationship_service.get_by_id(tenant_id, relationship_id)
        return RelationshipResponse(relationship=rel_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.put(
    "/{relationship_id}",
    response_model=RelationshipResponse,
    summary="Update a relationship",
    description="Update an existing relationship. Only provided fields will be updated.",
    responses={
        200: {"description": "Relationship updated successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Relationship or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def update_relationship(tenant_id: str, relationship_id: str, relationship: RelationshipUpdate):
    """Update an existing relationship."""
    try:
        if _relationship_service is None:
            raise RuntimeError("Relationship service not initialized")
        rel_type = relationship.relationship_type if relationship.relationship_type is not None else ""
        data = relationship.data if relationship.data is not None else ""
        rel_obj = await _relationship_service.update(tenant_id, relationship_id, rel_type, data)
        return RelationshipResponse(relationship=rel_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.delete(
    "/{relationship_id}",
    status_code=204,
    summary="Delete a relationship",
    description="Delete a relationship by its ID.",
    responses={
        204: {"description": "Relationship deleted successfully"},
        404: {"description": "Relationship or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def delete_relationship(tenant_id: str, relationship_id: str):
    """Delete a relationship."""
    try:
        if _relationship_service is None:
            raise RuntimeError("Relationship service not initialized")
        await _relationship_service.delete(tenant_id, relationship_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "",
    response_model=RelationshipListResponse,
    summary="List relationships",
    description="List all relationships within a tenant with optional filtering.",
    responses={
        200: {"description": "List of relationships"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_relationships(
    tenant_id: str,
    source_node_id: Optional[str] = Query(default=None, description="Filter by source node ID"),
    target_node_id: Optional[str] = Query(default=None, description="Filter by target node ID"),
    relationship_type: Optional[str] = Query(default=None, description="Filter by relationship type"),
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List relationships for a tenant."""
    try:
        if _relationship_service is None:
            raise RuntimeError("Relationship service not initialized")
        rels, pagination = await _relationship_service.list(
            tenant_id,
            source_node_id,
            target_node_id,
            relationship_type,
            page_size,
            page_token
        )
        return RelationshipListResponse(
            relationships=[r.to_dict() for r in rels],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

