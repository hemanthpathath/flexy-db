package service

import (
	"context"
	"fmt"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, email, displayName string) (*repository.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if displayName == "" {
		return nil, fmt.Errorf("display_name is required")
	}

	user := &repository.User{
		Email:       email,
		DisplayName: displayName,
	}

	return s.repo.Create(ctx, user)
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*repository.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}
	return s.repo.GetByID(ctx, id)
}

// Update updates an existing user
func (s *UserService) Update(ctx context.Context, id, email, displayName string) (*repository.User, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if email != "" {
		user.Email = email
	}
	if displayName != "" {
		user.DisplayName = displayName
	}

	return s.repo.Update(ctx, user)
}

// Delete deletes a user
func (s *UserService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.repo.Delete(ctx, id)
}

// List retrieves users with pagination
func (s *UserService) List(ctx context.Context, pageSize int32, pageToken string) ([]*repository.User, *repository.ListResult, error) {
	opts := repository.ListOptions{
		PageSize:  int(pageSize),
		PageToken: pageToken,
	}
	return s.repo.List(ctx, opts)
}

// AddToTenant adds a user to a tenant
func (s *UserService) AddToTenant(ctx context.Context, tenantID, userID, role string) (*repository.TenantUser, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	tenantUser := &repository.TenantUser{
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
	}

	return s.repo.AddToTenant(ctx, tenantUser)
}

// RemoveFromTenant removes a user from a tenant
func (s *UserService) RemoveFromTenant(ctx context.Context, tenantID, userID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	return s.repo.RemoveFromTenant(ctx, tenantID, userID)
}

// ListTenantUsers lists users in a tenant
func (s *UserService) ListTenantUsers(ctx context.Context, tenantID string, pageSize int32, pageToken string) ([]*repository.TenantUser, *repository.ListResult, error) {
	if tenantID == "" {
		return nil, nil, fmt.Errorf("tenant_id is required")
	}

	opts := repository.ListOptions{
		PageSize:  int(pageSize),
		PageToken: pageToken,
	}
	return s.repo.ListTenantUsers(ctx, tenantID, opts)
}
