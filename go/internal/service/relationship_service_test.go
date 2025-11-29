package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// mockRelationshipRepository is a mock implementation of RelationshipRepository
type mockRelationshipRepository struct {
	relationships map[string]*repository.Relationship // key: tenantID:id
	err           error
}

func newMockRelationshipRepository() *mockRelationshipRepository {
	return &mockRelationshipRepository{
		relationships: make(map[string]*repository.Relationship),
	}
}

func (m *mockRelationshipRepository) Create(ctx context.Context, rel *repository.Relationship) (*repository.Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	rel.ID = "rel-1"
	rel.CreatedAt = time.Now()
	rel.UpdatedAt = time.Now()
	key := rel.TenantID + ":" + rel.ID
	m.relationships[key] = rel
	return rel, nil
}

func (m *mockRelationshipRepository) GetByID(ctx context.Context, tenantID, id string) (*repository.Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := tenantID + ":" + id
	rel, ok := m.relationships[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return rel, nil
}

func (m *mockRelationshipRepository) Update(ctx context.Context, rel *repository.Relationship) (*repository.Relationship, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := rel.TenantID + ":" + rel.ID
	existing, ok := m.relationships[key]
	if !ok {
		return nil, errors.New("not found")
	}
	rel.UpdatedAt = time.Now()
	rel.CreatedAt = existing.CreatedAt
	m.relationships[key] = rel
	return rel, nil
}

func (m *mockRelationshipRepository) Delete(ctx context.Context, tenantID, id string) error {
	if m.err != nil {
		return m.err
	}
	key := tenantID + ":" + id
	if _, ok := m.relationships[key]; !ok {
		return errors.New("not found")
	}
	delete(m.relationships, key)
	return nil
}

func (m *mockRelationshipRepository) List(ctx context.Context, tenantID, sourceNodeID, targetNodeID, relType string, opts repository.ListOptions) ([]*repository.Relationship, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var relationships []*repository.Relationship
	for _, rel := range m.relationships {
		if rel.TenantID == tenantID {
			if sourceNodeID == "" || rel.SourceNodeID == sourceNodeID {
				if targetNodeID == "" || rel.TargetNodeID == targetNodeID {
					if relType == "" || rel.RelationshipType == relType {
						relationships = append(relationships, rel)
					}
				}
			}
		}
	}
	return relationships, &repository.ListResult{TotalCount: len(relationships)}, nil
}

func TestRelationshipService_Create(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		sourceNodeID  string
		targetNodeID  string
		relType       string
		data          string
		setupNodes    []*repository.Node
		nodeErr       error
		expectedError string
	}{
		{
			name:         "successful creation",
			tenantID:     "tenant-1",
			sourceNodeID: "node-1",
			targetNodeID: "node-2",
			relType:      "depends_on",
			data:         `{"priority": 1}`,
			setupNodes: []*repository.Node{
				{ID: "node-1", TenantID: "tenant-1"},
				{ID: "node-2", TenantID: "tenant-1"},
			},
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			sourceNodeID:  "node-1",
			targetNodeID:  "node-2",
			relType:       "depends_on",
			expectedError: "tenant_id is required",
		},
		{
			name:          "empty source node id",
			tenantID:      "tenant-1",
			sourceNodeID:  "",
			targetNodeID:  "node-2",
			relType:       "depends_on",
			expectedError: "source_node_id is required",
		},
		{
			name:          "empty target node id",
			tenantID:      "tenant-1",
			sourceNodeID:  "node-1",
			targetNodeID:  "",
			relType:       "depends_on",
			expectedError: "target_node_id is required",
		},
		{
			name:          "empty relationship type",
			tenantID:      "tenant-1",
			sourceNodeID:  "node-1",
			targetNodeID:  "node-2",
			relType:       "",
			expectedError: "relationship_type is required",
		},
		{
			name:          "source node not found",
			tenantID:      "tenant-1",
			sourceNodeID:  "non-existent",
			targetNodeID:  "node-2",
			relType:       "depends_on",
			nodeErr:       errors.New("not found"),
			expectedError: "invalid source_node_id: node not found or does not belong to this tenant",
		},
		{
			name:         "source node belongs to different tenant",
			tenantID:     "tenant-1",
			sourceNodeID: "node-1",
			targetNodeID: "node-2",
			relType:      "depends_on",
			// Don't add node-1 since it belongs to different tenant (will be "not found")
			setupNodes: []*repository.Node{
				{ID: "node-2", TenantID: "tenant-1"},
			},
			expectedError: "invalid source_node_id: node not found or does not belong to this tenant",
		},
		{
			name:         "target node belongs to different tenant",
			tenantID:     "tenant-1",
			sourceNodeID: "node-1",
			targetNodeID: "node-2",
			relType:      "depends_on",
			// Add source node, but not target (different tenant)
			setupNodes: []*repository.Node{
				{ID: "node-1", TenantID: "tenant-1"},
				// node-2 not added because it belongs to tenant-2
			},
			expectedError: "invalid target_node_id: node not found or does not belong to this tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelRepo := newMockRelationshipRepository()
			mockNodeRepo := newMockNodeRepository()
			mockNodeRepo.err = tt.nodeErr

			// Add only nodes that belong to the correct tenant
			// Nodes with different tenantID won't be found (simulating repository behavior)
			for _, node := range tt.setupNodes {
				if node.TenantID == tt.tenantID {
					key := node.TenantID + ":" + node.ID
					mockNodeRepo.nodes[key] = node
				}
			}

			service := NewRelationshipService(mockRelRepo, mockNodeRepo)

			ctx := context.Background()
			rel, err := service.Create(ctx, tt.tenantID, tt.sourceNodeID, tt.targetNodeID, tt.relType, tt.data)

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
				if rel == nil {
					t.Fatal("expected relationship, got nil")
				}
				if rel.TenantID != tt.tenantID {
					t.Errorf("expected tenant ID %q, got %q", tt.tenantID, rel.TenantID)
				}
				if rel.SourceNodeID != tt.sourceNodeID {
					t.Errorf("expected source node ID %q, got %q", tt.sourceNodeID, rel.SourceNodeID)
				}
				if rel.TargetNodeID != tt.targetNodeID {
					t.Errorf("expected target node ID %q, got %q", tt.targetNodeID, rel.TargetNodeID)
				}
				if rel.RelationshipType != tt.relType {
					t.Errorf("expected relationship type %q, got %q", tt.relType, rel.RelationshipType)
				}
			}
		})
	}
}

func TestRelationshipService_GetByID(t *testing.T) {
	tests := []struct {
		name            string
		tenantID        string
		id              string
		setupRel        *repository.Relationship
		expectedError   string
	}{
		{
			name:     "successful retrieval",
			tenantID: "tenant-1",
			id:       "rel-1",
			setupRel: &repository.Relationship{
				ID:       "rel-1",
				TenantID: "tenant-1",
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
			id:            "rel-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelRepo := newMockRelationshipRepository()
			mockNodeRepo := newMockNodeRepository()

			if tt.setupRel != nil {
				key := tt.setupRel.TenantID + ":" + tt.setupRel.ID
				mockRelRepo.relationships[key] = tt.setupRel
			}

			service := NewRelationshipService(mockRelRepo, mockNodeRepo)

			ctx := context.Background()
			rel, err := service.GetByID(ctx, tt.tenantID, tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if rel == nil {
					t.Fatal("expected relationship, got nil")
				}
			}
		})
	}
}

func TestRelationshipService_Update(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		relType       string
		data          string
		setupRel      *repository.Relationship
		expectedError string
	}{
		{
			name:    "successful update",
			tenantID: "tenant-1",
			id:      "rel-1",
			relType: "updated_type",
			data:    `{"priority": 2}`,
			setupRel: &repository.Relationship{
				ID:               "rel-1",
				TenantID:         "tenant-1",
				RelationshipType: "depends_on",
				Data:             `{"priority": 1}`,
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
			id:            "rel-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelRepo := newMockRelationshipRepository()
			mockNodeRepo := newMockNodeRepository()

			if tt.setupRel != nil {
				key := tt.setupRel.TenantID + ":" + tt.setupRel.ID
				mockRelRepo.relationships[key] = tt.setupRel
			}

			service := NewRelationshipService(mockRelRepo, mockNodeRepo)

			ctx := context.Background()
			rel, err := service.Update(ctx, tt.tenantID, tt.id, tt.relType, tt.data)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if rel == nil {
					t.Fatal("expected relationship, got nil")
				}
				if tt.relType != "" && rel.RelationshipType != tt.relType {
					t.Errorf("expected relationship type %q, got %q", tt.relType, rel.RelationshipType)
				}
			}
		})
	}
}

func TestRelationshipService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		id            string
		setupRel      *repository.Relationship
		expectedError string
	}{
		{
			name:     "successful deletion",
			tenantID: "tenant-1",
			id:       "rel-1",
			setupRel: &repository.Relationship{
				ID:       "rel-1",
				TenantID: "tenant-1",
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
			id:            "rel-1",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelRepo := newMockRelationshipRepository()
			mockNodeRepo := newMockNodeRepository()

			if tt.setupRel != nil {
				key := tt.setupRel.TenantID + ":" + tt.setupRel.ID
				mockRelRepo.relationships[key] = tt.setupRel
			}

			service := NewRelationshipService(mockRelRepo, mockNodeRepo)

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

func TestRelationshipService_List(t *testing.T) {
	tests := []struct {
		name            string
		tenantID        string
		sourceNodeID    string
		targetNodeID    string
		relType         string
		setupRels       []*repository.Relationship
		expectedCount   int
		expectedError   string
	}{
		{
			name:         "successful list",
			tenantID:     "tenant-1",
			sourceNodeID: "",
			targetNodeID: "",
			relType:      "",
			setupRels: []*repository.Relationship{
				{ID: "rel-1", TenantID: "tenant-1", SourceNodeID: "node-1", TargetNodeID: "node-2"},
				{ID: "rel-2", TenantID: "tenant-1", SourceNodeID: "node-2", TargetNodeID: "node-3"},
			},
			expectedCount: 2,
		},
		{
			name:         "filtered by source node",
			tenantID:     "tenant-1",
			sourceNodeID: "node-1",
			setupRels: []*repository.Relationship{
				{ID: "rel-1", TenantID: "tenant-1", SourceNodeID: "node-1", TargetNodeID: "node-2"},
				{ID: "rel-2", TenantID: "tenant-1", SourceNodeID: "node-2", TargetNodeID: "node-3"},
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
			mockRelRepo := newMockRelationshipRepository()
			mockNodeRepo := newMockNodeRepository()

			for _, rel := range tt.setupRels {
				key := rel.TenantID + ":" + rel.ID
				mockRelRepo.relationships[key] = rel
			}

			service := NewRelationshipService(mockRelRepo, mockNodeRepo)

			ctx := context.Background()
			rels, result, err := service.List(ctx, tt.tenantID, tt.sourceNodeID, tt.targetNodeID, tt.relType, 10, "")

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(rels) != tt.expectedCount {
					t.Errorf("expected %d relationships, got %d", tt.expectedCount, len(rels))
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
			}
		})
	}
}

