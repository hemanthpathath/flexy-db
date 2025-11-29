"""
Tenant REST API router.
"""

from fastapi import APIRouter, HTTPException, Query

from rest_wrapper.client import get_client, JSONRPCError
from rest_wrapper.models import (
    TenantCreate,
    TenantUpdate,
    TenantResponse,
    TenantListResponse,
    ErrorResponse,
)

router = APIRouter(prefix="/tenants", tags=["Tenants"])


def _handle_rpc_error(e: JSONRPCError) -> HTTPException:
    """Convert JSON-RPC error to HTTP exception."""
    if e.code == -32001:  # Not found
        return HTTPException(status_code=404, detail=e.message)
    elif e.code == -32602:  # Invalid params
        return HTTPException(status_code=400, detail=e.message)
    else:
        return HTTPException(status_code=500, detail=e.message)


@router.post(
    "",
    response_model=TenantResponse,
    status_code=201,
    summary="Create a tenant",
    description="Create a new tenant with the given slug and name.",
    responses={
        201: {"description": "Tenant created successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def create_tenant(tenant: TenantCreate):
    """Create a new tenant."""
    try:
        client = get_client()
        result = await client.call("create_tenant", {"slug": tenant.slug, "name": tenant.name})
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


@router.get(
    "/{tenant_id}",
    response_model=TenantResponse,
    summary="Get a tenant",
    description="Get a tenant by its ID.",
    responses={
        200: {"description": "Tenant found"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def get_tenant(tenant_id: str):
    """Get a tenant by ID."""
    try:
        client = get_client()
        result = await client.call("get_tenant", {"id": tenant_id})
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


@router.put(
    "/{tenant_id}",
    response_model=TenantResponse,
    summary="Update a tenant",
    description="Update an existing tenant. Only provided fields will be updated.",
    responses={
        200: {"description": "Tenant updated successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def update_tenant(tenant_id: str, tenant: TenantUpdate):
    """Update an existing tenant."""
    try:
        client = get_client()
        params = {"id": tenant_id}
        if tenant.slug is not None:
            params["slug"] = tenant.slug
        if tenant.name is not None:
            params["name"] = tenant.name
        if tenant.status is not None:
            params["status"] = tenant.status
        result = await client.call("update_tenant", params)
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


@router.delete(
    "/{tenant_id}",
    status_code=204,
    summary="Delete a tenant",
    description="Delete a tenant by its ID.",
    responses={
        204: {"description": "Tenant deleted successfully"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def delete_tenant(tenant_id: str):
    """Delete a tenant."""
    try:
        client = get_client()
        await client.call("delete_tenant", {"id": tenant_id})
        return None
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


@router.get(
    "",
    response_model=TenantListResponse,
    summary="List tenants",
    description="List all tenants with pagination.",
    responses={
        200: {"description": "List of tenants"},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_tenants(
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List tenants with pagination."""
    try:
        client = get_client()
        params = {"pagination": {"page_size": page_size, "page_token": page_token}}
        result = await client.call("list_tenants", params)
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)
