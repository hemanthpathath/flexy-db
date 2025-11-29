"""
NodeType REST API router.
"""

from fastapi import APIRouter, HTTPException, Query

from app.api.models import (
    NodeTypeCreate,
    NodeTypeUpdate,
    NodeTypeResponse,
    NodeTypeListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error

# Service instances will be set by main.py
_nodetype_service = None


def set_nodetype_service(service):
    """Set the node type service instance."""
    global _nodetype_service
    _nodetype_service = service


router = APIRouter(prefix="/tenants/{tenant_id}/node-types", tags=["Node Types"])


@router.post(
    "",
    response_model=NodeTypeResponse,
    status_code=201,
    summary="Create a node type",
    description="Create a new node type within a tenant.",
    responses={
        201: {"description": "Node type created successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def create_node_type(tenant_id: str, node_type: NodeTypeCreate):
    """Create a new node type."""
    try:
        if _nodetype_service is None:
            raise RuntimeError("Node type service not initialized")
        node_type_obj = await _nodetype_service.create(
            tenant_id,
            node_type.name,
            node_type.description or "",
            node_type.json_schema or ""
        )
        return NodeTypeResponse(node_type=node_type_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "/{node_type_id}",
    response_model=NodeTypeResponse,
    summary="Get a node type",
    description="Get a node type by its ID within a tenant.",
    responses={
        200: {"description": "Node type found"},
        404: {"description": "Node type or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def get_node_type(tenant_id: str, node_type_id: str):
    """Get a node type by ID."""
    try:
        if _nodetype_service is None:
            raise RuntimeError("Node type service not initialized")
        node_type_obj = await _nodetype_service.get_by_id(tenant_id, node_type_id)
        return NodeTypeResponse(node_type=node_type_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.put(
    "/{node_type_id}",
    response_model=NodeTypeResponse,
    summary="Update a node type",
    description="Update an existing node type. Only provided fields will be updated.",
    responses={
        200: {"description": "Node type updated successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Node type or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def update_node_type(tenant_id: str, node_type_id: str, node_type: NodeTypeUpdate):
    """Update an existing node type."""
    try:
        if _nodetype_service is None:
            raise RuntimeError("Node type service not initialized")
        name = node_type.name if node_type.name is not None else ""
        description = node_type.description if node_type.description is not None else ""
        schema = node_type.json_schema if node_type.json_schema is not None else ""
        node_type_obj = await _nodetype_service.update(tenant_id, node_type_id, name, description, schema)
        return NodeTypeResponse(node_type=node_type_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.delete(
    "/{node_type_id}",
    status_code=204,
    summary="Delete a node type",
    description="Delete a node type by its ID.",
    responses={
        204: {"description": "Node type deleted successfully"},
        404: {"description": "Node type or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def delete_node_type(tenant_id: str, node_type_id: str):
    """Delete a node type."""
    try:
        if _nodetype_service is None:
            raise RuntimeError("Node type service not initialized")
        await _nodetype_service.delete(tenant_id, node_type_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "",
    response_model=NodeTypeListResponse,
    summary="List node types",
    description="List all node types within a tenant with pagination.",
    responses={
        200: {"description": "List of node types"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_node_types(
    tenant_id: str,
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List node types for a tenant."""
    try:
        if _nodetype_service is None:
            raise RuntimeError("Node type service not initialized")
        node_types, pagination = await _nodetype_service.list(tenant_id, page_size, page_token)
        return NodeTypeListResponse(
            node_types=[nt.to_dict() for nt in node_types],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

