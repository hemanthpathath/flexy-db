"""
User REST API router.
"""

from fastapi import APIRouter, HTTPException, Query

from rest_wrapper.client import get_client, JSONRPCError
from rest_wrapper.models import (
    UserCreate,
    UserUpdate,
    UserResponse,
    UserListResponse,
    TenantUserAdd,
    TenantUserResponse,
    TenantUserListResponse,
    ErrorResponse,
)

router = APIRouter(prefix="/users", tags=["Users"])


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
        client = get_client()
        result = await client.call("create_user", {
            "email": user.email,
            "display_name": user.display_name,
        })
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        result = await client.call("get_user", {"id": user_id})
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        params = {"id": user_id}
        if user.email is not None:
            params["email"] = user.email
        if user.display_name is not None:
            params["display_name"] = user.display_name
        result = await client.call("update_user", params)
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        await client.call("delete_user", {"id": user_id})
        return None
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        params = {"pagination": {"page_size": page_size, "page_token": page_token}}
        result = await client.call("list_users", params)
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        result = await client.call("add_user_to_tenant", {
            "tenant_id": tenant_id,
            "user_id": tenant_user.user_id,
            "role": tenant_user.role,
        })
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        await client.call("remove_user_from_tenant", {
            "tenant_id": tenant_id,
            "user_id": user_id,
        })
        return None
    except JSONRPCError as e:
        raise _handle_rpc_error(e)


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
        client = get_client()
        params = {
            "tenant_id": tenant_id,
            "pagination": {"page_size": page_size, "page_token": page_token},
        }
        result = await client.call("list_tenant_users", params)
        return result
    except JSONRPCError as e:
        raise _handle_rpc_error(e)
