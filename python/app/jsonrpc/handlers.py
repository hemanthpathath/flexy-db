"""
JSON-RPC handlers for all services.
"""

from typing import Any, Dict, List, Optional
from jsonrpcserver import method, Result, Success, Error

from app.service import (
    TenantService,
    UserService,
    NodeTypeService,
    NodeService,
    RelationshipService,
)
from app.repository.errors import NotFoundError

# Global service instances (to be set by register_methods)
_tenant_service: Optional[TenantService] = None
_user_service: Optional[UserService] = None
_nodetype_service: Optional[NodeTypeService] = None
_node_service: Optional[NodeService] = None
_relationship_service: Optional[RelationshipService] = None


def register_methods(
    tenant_svc: TenantService,
    user_svc: UserService,
    nodetype_svc: NodeTypeService,
    node_svc: NodeService,
    relationship_svc: RelationshipService,
) -> None:
    """Register service instances for use by JSON-RPC methods."""
    global _tenant_service, _user_service, _nodetype_service, _node_service, _relationship_service
    _tenant_service = tenant_svc
    _user_service = user_svc
    _nodetype_service = nodetype_svc
    _node_service = node_svc
    _relationship_service = relationship_svc


def _handle_error(err: Exception) -> Error:
    """Convert exception to JSON-RPC error."""
    if isinstance(err, NotFoundError):
        return Error(-32001, str(err))
    if isinstance(err, ValueError):
        return Error(-32602, str(err))
    return Error(-32603, str(err))


# ============================================================================
# Tenant Service Methods
# ============================================================================

@method
async def create_tenant(slug: str, name: str) -> Result:
    """Create a new tenant."""
    try:
        tenant = await _tenant_service.create(slug, name)
        return Success({"tenant": tenant.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def get_tenant(id: str) -> Result:
    """Get a tenant by ID."""
    try:
        tenant = await _tenant_service.get_by_id(id)
        return Success({"tenant": tenant.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def update_tenant(id: str, slug: str = "", name: str = "", status: str = "") -> Result:
    """Update an existing tenant."""
    try:
        tenant = await _tenant_service.update(id, slug, name, status)
        return Success({"tenant": tenant.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def delete_tenant(id: str) -> Result:
    """Delete a tenant."""
    try:
        await _tenant_service.delete(id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_tenants(pagination: Dict[str, Any] = None) -> Result:
    """List tenants with pagination."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        tenants, result = await _tenant_service.list(page_size, page_token)
        return Success({
            "tenants": [t.to_dict() for t in tenants],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


# ============================================================================
# User Service Methods
# ============================================================================

@method
async def create_user(email: str, display_name: str) -> Result:
    """Create a new user."""
    try:
        user = await _user_service.create(email, display_name)
        return Success({"user": user.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def get_user(id: str) -> Result:
    """Get a user by ID."""
    try:
        user = await _user_service.get_by_id(id)
        return Success({"user": user.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def update_user(id: str, email: str = "", display_name: str = "") -> Result:
    """Update an existing user."""
    try:
        user = await _user_service.update(id, email, display_name)
        return Success({"user": user.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def delete_user(id: str) -> Result:
    """Delete a user."""
    try:
        await _user_service.delete(id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_users(pagination: Dict[str, Any] = None) -> Result:
    """List users with pagination."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        users, result = await _user_service.list(page_size, page_token)
        return Success({
            "users": [u.to_dict() for u in users],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


@method
async def add_user_to_tenant(tenant_id: str, user_id: str, role: str = "") -> Result:
    """Add a user to a tenant."""
    try:
        tenant_user = await _user_service.add_to_tenant(tenant_id, user_id, role)
        return Success({"tenant_user": tenant_user.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def remove_user_from_tenant(tenant_id: str, user_id: str) -> Result:
    """Remove a user from a tenant."""
    try:
        await _user_service.remove_from_tenant(tenant_id, user_id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_tenant_users(tenant_id: str, pagination: Dict[str, Any] = None) -> Result:
    """List users in a tenant."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        tenant_users, result = await _user_service.list_tenant_users(tenant_id, page_size, page_token)
        return Success({
            "tenant_users": [tu.to_dict() for tu in tenant_users],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


# ============================================================================
# NodeType Service Methods
# ============================================================================

@method
async def create_node_type(tenant_id: str, name: str, description: str = "", schema: str = "") -> Result:
    """Create a new node type."""
    try:
        node_type = await _nodetype_service.create(tenant_id, name, description, schema)
        return Success({"node_type": node_type.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def get_node_type(id: str, tenant_id: str) -> Result:
    """Get a node type by ID."""
    try:
        node_type = await _nodetype_service.get_by_id(tenant_id, id)
        return Success({"node_type": node_type.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def update_node_type(id: str, tenant_id: str, name: str = "", description: str = "", schema: str = "") -> Result:
    """Update an existing node type."""
    try:
        node_type = await _nodetype_service.update(tenant_id, id, name, description, schema)
        return Success({"node_type": node_type.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def delete_node_type(id: str, tenant_id: str) -> Result:
    """Delete a node type."""
    try:
        await _nodetype_service.delete(tenant_id, id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_node_types(tenant_id: str, pagination: Dict[str, Any] = None) -> Result:
    """List node types for a tenant."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        node_types, result = await _nodetype_service.list(tenant_id, page_size, page_token)
        return Success({
            "node_types": [nt.to_dict() for nt in node_types],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


# ============================================================================
# Node Service Methods
# ============================================================================

@method
async def create_node(tenant_id: str, node_type_id: str, data: str = "{}") -> Result:
    """Create a new node."""
    try:
        node = await _node_service.create(tenant_id, node_type_id, data)
        return Success({"node": node.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def get_node(id: str, tenant_id: str) -> Result:
    """Get a node by ID."""
    try:
        node = await _node_service.get_by_id(tenant_id, id)
        return Success({"node": node.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def update_node(id: str, tenant_id: str, data: str = "") -> Result:
    """Update an existing node."""
    try:
        node = await _node_service.update(tenant_id, id, data)
        return Success({"node": node.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def delete_node(id: str, tenant_id: str) -> Result:
    """Delete a node."""
    try:
        await _node_service.delete(tenant_id, id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_nodes(tenant_id: str, node_type_id: str = "", pagination: Dict[str, Any] = None) -> Result:
    """List nodes for a tenant with optional filtering."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        nodes, result = await _node_service.list(tenant_id, node_type_id or None, page_size, page_token)
        return Success({
            "nodes": [n.to_dict() for n in nodes],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


# ============================================================================
# Relationship Service Methods
# ============================================================================

@method
async def create_relationship(
    tenant_id: str,
    source_node_id: str,
    target_node_id: str,
    relationship_type: str,
    data: str = "{}"
) -> Result:
    """Create a new relationship."""
    try:
        rel = await _relationship_service.create(tenant_id, source_node_id, target_node_id, relationship_type, data)
        return Success({"relationship": rel.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def get_relationship(id: str, tenant_id: str) -> Result:
    """Get a relationship by ID."""
    try:
        rel = await _relationship_service.get_by_id(tenant_id, id)
        return Success({"relationship": rel.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def update_relationship(id: str, tenant_id: str, relationship_type: str = "", data: str = "") -> Result:
    """Update an existing relationship."""
    try:
        rel = await _relationship_service.update(tenant_id, id, relationship_type, data)
        return Success({"relationship": rel.to_dict()})
    except Exception as e:
        return _handle_error(e)


@method
async def delete_relationship(id: str, tenant_id: str) -> Result:
    """Delete a relationship."""
    try:
        await _relationship_service.delete(tenant_id, id)
        return Success({})
    except Exception as e:
        return _handle_error(e)


@method
async def list_relationships(
    tenant_id: str,
    source_node_id: str = "",
    target_node_id: str = "",
    relationship_type: str = "",
    pagination: Dict[str, Any] = None
) -> Result:
    """List relationships for a tenant with optional filtering."""
    try:
        page_size = 10
        page_token = ""
        if pagination:
            page_size = pagination.get("page_size", 10)
            page_token = pagination.get("page_token", "")
        
        rels, result = await _relationship_service.list(
            tenant_id,
            source_node_id or None,
            target_node_id or None,
            relationship_type or None,
            page_size,
            page_token
        )
        return Success({
            "relationships": [r.to_dict() for r in rels],
            "pagination": result.to_dict(),
        })
    except Exception as e:
        return _handle_error(e)


# ============================================================================
# RPC Discovery Methods (OpenRPC Introspection)
# ============================================================================

@method
async def rpc_discover() -> Result:
    """
    Discover available JSON-RPC methods and their schemas.
    
    This method implements OpenRPC introspection, allowing clients to
    dynamically discover all available methods, their parameters, and
    return types.
    
    Returns:
        openrpc: OpenRPC specification object containing all available methods
    
    Note: This method is registered as "rpc_discover" but the OpenRPC spec
    shows it as "rpc.discover" (with dot) for standards compliance.
    """
    try:
        from app.jsonrpc.openrpc import generate_openrpc_spec
        spec = generate_openrpc_spec()
        return Success({"openrpc": spec})
    except Exception as e:
        return _handle_error(e)
