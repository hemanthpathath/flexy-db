package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// mockNodeTypeRepository is a mock implementation of NodeTypeRepository
type mockNodeTypeRepository struct {
	nodeTypes map[string]*repository.NodeType // key: tenantID:id
	err       error
}

func newMockNodeTypeRepository() *mockNodeTypeRepository {
	return &mockNodeTypeRepository{
		nodeTypes: make(map[string]*repository.NodeType),
	}
}

func (m *mockNodeTypeRepository) Create(ctx context.Context, nodeType *repository.NodeType) (*repository.NodeType, error) {
	if m.err != nil {
		return nil, m.err
	}
	nodeType.ID = "nodetype-" + nodeType.Name
	nodeType.CreatedAt = time.Now()
	nodeType.UpdatedAt = time.Now()
	key := nodeType.TenantID + ":" + nodeType.ID
	m.nodeTypes[key] = nodeType
	return nodeType, nil
}

func (m *mockNodeTypeRepository) GetByID(ctx context.Context, tenantID, id string) (*repository.NodeType, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := tenantID + ":" + id
	nodeType, ok := m.nodeTypes[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return nodeType, nil
}

func (m *mockNodeTypeRepository) Update(ctx context.Context, nodeType *repository.NodeType) (*repository.NodeType, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := nodeType.TenantID + ":" + nodeType.ID
	existing, ok := m.nodeTypes[key]
	if !ok {
		return nil, errors.New("not found")
	}
	nodeType.UpdatedAt = time.Now()
	nodeType.CreatedAt = existing.CreatedAt
	m.nodeTypes[key] = nodeType
	return nodeType, nil
}

func (m *mockNodeTypeRepository) Delete(ctx context.Context, tenantID, id string) error {
	if m.err != nil {
		return m.err
	}
	key := tenantID + ":" + id
	if _, ok := m.nodeTypes[key]; !ok {
		return errors.New("not found")
	}
	delete(m.nodeTypes, key)
	return nil
}

func (m *mockNodeTypeRepository) List(ctx context.Context, tenantID string, opts repository.ListOptions) ([]*repository.NodeType, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var nodeTypes []*repository.NodeType
	for _, nt := range m.nodeTypes {
		if nt.TenantID == tenantID {
			nodeTypes = append(nodeTypes, nt)
		}
	}
	return nodeTypes, &repository.ListResult{TotalCount: len(nodeTypes)}, nil
}

func TestNodeTypeService_Create(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		nodeTypeName  string
		description   string
		schema        string
		expectedError string
	}{
		{
			name:         "successful creation",
			tenantID:     "tenant-1",
			nodeTypeName: "Task",
			description:  "A task node type",
			schema:       `{"type": "object"}`,
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			nodeTypeName:  "Task",
			expectedError: "tenant_id is required",
		},
		{
			name:          "empty name",
			tenantID:      "tenant-1",
			nodeTypeName:  "",
			expectedError: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockNodeTypeRepository()
			service := NewNodeTypeService(mockRepo)

			ctx := context.Background()
			nodeType, err := service.Create(ctx, tt.tenantID, tt.nodeTypeName, tt.description, tt.schema)

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
				if nodeType == nil {
					t.Fatal("expected node type, got nil")
				}
				if nodeType.TenantID != tt.tenantID {
					t.Errorf("expected tenant ID %q, got %q", tt.tenantID, nodeType.TenantID)
				}
				if nodeType.Name != tt.nodeTypeName {
					t.Errorf("expected name %q, got %q", tt.nodeTypeName, nodeType.Name)
				}
			}
		})
	}
}

func TestNodeTypeService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		setupNodeType *repository.NodeType
		expectedError string
	}{
		{
			name:     "successful retrieval",
			tenantID: "tenant-1",
			id:       "nodetype-1",
			setupNodeType: &repository.NodeType{
				ID:       "nodetype-1",
				TenantID: "tenant-1",
				Name:     "Task",
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
			id:            "nodetype-1",
			expectedError: "tenant_id is required",
		},
		{
			name:          "not found",
			tenantID:      "tenant-1",
			id:            "non-existent",
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockNodeTypeRepository()
			if tt.setupNodeType != nil {
				key := tt.setupNodeType.TenantID + ":" + tt.setupNodeType.ID
				mockRepo.nodeTypes[key] = tt.setupNodeType
			}
			service := NewNodeTypeService(mockRepo)

			ctx := context.Background()
			nodeType, err := service.GetByID(ctx, tt.tenantID, tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if nodeType == nil {
					t.Fatal("expected node type, got nil")
				}
				if nodeType.ID != tt.id {
					t.Errorf("expected ID %q, got %q", tt.id, nodeType.ID)
				}
			}
		})
	}
}

func TestNodeTypeService_Update(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		updateName    string
		description   string
		schema        string
		setupNodeType *repository.NodeType
		expectedError string
	}{
		{
			name:       "successful update",
			tenantID:   "tenant-1",
			id:         "nodetype-1",
			updateName: "Updated Task",
			setupNodeType: &repository.NodeType{
				ID:       "nodetype-1",
				TenantID: "tenant-1",
				Name:     "Task",
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
			id:            "nodetype-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockNodeTypeRepository()
			if tt.setupNodeType != nil {
				key := tt.setupNodeType.TenantID + ":" + tt.setupNodeType.ID
				mockRepo.nodeTypes[key] = tt.setupNodeType
			}
			service := NewNodeTypeService(mockRepo)

			ctx := context.Background()
			nodeType, err := service.Update(ctx, tt.tenantID, tt.id, tt.updateName, tt.description, tt.schema)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if nodeType == nil {
					t.Fatal("expected node type, got nil")
				}
				if tt.updateName != "" && nodeType.Name != tt.updateName {
					t.Errorf("expected name %q, got %q", tt.updateName, nodeType.Name)
				}
			}
		})
	}
}

