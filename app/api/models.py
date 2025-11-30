"""
Pydantic models for request/response validation.
"""

from typing import List, Optional
from pydantic import BaseModel, Field


# ============================================================================
# Pagination Models
# ============================================================================

class PaginationParams(BaseModel):
    """Pagination query parameters."""
    page_size: int = Field(default=10, ge=1, le=100, description="Number of items per page")
    page_token: str = Field(default="", description="Token for the next page")


class PaginationResult(BaseModel):
    """Pagination result metadata."""
    next_page_token: str = Field(default="", description="Token for the next page")
    total_count: int = Field(default=0, ge=0, description="Total number of items")


# ============================================================================
# Tenant Models
# ============================================================================

class TenantBase(BaseModel):
    """Base tenant model."""
    slug: str = Field(..., min_length=1, description="Unique tenant slug")
    name: str = Field(..., min_length=1, description="Tenant name")


class TenantCreate(TenantBase):
    """Request model for creating a tenant."""
    pass


class TenantUpdate(BaseModel):
    """Request model for updating a tenant."""
    slug: Optional[str] = Field(default=None, description="New tenant slug")
    name: Optional[str] = Field(default=None, description="New tenant name")
    status: Optional[str] = Field(default=None, description="New tenant status")


class Tenant(BaseModel):
    """Tenant response model."""
    id: str = Field(..., description="Tenant ID")
    slug: str = Field(..., description="Tenant slug")
    name: str = Field(..., description="Tenant name")
    status: str = Field(..., description="Tenant status")
    created_at: str = Field(..., description="Creation timestamp")
    updated_at: str = Field(..., description="Last update timestamp")


class TenantResponse(BaseModel):
    """Single tenant response wrapper."""
    tenant: Tenant


class TenantListResponse(BaseModel):
    """List tenants response wrapper."""
    tenants: List[Tenant]
    pagination: PaginationResult


# ============================================================================
# User Models
# ============================================================================

class UserBase(BaseModel):
    """Base user model."""
    email: str = Field(..., description="User email address")
    display_name: str = Field(..., description="User display name")


class UserCreate(UserBase):
    """Request model for creating a user."""
    pass


class UserUpdate(BaseModel):
    """Request model for updating a user."""
    email: Optional[str] = Field(default=None, description="New user email")
    display_name: Optional[str] = Field(default=None, description="New user display name")


class User(BaseModel):
    """User response model."""
    id: str = Field(..., description="User ID")
    email: str = Field(..., description="User email")
    display_name: str = Field(..., description="User display name")
    created_at: str = Field(..., description="Creation timestamp")
    updated_at: str = Field(..., description="Last update timestamp")


class UserResponse(BaseModel):
    """Single user response wrapper."""
    user: User


class UserListResponse(BaseModel):
    """List users response wrapper."""
    users: List[User]
    pagination: PaginationResult


# ============================================================================
# TenantUser Models
# ============================================================================

class TenantUserAdd(BaseModel):
    """Request model for adding a user to a tenant."""
    user_id: str = Field(..., description="User ID to add")
    role: str = Field(default="member", description="User role in the tenant")


class TenantUser(BaseModel):
    """Tenant user response model."""
    tenant_id: str = Field(..., description="Tenant ID")
    user_id: str = Field(..., description="User ID")
    role: str = Field(..., description="User role")
    status: str = Field(..., description="Membership status")


class TenantUserResponse(BaseModel):
    """Single tenant user response wrapper."""
    tenant_user: TenantUser


class TenantUserListResponse(BaseModel):
    """List tenant users response wrapper."""
    tenant_users: List[TenantUser]
    pagination: PaginationResult


# ============================================================================
# NodeType Models
# ============================================================================

class NodeTypeBase(BaseModel):
    """Base node type model."""
    name: str = Field(..., min_length=1, description="Node type name")
    description: Optional[str] = Field(default="", description="Node type description")
    json_schema: Optional[str] = Field(default="", alias="schema", description="JSON schema for node data validation")


class NodeTypeCreate(NodeTypeBase):
    """Request model for creating a node type."""
    pass


class NodeTypeUpdate(BaseModel):
    """Request model for updating a node type."""
    name: Optional[str] = Field(default=None, description="New node type name")
    description: Optional[str] = Field(default=None, description="New node type description")
    json_schema: Optional[str] = Field(default=None, alias="schema", description="New JSON schema")


class NodeType(BaseModel):
    """Node type response model."""
    id: str = Field(..., description="Node type ID")
    tenant_id: str = Field(..., description="Tenant ID")
    name: str = Field(..., description="Node type name")
    description: str = Field(..., description="Node type description")
    json_schema: str = Field(..., alias="schema", description="JSON schema")
    created_at: str = Field(..., description="Creation timestamp")
    updated_at: str = Field(..., description="Last update timestamp")


class NodeTypeResponse(BaseModel):
    """Single node type response wrapper."""
    node_type: NodeType


class NodeTypeListResponse(BaseModel):
    """List node types response wrapper."""
    node_types: List[NodeType]
    pagination: PaginationResult


# ============================================================================
# Node Models
# ============================================================================

class NodeBase(BaseModel):
    """Base node model."""
    node_type_id: str = Field(..., description="Node type ID")
    data: Optional[str] = Field(default="{}", description="Node data as JSON string")


class NodeCreate(NodeBase):
    """Request model for creating a node."""
    pass


class NodeUpdate(BaseModel):
    """Request model for updating a node."""
    data: Optional[str] = Field(default=None, description="New node data as JSON string")


class Node(BaseModel):
    """Node response model."""
    id: str = Field(..., description="Node ID")
    tenant_id: str = Field(..., description="Tenant ID")
    node_type_id: str = Field(..., description="Node type ID")
    data: str = Field(..., description="Node data as JSON string")
    created_at: str = Field(..., description="Creation timestamp")
    updated_at: str = Field(..., description="Last update timestamp")


class NodeResponse(BaseModel):
    """Single node response wrapper."""
    node: Node


class NodeListResponse(BaseModel):
    """List nodes response wrapper."""
    nodes: List[Node]
    pagination: PaginationResult


# ============================================================================
# Relationship Models
# ============================================================================

class RelationshipBase(BaseModel):
    """Base relationship model."""
    source_node_id: str = Field(..., description="Source node ID")
    target_node_id: str = Field(..., description="Target node ID")
    relationship_type: str = Field(..., description="Relationship type")
    data: Optional[str] = Field(default="{}", description="Relationship data as JSON string")


class RelationshipCreate(RelationshipBase):
    """Request model for creating a relationship."""
    pass


class RelationshipUpdate(BaseModel):
    """Request model for updating a relationship."""
    relationship_type: Optional[str] = Field(default=None, description="New relationship type")
    data: Optional[str] = Field(default=None, description="New relationship data as JSON string")


class Relationship(BaseModel):
    """Relationship response model."""
    id: str = Field(..., description="Relationship ID")
    tenant_id: str = Field(..., description="Tenant ID")
    source_node_id: str = Field(..., description="Source node ID")
    target_node_id: str = Field(..., description="Target node ID")
    relationship_type: str = Field(..., description="Relationship type")
    data: str = Field(..., description="Relationship data as JSON string")
    created_at: str = Field(..., description="Creation timestamp")
    updated_at: str = Field(..., description="Last update timestamp")


class RelationshipResponse(BaseModel):
    """Single relationship response wrapper."""
    relationship: Relationship


class RelationshipListResponse(BaseModel):
    """List relationships response wrapper."""
    relationships: List[Relationship]
    pagination: PaginationResult


# ============================================================================
# Error Models
# ============================================================================

class ErrorDetail(BaseModel):
    """Error detail model."""
    code: int = Field(..., description="Error code")
    message: str = Field(..., description="Error message")


class ErrorResponse(BaseModel):
    """Error response model."""
    error: ErrorDetail

