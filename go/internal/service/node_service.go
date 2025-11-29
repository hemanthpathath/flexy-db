package service

import (
	"context"
	"fmt"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// NodeService handles node business logic
type NodeService struct {
	repo         repository.NodeRepository
	nodeTypeRepo repository.NodeTypeRepository
}

// NewNodeService creates a new NodeService
func NewNodeService(repo repository.NodeRepository, nodeTypeRepo repository.NodeTypeRepository) *NodeService {
	return &NodeService{repo: repo, nodeTypeRepo: nodeTypeRepo}
}

// Create creates a new node
func (s *NodeService) Create(ctx context.Context, tenantID, nodeTypeID, data string) (*repository.Node, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if nodeTypeID == "" {
		return nil, fmt.Errorf("node_type_id is required")
	}

	// Validate that the node type belongs to the same tenant
	nodeType, err := s.nodeTypeRepo.GetByID(ctx, tenantID, nodeTypeID)
	if err != nil {
		return nil, fmt.Errorf("invalid node_type_id: node type not found or does not belong to this tenant")
	}
	if nodeType.TenantID != tenantID {
		return nil, fmt.Errorf("invalid node_type_id: node type does not belong to this tenant")
	}

	node := &repository.Node{
		TenantID:   tenantID,
		NodeTypeID: nodeTypeID,
		Data:       data,
	}

	return s.repo.Create(ctx, node)
}

// GetByID retrieves a node by ID
func (s *NodeService) GetByID(ctx context.Context, tenantID, id string) (*repository.Node, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	return s.repo.GetByID(ctx, tenantID, id)
}

// Update updates an existing node
func (s *NodeService) Update(ctx context.Context, tenantID, id, data string) (*repository.Node, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	node, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	if data != "" {
		node.Data = data
	}

	return s.repo.Update(ctx, node)
}

// Delete deletes a node
func (s *NodeService) Delete(ctx context.Context, tenantID, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return s.repo.Delete(ctx, tenantID, id)
}

// List retrieves nodes with pagination and optional filtering
func (s *NodeService) List(ctx context.Context, tenantID, nodeTypeID string, pageSize int32, pageToken string) ([]*repository.Node, *repository.ListResult, error) {
	if tenantID == "" {
		return nil, nil, fmt.Errorf("tenant_id is required")
	}

	opts := repository.ListOptions{
		PageSize:  int(pageSize),
		PageToken: pageToken,
	}
	return s.repo.List(ctx, tenantID, nodeTypeID, opts)
}
