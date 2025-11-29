package repository

import (
	"time"
)

// Tenant represents a tenant entity
type Tenant struct {
	ID        string
	Slug      string
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User represents a user entity
type User struct {
	ID          string
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TenantUser represents a user's membership in a tenant
type TenantUser struct {
	TenantID string
	UserID   string
	Role     string
	Status   string
}

// NodeType represents a node type entity
type NodeType struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Schema      string // JSON string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Node represents a node entity
type Node struct {
	ID         string
	TenantID   string
	NodeTypeID string
	Data       string // JSON string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Relationship represents a relationship between nodes
type Relationship struct {
	ID               string
	TenantID         string
	SourceNodeID     string
	TargetNodeID     string
	RelationshipType string
	Data             string // JSON string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ListOptions contains common pagination options
type ListOptions struct {
	PageSize  int
	PageToken string
}

// ListResult contains common pagination result metadata
type ListResult struct {
	NextPageToken string
	TotalCount    int
}
