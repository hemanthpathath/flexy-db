package integration

import (
	"context"
	"testing"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		user := &repository.User{
			Email:       "john@example.com",
			DisplayName: "John Doe",
		}

		created, err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "john@example.com", created.Email)
		assert.Equal(t, "John Doe", created.DisplayName)
		assert.False(t, created.CreatedAt.IsZero())
	})

	t.Run("duplicate email fails", func(t *testing.T) {
		user1 := &repository.User{
			Email:       "duplicate@example.com",
			DisplayName: "First User",
		}
		_, err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &repository.User{
			Email:       "duplicate@example.com",
			DisplayName: "Second User",
		}
		_, err = repo.Create(ctx, user2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unique")
	})
}

func TestUserRepository_AddToTenant(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	userRepo := repository.NewPostgresUserRepository(pool)
	tenantRepo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	// Create tenant and user
	tenant := &repository.Tenant{
		Slug: "test-tenant",
		Name: "Test Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	user := &repository.User{
		Email:       "user@example.com",
		DisplayName: "Test User",
	}
	createdUser, err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("successful addition", func(t *testing.T) {
		tenantUser := &repository.TenantUser{
			TenantID: createdTenant.ID,
			UserID:   createdUser.ID,
			Role:     "admin",
		}

		result, err := userRepo.AddToTenant(ctx, tenantUser)
		require.NoError(t, err)
		assert.Equal(t, createdTenant.ID, result.TenantID)
		assert.Equal(t, createdUser.ID, result.UserID)
		assert.Equal(t, "admin", result.Role)
	})

	t.Run("duplicate addition fails", func(t *testing.T) {
		tenantUser := &repository.TenantUser{
			TenantID: createdTenant.ID,
			UserID:   createdUser.ID,
			Role:     "member",
		}

		// Try to add same user to same tenant again
		_, err := userRepo.AddToTenant(ctx, tenantUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unique") // Primary key violation
	})
}

func TestUserRepository_ListTenantUsers(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	userRepo := repository.NewPostgresUserRepository(pool)
	tenantRepo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	// Create tenant
	tenant := &repository.Tenant{
		Slug: "multi-user-tenant",
		Name: "Multi User Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	// Create multiple users
	users := []*repository.User{
		{Email: "user1@example.com", DisplayName: "User 1"},
		{Email: "user2@example.com", DisplayName: "User 2"},
		{Email: "user3@example.com", DisplayName: "User 3"},
	}
	createdUsers := make([]*repository.User, len(users))
	for i, user := range users {
		created, err := userRepo.Create(ctx, user)
		require.NoError(t, err)
		createdUsers[i] = created
	}

	// Add all users to tenant
	for _, user := range createdUsers {
		tenantUser := &repository.TenantUser{
			TenantID: createdTenant.ID,
			UserID:   user.ID,
			Role:     "member",
		}
		_, err := userRepo.AddToTenant(ctx, tenantUser)
		require.NoError(t, err)
	}

	t.Run("list all tenant users", func(t *testing.T) {
		tenantUsers, result, err := userRepo.ListTenantUsers(ctx, createdTenant.ID, repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, tenantUsers, 3)
		assert.Equal(t, 3, result.TotalCount)
	})

	t.Run("remove user from tenant", func(t *testing.T) {
		err := userRepo.RemoveFromTenant(ctx, createdTenant.ID, createdUsers[0].ID)
		require.NoError(t, err)

		tenantUsers, _, err := userRepo.ListTenantUsers(ctx, createdTenant.ID, repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, tenantUsers, 2)
	})
}

