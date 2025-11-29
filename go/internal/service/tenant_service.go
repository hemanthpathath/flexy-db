package service

import (
	"context"
	"fmt"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// TenantService handles tenant business logic
type TenantService struct {
	repo repository.TenantRepository
}

// NewTenantService creates a new TenantService
func NewTenantService(repo repository.TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

// Create creates a new tenant
func (s *TenantService) Create(ctx context.Context, slug, name string) (*repository.Tenant, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	tenant := &repository.Tenant{
		Slug: slug,
		Name: name,
	}

	return s.repo.Create(ctx, tenant)
}

// GetByID retrieves a tenant by ID
func (s *TenantService) GetByID(ctx context.Context, id string) (*repository.Tenant, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// Update updates an existing tenant
func (s *TenantService) Update(ctx context.Context, id, slug, name, status string) (*repository.Tenant, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if slug != "" {
		tenant.Slug = slug
	}
	if name != "" {
		tenant.Name = name
	}
	if status != "" {
		tenant.Status = status
	}

	return s.repo.Update(ctx, tenant)
}

// Delete deletes a tenant
func (s *TenantService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.Delete(ctx, id)
}

// List retrieves tenants with pagination
func (s *TenantService) List(ctx context.Context, pageSize int32, pageToken string) ([]*repository.Tenant, *repository.ListResult, error) {
	opts := repository.ListOptions{
		PageSize:  int(pageSize),
		PageToken: pageToken,
	}
	return s.repo.List(ctx, opts)
}
