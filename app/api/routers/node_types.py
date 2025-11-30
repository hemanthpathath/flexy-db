"""
NodeType REST API router.
"""

from fastapi import APIRouter, Query

from app.api.models import (
    NodeTypeCreate,
    NodeTypeUpdate,
    NodeTypeResponse,
    NodeTypeListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error
from app.api.dependencies import resolve_tenant_services


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
        services = await resolve_tenant_services(tenant_id)
        node_type_obj = await services["node_type"].create(
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
        services = await resolve_tenant_services(tenant_id)
        node_type_obj = await services["node_type"].get_by_id(node_type_id)
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
        services = await resolve_tenant_services(tenant_id)
        # Only pass non-None values to service (service layer handles empty strings)
        name = node_type.name or ""
        description = node_type.description or ""
        schema = node_type.json_schema or ""
        node_type_obj = await services["node_type"].update(node_type_id, name, description, schema)
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
        services = await resolve_tenant_services(tenant_id)
        await services["node_type"].delete(node_type_id)
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
        services = await resolve_tenant_services(tenant_id)
        node_types, pagination = await services["node_type"].list(page_size, page_token)
        return NodeTypeListResponse(
            node_types=[nt.to_dict() for nt in node_types],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

