package repository

import (
	"context"
)

// TenantRepository defines operations for tenant management
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) (*Tenant, error)
	GetByID(ctx context.Context, id string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ListOptions) ([]*Tenant, *ListResult, error)
}

// UserRepository defines operations for user management
type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opts ListOptions) ([]*User, *ListResult, error)
	AddToTenant(ctx context.Context, tenantUser *TenantUser) (*TenantUser, error)
	RemoveFromTenant(ctx context.Context, tenantID, userID string) error
	ListTenantUsers(ctx context.Context, tenantID string, opts ListOptions) ([]*TenantUser, *ListResult, error)
}

// NodeTypeRepository defines operations for node type management
type NodeTypeRepository interface {
	Create(ctx context.Context, nodeType *NodeType) (*NodeType, error)
	GetByID(ctx context.Context, tenantID, id string) (*NodeType, error)
	Update(ctx context.Context, nodeType *NodeType) (*NodeType, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID string, opts ListOptions) ([]*NodeType, *ListResult, error)
}

// NodeRepository defines operations for node management
type NodeRepository interface {
	Create(ctx context.Context, node *Node) (*Node, error)
	GetByID(ctx context.Context, tenantID, id string) (*Node, error)
	Update(ctx context.Context, node *Node) (*Node, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID, nodeTypeID string, opts ListOptions) ([]*Node, *ListResult, error)
}

// RelationshipRepository defines operations for relationship management
type RelationshipRepository interface {
	Create(ctx context.Context, rel *Relationship) (*Relationship, error)
	GetByID(ctx context.Context, tenantID, id string) (*Relationship, error)
	Update(ctx context.Context, rel *Relationship) (*Relationship, error)
	Delete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID, sourceNodeID, targetNodeID, relType string, opts ListOptions) ([]*Relationship, *ListResult, error)
}
