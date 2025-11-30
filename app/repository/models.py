"""
Repository models module.
"""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional


@dataclass
class Tenant:
    """Tenant entity."""
    id: str = ""
    slug: str = ""
    name: str = ""
    status: str = "active"
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "slug": self.slug,
            "name": self.name,
            "status": self.status,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class User:
    """User entity."""
    id: str = ""
    email: str = ""
    display_name: str = ""
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "email": self.email,
            "display_name": self.display_name,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class TenantUser:
    """User's membership in a tenant."""
    tenant_id: str = ""
    user_id: str = ""
    role: str = "member"
    status: str = "active"

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "tenant_id": self.tenant_id,
            "user_id": self.user_id,
            "role": self.role,
            "status": self.status,
        }


@dataclass
class NodeType:
    """Node type entity."""
    id: str = ""
    tenant_id: str = ""
    name: str = ""
    description: str = ""
    schema: str = ""  # JSON string
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "tenant_id": self.tenant_id,
            "name": self.name,
            "description": self.description,
            "schema": self.schema,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class Node:
    """Node entity."""
    id: str = ""
    tenant_id: str = ""
    node_type_id: str = ""
    data: str = "{}"  # JSON string
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "tenant_id": self.tenant_id,
            "node_type_id": self.node_type_id,
            "data": self.data,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class Relationship:
    """Relationship between nodes."""
    id: str = ""
    tenant_id: str = ""
    source_node_id: str = ""
    target_node_id: str = ""
    relationship_type: str = ""
    data: str = "{}"  # JSON string
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "tenant_id": self.tenant_id,
            "source_node_id": self.source_node_id,
            "target_node_id": self.target_node_id,
            "relationship_type": self.relationship_type,
            "data": self.data,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class ListOptions:
    """Common pagination options."""
    page_size: int = 10
    page_token: str = ""


@dataclass
class ListResult:
    """Common pagination result metadata."""
    next_page_token: str = ""
    total_count: int = 0

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "next_page_token": self.next_page_token,
            "total_count": self.total_count,
        }
