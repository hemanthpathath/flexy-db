"""
Node REST API router.
"""

from fastapi import APIRouter, Query
from typing import Optional

from app.api.models import (
    NodeCreate,
    NodeUpdate,
    NodeResponse,
    NodeListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error
from app.api.dependencies import resolve_tenant_services


router = APIRouter(prefix="/tenants/{tenant_id}/nodes", tags=["Nodes"])


@router.post(
    "",
    response_model=NodeResponse,
    status_code=201,
    summary="Create a node",
    description="Create a new node within a tenant.",
    responses={
        201: {"description": "Node created successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Tenant or node type not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def create_node(tenant_id: str, node: NodeCreate):
    """Create a new node."""
    try:
        services = await resolve_tenant_services(tenant_id)
        node_obj = await services["node"].create(node.node_type_id, node.data or "{}")
        return NodeResponse(node=node_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "/{node_id}",
    response_model=NodeResponse,
    summary="Get a node",
    description="Get a node by its ID within a tenant.",
    responses={
        200: {"description": "Node found"},
        404: {"description": "Node or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def get_node(tenant_id: str, node_id: str):
    """Get a node by ID."""
    try:
        services = await resolve_tenant_services(tenant_id)
        node_obj = await services["node"].get_by_id(node_id)
        return NodeResponse(node=node_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.put(
    "/{node_id}",
    response_model=NodeResponse,
    summary="Update a node",
    description="Update an existing node. Only provided fields will be updated.",
    responses={
        200: {"description": "Node updated successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Node or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def update_node(tenant_id: str, node_id: str, node: NodeUpdate):
    """Update an existing node."""
    try:
        services = await resolve_tenant_services(tenant_id)
        # Only pass non-None values to service (service layer handles empty strings)
        data = node.data or ""
        node_obj = await services["node"].update(node_id, data)
        return NodeResponse(node=node_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.delete(
    "/{node_id}",
    status_code=204,
    summary="Delete a node",
    description="Delete a node by its ID.",
    responses={
        204: {"description": "Node deleted successfully"},
        404: {"description": "Node or tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def delete_node(tenant_id: str, node_id: str):
    """Delete a node."""
    try:
        services = await resolve_tenant_services(tenant_id)
        await services["node"].delete(node_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "",
    response_model=NodeListResponse,
    summary="List nodes",
    description="List all nodes within a tenant with optional filtering by node type.",
    responses={
        200: {"description": "List of nodes"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_nodes(
    tenant_id: str,
    node_type_id: Optional[str] = Query(default=None, description="Filter by node type ID"),
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List nodes for a tenant."""
    try:
        services = await resolve_tenant_services(tenant_id)
        nodes, pagination = await services["node"].list(node_type_id or None, page_size, page_token)
        return NodeListResponse(
            nodes=[n.to_dict() for n in nodes],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

