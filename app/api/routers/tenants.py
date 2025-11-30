"""
Tenant REST API router.
"""

from fastapi import APIRouter, Query

from app.api.models import (
    TenantCreate,
    TenantUpdate,
    TenantResponse,
    TenantListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error

# Service instances will be set by main.py
_tenant_service = None


def set_tenant_service(service):
    """Set the tenant service instance."""
    global _tenant_service
    _tenant_service = service


router = APIRouter(prefix="/tenants", tags=["Tenants"])


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
        if _tenant_service is None:
            raise RuntimeError("Tenant service not initialized")
        tenant_obj = await _tenant_service.create(tenant.slug, tenant.name)
        return TenantResponse(tenant=tenant_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


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
        if _tenant_service is None:
            raise RuntimeError("Tenant service not initialized")
        tenant_obj = await _tenant_service.get_by_id(tenant_id)
        return TenantResponse(tenant=tenant_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


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
        if _tenant_service is None:
            raise RuntimeError("Tenant service not initialized")
        # Only pass non-None values to service (service layer handles empty strings)
        slug = tenant.slug or ""
        name = tenant.name or ""
        status = tenant.status or ""
        tenant_obj = await _tenant_service.update(tenant_id, slug, name, status)
        return TenantResponse(tenant=tenant_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


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
        if _tenant_service is None:
            raise RuntimeError("Tenant service not initialized")
        await _tenant_service.delete(tenant_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


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
        if _tenant_service is None:
            raise RuntimeError("Tenant service not initialized")
        tenants, pagination = await _tenant_service.list(page_size, page_token)
        return TenantListResponse(
            tenants=[t.to_dict() for t in tenants],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

