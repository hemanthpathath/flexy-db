"""
User REST API router.
"""

from fastapi import APIRouter, HTTPException, Query

from app.api.models import (
    UserCreate,
    UserUpdate,
    UserResponse,
    UserListResponse,
    TenantUserAdd,
    TenantUserResponse,
    TenantUserListResponse,
    ErrorResponse,
)
from app.api.errors import handle_service_error

# Service instances will be set by main.py
_user_service = None


def set_user_service(service):
    """Set the user service instance."""
    global _user_service
    _user_service = service


router = APIRouter(prefix="/users", tags=["Users"])


@router.post(
    "",
    response_model=UserResponse,
    status_code=201,
    summary="Create a user",
    description="Create a new user with the given email and display name.",
    responses={
        201: {"description": "User created successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def create_user(user: UserCreate):
    """Create a new user."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        user_obj = await _user_service.create(user.email, user.display_name)
        return UserResponse(user=user_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "/{user_id}",
    response_model=UserResponse,
    summary="Get a user",
    description="Get a user by their ID.",
    responses={
        200: {"description": "User found"},
        404: {"description": "User not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def get_user(user_id: str):
    """Get a user by ID."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        user_obj = await _user_service.get_by_id(user_id)
        return UserResponse(user=user_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.put(
    "/{user_id}",
    response_model=UserResponse,
    summary="Update a user",
    description="Update an existing user. Only provided fields will be updated.",
    responses={
        200: {"description": "User updated successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "User not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def update_user(user_id: str, user: UserUpdate):
    """Update an existing user."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        email = user.email if user.email is not None else ""
        display_name = user.display_name if user.display_name is not None else ""
        user_obj = await _user_service.update(user_id, email, display_name)
        return UserResponse(user=user_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@router.delete(
    "/{user_id}",
    status_code=204,
    summary="Delete a user",
    description="Delete a user by their ID.",
    responses={
        204: {"description": "User deleted successfully"},
        404: {"description": "User not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def delete_user(user_id: str):
    """Delete a user."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        await _user_service.delete(user_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


@router.get(
    "",
    response_model=UserListResponse,
    summary="List users",
    description="List all users with pagination.",
    responses={
        200: {"description": "List of users"},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_users(
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List users with pagination."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        users, pagination = await _user_service.list(page_size, page_token)
        return UserListResponse(
            users=[u.to_dict() for u in users],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)


# ============================================================================
# Tenant-User membership endpoints
# ============================================================================

tenant_users_router = APIRouter(prefix="/tenants/{tenant_id}/users", tags=["Tenant Users"])


@tenant_users_router.post(
    "",
    response_model=TenantUserResponse,
    status_code=201,
    summary="Add user to tenant",
    description="Add a user to a tenant with the specified role.",
    responses={
        201: {"description": "User added to tenant successfully"},
        400: {"description": "Invalid parameters", "model": ErrorResponse},
        404: {"description": "Tenant or user not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def add_user_to_tenant(tenant_id: str, tenant_user: TenantUserAdd):
    """Add a user to a tenant."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        tenant_user_obj = await _user_service.add_to_tenant(tenant_id, tenant_user.user_id, tenant_user.role)
        return TenantUserResponse(tenant_user=tenant_user_obj.to_dict())
    except Exception as e:
        raise handle_service_error(e)


@tenant_users_router.delete(
    "/{user_id}",
    status_code=204,
    summary="Remove user from tenant",
    description="Remove a user from a tenant.",
    responses={
        204: {"description": "User removed from tenant successfully"},
        404: {"description": "Tenant or user not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def remove_user_from_tenant(tenant_id: str, user_id: str):
    """Remove a user from a tenant."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        await _user_service.remove_from_tenant(tenant_id, user_id)
        return None
    except Exception as e:
        raise handle_service_error(e)


@tenant_users_router.get(
    "",
    response_model=TenantUserListResponse,
    summary="List tenant users",
    description="List all users in a tenant with pagination.",
    responses={
        200: {"description": "List of tenant users"},
        404: {"description": "Tenant not found", "model": ErrorResponse},
        500: {"description": "Internal server error", "model": ErrorResponse},
    },
)
async def list_tenant_users(
    tenant_id: str,
    page_size: int = Query(default=10, ge=1, le=100, description="Number of items per page"),
    page_token: str = Query(default="", description="Token for the next page"),
):
    """List users in a tenant."""
    try:
        if _user_service is None:
            raise RuntimeError("User service not initialized")
        tenant_users, pagination = await _user_service.list_tenant_users(tenant_id, page_size, page_token)
        return TenantUserListResponse(
            tenant_users=[tu.to_dict() for tu in tenant_users],
            pagination=pagination.to_dict()
        )
    except Exception as e:
        raise handle_service_error(e)

