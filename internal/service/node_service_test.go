package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/internal/repository"
)

// mockNodeRepository is a mock implementation of NodeRepository
type mockNodeRepository struct {
	nodes map[string]*repository.Node // key: tenantID:id
	err   error
}

func newMockNodeRepository() *mockNodeRepository {
	return &mockNodeRepository{
		nodes: make(map[string]*repository.Node),
	}
}

func (m *mockNodeRepository) Create(ctx context.Context, node *repository.Node) (*repository.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	node.ID = "node-" + node.NodeTypeID
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	key := node.TenantID + ":" + node.ID
	m.nodes[key] = node
	return node, nil
}

func (m *mockNodeRepository) GetByID(ctx context.Context, tenantID, id string) (*repository.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := tenantID + ":" + id
	node, ok := m.nodes[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return node, nil
}

func (m *mockNodeRepository) Update(ctx context.Context, node *repository.Node) (*repository.Node, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := node.TenantID + ":" + node.ID
	existing, ok := m.nodes[key]
	if !ok {
		return nil, errors.New("not found")
	}
	node.UpdatedAt = time.Now()
	node.CreatedAt = existing.CreatedAt
	m.nodes[key] = node
	return node, nil
}

func (m *mockNodeRepository) Delete(ctx context.Context, tenantID, id string) error {
	if m.err != nil {
		return m.err
	}
	key := tenantID + ":" + id
	if _, ok := m.nodes[key]; !ok {
		return errors.New("not found")
	}
	delete(m.nodes, key)
	return nil
}

func (m *mockNodeRepository) List(ctx context.Context, tenantID, nodeTypeID string, opts repository.ListOptions) ([]*repository.Node, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var nodes []*repository.Node
	for _, node := range m.nodes {
		if node.TenantID == tenantID {
			if nodeTypeID == "" || node.NodeTypeID == nodeTypeID {
				nodes = append(nodes, node)
			}
		}
	}
	return nodes, &repository.ListResult{TotalCount: len(nodes)}, nil
}

func TestNodeService_Create(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		nodeTypeID    string
		data          string
		setupNodeType *repository.NodeType
		nodeTypeErr   error
		expectedError string
	}{
		{
			name:       "successful creation",
			tenantID:   "tenant-1",
			nodeTypeID: "nodetype-1",
			data:       `{"title": "Task 1"}`,
			setupNodeType: &repository.NodeType{
				ID:       "nodetype-1",
				TenantID: "tenant-1",
				Name:     "Task",
			},
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			nodeTypeID:    "nodetype-1",
			expectedError: "tenant_id is required",
		},
		{
			name:          "empty node type id",
			tenantID:      "tenant-1",
			nodeTypeID:    "",
			expectedError: "node_type_id is required",
		},
		{
			name:          "node type not found",
			tenantID:      "tenant-1",
			nodeTypeID:    "non-existent",
			nodeTypeErr:   errors.New("not found"),
			expectedError: "invalid node_type_id: node type not found or does not belong to this tenant",
		},
		{
			name:       "node type belongs to different tenant",
			tenantID:   "tenant-1",
			nodeTypeID: "nodetype-1",
			setupNodeType: &repository.NodeType{
				ID:       "nodetype-1",
				TenantID: "tenant-2", // Different tenant
				Name:     "Task",
			},
			expectedError: "invalid node_type_id: node type not found or does not belong to this tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeRepo := newMockNodeRepository()
			mockNodeTypeRepo := newMockNodeTypeRepository()
			mockNodeTypeRepo.err = tt.nodeTypeErr
			
			if tt.setupNodeType != nil {
				key := tt.setupNodeType.TenantID + ":" + tt.setupNodeType.ID
				mockNodeTypeRepo.nodeTypes[key] = tt.setupNodeType
			}

			service := NewNodeService(mockNodeRepo, mockNodeTypeRepo)

			ctx := context.Background()
			node, err := service.Create(ctx, tt.tenantID, tt.nodeTypeID, tt.data)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if node == nil {
					t.Fatal("expected node, got nil")
				}
				if node.TenantID != tt.tenantID {
					t.Errorf("expected tenant ID %q, got %q", tt.tenantID, node.TenantID)
				}
				if node.NodeTypeID != tt.nodeTypeID {
					t.Errorf("expected node type ID %q, got %q", tt.nodeTypeID, node.NodeTypeID)
				}
			}
		})
	}
}

func TestNodeService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		setupNode     *repository.Node
		expectedError string
	}{
		{
			name:     "successful retrieval",
			tenantID: "tenant-1",
			id:       "node-1",
			setupNode: &repository.Node{
				ID:         "node-1",
				TenantID:   "tenant-1",
				NodeTypeID: "nodetype-1",
			},
		},
		{
			name:          "empty id",
			tenantID:      "tenant-1",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			id:            "node-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeRepo := newMockNodeRepository()
			mockNodeTypeRepo := newMockNodeTypeRepository()
			
			if tt.setupNode != nil {
				key := tt.setupNode.TenantID + ":" + tt.setupNode.ID
				mockNodeRepo.nodes[key] = tt.setupNode
			}

			service := NewNodeService(mockNodeRepo, mockNodeTypeRepo)

			ctx := context.Background()
			node, err := service.GetByID(ctx, tt.tenantID, tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if node == nil {
					t.Fatal("expected node, got nil")
				}
			}
		})
	}
}

func TestNodeService_Update(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		data          string
		setupNode     *repository.Node
		expectedError string
	}{
		{
			name:     "successful update",
			tenantID: "tenant-1",
			id:       "node-1",
			data:     `{"title": "Updated Task"}`,
			setupNode: &repository.Node{
				ID:         "node-1",
				TenantID:   "tenant-1",
				NodeTypeID: "nodetype-1",
				Data:       `{"title": "Original Task"}`,
			},
		},
		{
			name:          "empty id",
			tenantID:      "tenant-1",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			id:            "node-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeRepo := newMockNodeRepository()
			mockNodeTypeRepo := newMockNodeTypeRepository()

			if tt.setupNode != nil {
				key := tt.setupNode.TenantID + ":" + tt.setupNode.ID
				mockNodeRepo.nodes[key] = tt.setupNode
			}

			service := NewNodeService(mockNodeRepo, mockNodeTypeRepo)

			ctx := context.Background()
			node, err := service.Update(ctx, tt.tenantID, tt.id, tt.data)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if node == nil {
					t.Fatal("expected node, got nil")
				}
				if tt.data != "" && node.Data != tt.data {
					t.Errorf("expected data %q, got %q", tt.data, node.Data)
				}
			}
		})
	}
}

func TestNodeService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		setupNode     *repository.Node
		expectedError string
	}{
		{
			name:     "successful deletion",
			tenantID: "tenant-1",
			id:       "node-1",
			setupNode: &repository.Node{
				ID:         "node-1",
				TenantID:   "tenant-1",
				NodeTypeID: "nodetype-1",
			},
		},
		{
			name:          "empty id",
			tenantID:      "tenant-1",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			id:            "node-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeRepo := newMockNodeRepository()
			mockNodeTypeRepo := newMockNodeTypeRepository()

			if tt.setupNode != nil {
				key := tt.setupNode.TenantID + ":" + tt.setupNode.ID
				mockNodeRepo.nodes[key] = tt.setupNode
			}

			service := NewNodeService(mockNodeRepo, mockNodeTypeRepo)

			ctx := context.Background()
			err := service.Delete(ctx, tt.tenantID, tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNodeService_List(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		nodeTypeID    string
		setupNodes    []*repository.Node
		expectedCount int
		expectedError string
	}{
		{
			name:       "successful list",
			tenantID:   "tenant-1",
			nodeTypeID: "",
			setupNodes: []*repository.Node{
				{ID: "node-1", TenantID: "tenant-1", NodeTypeID: "nodetype-1"},
				{ID: "node-2", TenantID: "tenant-1", NodeTypeID: "nodetype-2"},
			},
			expectedCount: 2,
		},
		{
			name:       "filtered by node type",
			tenantID:   "tenant-1",
			nodeTypeID: "nodetype-1",
			setupNodes: []*repository.Node{
				{ID: "node-1", TenantID: "tenant-1", NodeTypeID: "nodetype-1"},
				{ID: "node-2", TenantID: "tenant-1", NodeTypeID: "nodetype-2"},
			},
			expectedCount: 1,
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNodeRepo := newMockNodeRepository()
			mockNodeTypeRepo := newMockNodeTypeRepository()

			for _, node := range tt.setupNodes {
				key := node.TenantID + ":" + node.ID
				mockNodeRepo.nodes[key] = node
			}

			service := NewNodeService(mockNodeRepo, mockNodeTypeRepo)

			ctx := context.Background()
			nodes, result, err := service.List(ctx, tt.tenantID, tt.nodeTypeID, 10, "")

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(nodes) != tt.expectedCount {
					t.Errorf("expected %d nodes, got %d", tt.expectedCount, len(nodes))
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
			}
		})
	}
}

