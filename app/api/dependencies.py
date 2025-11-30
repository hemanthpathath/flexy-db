"""
FastAPI dependencies for tenant-scoped services.

These dependencies resolve tenant databases and create tenant-scoped service instances
per request.
"""

from typing import Optional
from fastapi import Depends, HTTPException, status

from app.db.database import Database
from app.db.tenant_db_manager import TenantDatabaseManager
from app.repository import (
    NodeRepository,
    NodeTypeRepository,
    RelationshipRepository,
)
from app.service import (
    NodeService,
    NodeTypeService,
    RelationshipService,
)


# Global tenant database manager (set by main.py)
_tenant_db_manager: Optional[TenantDatabaseManager] = None


def set_tenant_db_manager(manager: TenantDatabaseManager) -> None:
    """Set the global tenant database manager."""
    global _tenant_db_manager
    _tenant_db_manager = manager


async def get_tenant_db(tenant_id: str) -> Database:
    """
    Get tenant database connection for a tenant.
    
    This dependency resolves the tenant database from tenant_id.
    """
    if not _tenant_db_manager:
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail="Tenant database manager not initialized"
        )
    
    try:
        return await _tenant_db_manager.get_tenant_db(tenant_id)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e)
        )
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to get tenant database: {str(e)}"
        )


def create_tenant_services(tenant_db: Database):
    """
    Create tenant-scoped service instances.
    
    Args:
        tenant_db: Tenant database connection
        
    Returns:
        Tuple of (NodeTypeService, NodeService, RelationshipService)
    """
    # Create tenant-scoped repositories
    node_type_repo = NodeTypeRepository(tenant_db)
    node_repo = NodeRepository(tenant_db)
    relationship_repo = RelationshipRepository(tenant_db)
    
    # Create tenant-scoped services
    node_type_svc = NodeTypeService(node_type_repo)
    node_svc = NodeService(node_repo, node_type_repo)
    relationship_svc = RelationshipService(relationship_repo, node_repo)
    
    return {
        "node_type": node_type_svc,
        "node": node_svc,
        "relationship": relationship_svc,
    }


# Helper function for route handlers
async def resolve_tenant_services(tenant_id: str) -> dict:
    """
    Resolve tenant services for a given tenant_id.
    
    This is used by route handlers to get tenant-scoped services.
    """
    tenant_db = await get_tenant_db(tenant_id)
    return create_tenant_services(tenant_db)

