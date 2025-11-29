package service

import (
	"context"
	"fmt"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// NodeTypeService handles node type business logic
type NodeTypeService struct {
	repo repository.NodeTypeRepository
}

// NewNodeTypeService creates a new NodeTypeService
func NewNodeTypeService(repo repository.NodeTypeRepository) *NodeTypeService {
	return &NodeTypeService{repo: repo}
}

// Create creates a new node type
func (s *NodeTypeService) Create(ctx context.Context, tenantID, name, description, schema string) (*repository.NodeType, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	nodeType := &repository.NodeType{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Schema:      schema,
	}

	return s.repo.Create(ctx, nodeType)
}

// GetByID retrieves a node type by ID
func (s *NodeTypeService) GetByID(ctx context.Context, tenantID, id string) (*repository.NodeType, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates an existing node type
func (s *NodeTypeService) Update(ctx context.Context, tenantID, id, name, description, schema string) (*repository.NodeType, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	nodeType, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		nodeType.Name = name
	}
	if description != "" {
		nodeType.Description = description
	}
	if schema != "" {
		nodeType.Schema = schema
	}

	return s.repo.Update(ctx, nodeType)
}

// Delete deletes a node type
func (s *NodeTypeService) Delete(ctx context.Context, tenantID, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// List retrieves node types with pagination
func (s *NodeTypeService) List(ctx context.Context, tenantID string, pageSize int32, pageToken string) ([]*repository.NodeType, *repository.ListResult, error) {
	if tenantID == "" {
		return nil, nil, fmt.Errorf("tenant_id is required")
	}

	opts := repository.ListOptions{
		PageSize:  int(pageSize),
		PageToken: pageToken,
	}
	return s.repo.List(ctx, tenantID, opts)
}
