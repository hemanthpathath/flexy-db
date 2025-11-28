package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/internal/repository"
)

// mockUserRepository is a mock implementation of UserRepository
type mockUserRepository struct {
	users        map[string]*repository.User
	tenantUsers  map[string]*repository.TenantUser // key: tenantID:userID
	err          error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:       make(map[string]*repository.User),
		tenantUsers: make(map[string]*repository.TenantUser),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *repository.User) (*repository.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	user.ID = "user-" + user.Email
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *repository.User) (*repository.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	existing, ok := m.users[user.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	user.UpdatedAt = time.Now()
	user.CreatedAt = existing.CreatedAt
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.users[id]; !ok {
		return errors.New("not found")
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, opts repository.ListOptions) ([]*repository.User, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var users []*repository.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, &repository.ListResult{TotalCount: len(users)}, nil
}

func (m *mockUserRepository) AddToTenant(ctx context.Context, tenantUser *repository.TenantUser) (*repository.TenantUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := tenantUser.TenantID + ":" + tenantUser.UserID
	m.tenantUsers[key] = tenantUser
	return tenantUser, nil
}

func (m *mockUserRepository) RemoveFromTenant(ctx context.Context, tenantID, userID string) error {
	if m.err != nil {
		return m.err
	}
	key := tenantID + ":" + userID
	if _, ok := m.tenantUsers[key]; !ok {
		return errors.New("not found")
	}
	delete(m.tenantUsers, key)
	return nil
}

func (m *mockUserRepository) ListTenantUsers(ctx context.Context, tenantID string, opts repository.ListOptions) ([]*repository.TenantUser, *repository.ListResult, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	var tenantUsers []*repository.TenantUser
	for _, tu := range m.tenantUsers {
		if tu.TenantID == tenantID {
			tenantUsers = append(tenantUsers, tu)
		}
	}
	return tenantUsers, &repository.ListResult{TotalCount: len(tenantUsers)}, nil
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		displayName   string
		repoErr       error
		expectedError string
	}{
		{
			name:        "successful creation",
			email:       "john@example.com",
			displayName: "John Doe",
		},
		{
			name:          "empty email",
			email:         "",
			displayName:   "John Doe",
			expectedError: "email is required",
		},
		{
			name:          "empty display name",
			email:         "john@example.com",
			displayName:   "",
			expectedError: "display_name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			mockRepo.err = tt.repoErr
			service := NewUserService(mockRepo)

			ctx := context.Background()
			user, err := service.Create(ctx, tt.email, tt.displayName)

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
				if user == nil {
					t.Fatal("expected user, got nil")
				}
				if user.Email != tt.email {
					t.Errorf("expected email %q, got %q", tt.email, user.Email)
				}
				if user.DisplayName != tt.displayName {
					t.Errorf("expected display name %q, got %q", tt.displayName, user.DisplayName)
				}
			}
		})
	}
}

func TestUserService_AddToTenant(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		userID        string
		role          string
		expectedError string
	}{
		{
			name:     "successful addition",
			tenantID: "tenant-1",
			userID:   "user-1",
			role:     "admin",
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			userID:        "user-1",
			expectedError: "tenant_id is required",
		},
		{
			name:          "empty user id",
			tenantID:      "tenant-1",
			userID:        "",
			expectedError: "user_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			service := NewUserService(mockRepo)

			ctx := context.Background()
			tenantUser, err := service.AddToTenant(ctx, tt.tenantID, tt.userID, tt.role)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tenantUser == nil {
					t.Fatal("expected tenant user, got nil")
				}
				if tenantUser.TenantID != tt.tenantID {
					t.Errorf("expected tenant ID %q, got %q", tt.tenantID, tenantUser.TenantID)
				}
				if tenantUser.UserID != tt.userID {
					t.Errorf("expected user ID %q, got %q", tt.userID, tenantUser.UserID)
				}
				if tenantUser.Role != tt.role {
					t.Errorf("expected role %q, got %q", tt.role, tenantUser.Role)
				}
			}
		})
	}
}

func TestUserService_RemoveFromTenant(t *testing.T) {
	tests := []struct {
		name          string
		tenantID      string
		userID        string
		setupTenantUser *repository.TenantUser
		expectedError string
	}{
		{
			name:     "successful removal",
			tenantID: "tenant-1",
			userID:   "user-1",
			setupTenantUser: &repository.TenantUser{
				TenantID: "tenant-1",
				UserID:   "user-1",
			},
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			userID:        "user-1",
			expectedError: "tenant_id is required",
		},
		{
			name:          "empty user id",
			tenantID:      "tenant-1",
			userID:        "",
			expectedError: "user_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			if tt.setupTenantUser != nil {
				key := tt.setupTenantUser.TenantID + ":" + tt.setupTenantUser.UserID
				mockRepo.tenantUsers[key] = tt.setupTenantUser
			}
			service := NewUserService(mockRepo)

			ctx := context.Background()
			err := service.RemoveFromTenant(ctx, tt.tenantID, tt.userID)

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

func TestUserService_ListTenantUsers(t *testing.T) {
	tests := []struct {
		name           string
		tenantID       string
		setupTenantUsers []*repository.TenantUser
		expectedCount  int
		expectedError  string
	}{
		{
			name:     "successful list",
			tenantID: "tenant-1",
			setupTenantUsers: []*repository.TenantUser{
				{TenantID: "tenant-1", UserID: "user-1"},
				{TenantID: "tenant-1", UserID: "user-2"},
			},
			expectedCount: 2,
		},
		{
			name:          "empty tenant id",
			tenantID:      "",
			expectedError: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := newMockUserRepository()
			for _, tu := range tt.setupTenantUsers {
				key := tu.TenantID + ":" + tu.UserID
				mockRepo.tenantUsers[key] = tu
			}
			service := NewUserService(mockRepo)

			ctx := context.Background()
			tenantUsers, result, err := service.ListTenantUsers(ctx, tt.tenantID, 10, "")

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(tenantUsers) != tt.expectedCount {
					t.Errorf("expected %d tenant users, got %d", tt.expectedCount, len(tenantUsers))
				}
				if result == nil {
					t.Error("expected non-nil result")
				}
			}
		})
	}
}

