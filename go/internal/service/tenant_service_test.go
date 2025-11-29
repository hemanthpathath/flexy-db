package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// mockTenantRepository is a mock implementation of TenantRepository
type mockTenantRepository struct {
	tenants map[string]*repository.Tenant
	err     error
}

func newMockTenantRepository() *mockTenantRepository {
	return &mockTenantRepository{
		tenants: make(map[string]*repository.Tenant),
	}
}

func (m *mockTenantRepository) Create(ctx context.Context, tenant *repository.Tenant) (*repository.Tenant, error) {
	if m.err != nil {
		return nil, m.err
	}
	tenant.ID = "tenant-" + tenant.Slug
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	return tenant, nil
}

func (m *mockTenantRepository) GetByID(ctx context.Context, id string) (*repository.Tenant, error) {
	if m.err != nil {
		return nil, m.err
	}
	tenant, ok := m.tenants[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return tenant, nil
}

func (m *mockTenantRepository) Update(ctx context.Context, tenant *repository.Tenant) (*repository.Tenant, error) {
	if m.err != nil {
		return nil, m.err
	}
	existing, ok := m.tenants[tenant.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	tenant.UpdatedAt = time.Now()
	tenant.CreatedAt = existing.CreatedAt
	m.tenants[tenant.ID] = tenant
	return tenant, nil
}

func (m *mockTenantRepository) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.tenants[id]; !ok {
		return errors.New("not found")
	}
	delete(m.tenants, id)
	return nil
}

func (m *mockTenantRepository) List(ctx context.Context, opts repository.ListOptions) ([]*repository.Tenant, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var tenants []*repository.Tenant
	for _, tenant := range m.tenants {
		tenants = append(tenants, tenant)
	}
	return tenants, &repository.ListResult{TotalCount: len(tenants)}, nil
}

func TestTenantService_Create(t *testing.T) {
	tests := []struct {
		name          string
		slug          string
		tenantName    string
		repoErr       error
		expectedError string
	}{
		{
			name:       "successful creation",
			slug:       "acme-corp",
			tenantName: "Acme Corporation",
		},
		{
			name:          "empty slug",
			slug:          "",
			tenantName:    "Acme Corporation",
			expectedError: "slug is required",
		},
		{
			name:          "empty name",
			slug:          "acme-corp",
			tenantName:    "",
			expectedError: "name is required",
		},
		{
			name:       "repository error",
			slug:       "acme-corp",
			tenantName: "Acme Corporation",
			repoErr:    errors.New("database error"),
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockTenantRepository()
			mockRepo.err = tt.repoErr
			service := NewTenantService(mockRepo)

			ctx := context.Background()
			tenant, err := service.Create(ctx, tt.slug, tt.tenantName)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				if tenant != nil {
					t.Errorf("expected nil tenant, got %+v", tenant)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tenant.Slug != tt.slug {
					t.Errorf("expected slug %q, got %q", tt.slug, tenant.Slug)
				}
				if tenant.Name != tt.tenantName {
					t.Errorf("expected name %q, got %q", tt.tenantName, tenant.Name)
				}
				if tenant.ID == "" {
					t.Error("expected non-empty ID")
				}
			}
		})
	}
}

func TestTenantService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		setupTenant   *repository.Tenant
		repoErr       error
		expectedError string
	}{
		{
			name: "successful retrieval",
			id:   "tenant-1",
			setupTenant: &repository.Tenant{
				ID:   "tenant-1",
				Slug: "acme",
				Name: "Acme Corp",
			},
		},
		{
			name:          "empty id",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "not found",
			id:            "non-existent",
			expectedError: "not found",
		},
		{
			name:          "repository error",
			id:            "tenant-1",
			repoErr:       errors.New("database error"),
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockTenantRepository()
			mockRepo.err = tt.repoErr
			if tt.setupTenant != nil {
				mockRepo.tenants[tt.setupTenant.ID] = tt.setupTenant
			}
			service := NewTenantService(mockRepo)

			ctx := context.Background()
			tenant, err := service.GetByID(ctx, tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				if tenant != nil {
					t.Errorf("expected nil tenant, got %+v", tenant)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tenant.ID != tt.id {
					t.Errorf("expected ID %q, got %q", tt.id, tenant.ID)
				}
			}
		})
	}
}

func TestTenantService_Update(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		slug          string
		tenantName    string
		status        string
		setupTenant   *repository.Tenant
		repoErr       error
		expectedError string
		expectChanges bool
	}{
		{
			name:       "successful update all fields",
			id:         "tenant-1",
			slug:       "new-slug",
			tenantName: "New Name",
			status:     "active",
			setupTenant: &repository.Tenant{
				ID:     "tenant-1",
				Slug:   "old-slug",
				Name:   "Old Name",
				Status: "inactive",
			},
			expectChanges: true,
		},
		{
			name:       "partial update",
			id:         "tenant-1",
			slug:       "new-slug",
			tenantName: "",
			status:     "",
			setupTenant: &repository.Tenant{
				ID:   "tenant-1",
				Slug: "old-slug",
				Name: "Old Name",
			},
			expectChanges: true,
		},
		{
			name:          "empty id",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "tenant not found",
			id:            "non-existent",
			setupTenant:   nil,
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockTenantRepository()
			mockRepo.err = tt.repoErr
			if tt.setupTenant != nil {
				mockRepo.tenants[tt.setupTenant.ID] = tt.setupTenant
			}
			service := NewTenantService(mockRepo)

			ctx := context.Background()
			tenant, err := service.Update(ctx, tt.id, tt.slug, tt.tenantName, tt.status)

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
				if tenant == nil {
					t.Fatal("expected tenant, got nil")
				}
				if tt.expectChanges {
					if tt.slug != "" && tenant.Slug != tt.slug {
						t.Errorf("expected slug %q, got %q", tt.slug, tenant.Slug)
					}
					if tt.tenantName != "" && tenant.Name != tt.tenantName {
						t.Errorf("expected name %q, got %q", tt.tenantName, tenant.Name)
					}
					if tt.status != "" && tenant.Status != tt.status {
						t.Errorf("expected status %q, got %q", tt.status, tenant.Status)
					}
				}
			}
		})
	}
}

func TestTenantService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		setupTenant   *repository.Tenant
		repoErr       error
		expectedError string
	}{
		{
			name: "successful deletion",
			id:   "tenant-1",
			setupTenant: &repository.Tenant{
				ID:   "tenant-1",
				Slug: "acme",
				Name: "Acme Corp",
			},
		},
		{
			name:          "empty id",
			id:            "",
			expectedError: "id is required",
		},
		{
			name:          "not found",
			id:            "non-existent",
			expectedError: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockTenantRepository()
			mockRepo.err = tt.repoErr
			if tt.setupTenant != nil {
				mockRepo.tenants[tt.setupTenant.ID] = tt.setupTenant
			}
			service := NewTenantService(mockRepo)

			ctx := context.Background()
			err := service.Delete(ctx, tt.id)

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
				// Verify tenant was deleted
				if _, exists := mockRepo.tenants[tt.id]; exists {
					t.Error("tenant should have been deleted")
				}
			}
		})
	}
}

func TestTenantService_List(t *testing.T) {
	tests := []struct {
		name           string
		pageSize       int32
		pageToken      string
		setupTenants   []*repository.Tenant
		repoErr        error
		expectedCount  int
		expectedError  string
	}{
		{
			name:          "successful list",
			pageSize:      10,
			setupTenants:  []*repository.Tenant{{ID: "t1"}, {ID: "t2"}},
			expectedCount: 2,
		},
		{
			name:          "empty list",
			pageSize:      10,
			setupTenants:  []*repository.Tenant{},
			expectedCount: 0,
		},
		{
			name:         "repository error",
			pageSize:     10,
			repoErr:      errors.New("database error"),
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockTenantRepository()
			mockRepo.err = tt.repoErr
			for _, tenant := range tt.setupTenants {
				mockRepo.tenants[tenant.ID] = tenant
			}
			service := NewTenantService(mockRepo)

			ctx := context.Background()
			tenants, result, err := service.List(ctx, tt.pageSize, tt.pageToken)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(tenants) != tt.expectedCount {
					t.Errorf("expected %d tenants, got %d", tt.expectedCount, len(tenants))
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
			}
		})
	}
}

