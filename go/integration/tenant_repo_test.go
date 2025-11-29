package integration

import (
	"context"
	"testing"
	"time"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		tenant := &repository.Tenant{
			Slug: "acme-corp",
			Name: "Acme Corporation",
		}

		created, err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "acme-corp", created.Slug)
		assert.Equal(t, "Acme Corporation", created.Name)
		assert.Equal(t, "active", created.Status)
		assert.False(t, created.CreatedAt.IsZero())
		assert.False(t, created.UpdatedAt.IsZero())
	})

	t.Run("duplicate slug fails", func(t *testing.T) {
		tenant1 := &repository.Tenant{
			Slug: "duplicate-slug",
			Name: "First Tenant",
		}
		_, err := repo.Create(ctx, tenant1)
		require.NoError(t, err)

		tenant2 := &repository.Tenant{
			Slug: "duplicate-slug",
			Name: "Second Tenant",
		}
		_, err = repo.Create(ctx, tenant2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unique") // Unique constraint violation
	})

	t.Run("custom status", func(t *testing.T) {
		tenant := &repository.Tenant{
			Slug:   "inactive-tenant",
			Name:   "Inactive Tenant",
			Status: "inactive",
		}

		created, err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		assert.Equal(t, "inactive", created.Status)
	})
}

func TestTenantRepository_GetByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	// Create a tenant first
	tenant := &repository.Tenant{
		Slug: "test-tenant",
		Name: "Test Tenant",
	}
	created, err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		found, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Slug, found.Slug)
		assert.Equal(t, created.Name, found.Name)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestTenantRepository_Update(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	// Create a tenant
	tenant := &repository.Tenant{
		Slug: "original-slug",
		Name: "Original Name",
	}
	created, err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		created.Slug = "updated-slug"
		created.Name = "Updated Name"
		created.Status = "inactive"

		updated, err := repo.Update(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, "updated-slug", updated.Slug)
		assert.Equal(t, "Updated Name", updated.Name)
		assert.Equal(t, "inactive", updated.Status)
		assert.True(t, updated.UpdatedAt.After(created.CreatedAt))
	})

	t.Run("not found", func(t *testing.T) {
		nonExistent := &repository.Tenant{
			ID:   "non-existent-id",
			Slug: "test",
			Name: "Test",
		}
		_, err := repo.Update(ctx, nonExistent)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestTenantRepository_Delete(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		tenant := &repository.Tenant{
			Slug: "to-delete",
			Name: "To Delete",
		}
		created, err := repo.Create(ctx, tenant)
		require.NoError(t, err)

		err = repo.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = repo.GetByID(ctx, created.ID)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})

	t.Run("delete non-existent", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestTenantRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	repo := repository.NewPostgresTenantRepository(pool)
	ctx := context.Background()

	// Create multiple tenants
	tenantNames := []string{"tenant-a", "tenant-b", "tenant-c", "tenant-d", "tenant-e"}
	for _, name := range tenantNames {
		tenant := &repository.Tenant{
			Slug: name,
			Name: "Tenant " + name,
		}
		_, err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		// Add small delay to ensure different timestamps for ordering
		time.Sleep(time.Millisecond * 10)
	}

	t.Run("list all with pagination", func(t *testing.T) {
		tenants, result, err := repo.List(ctx, repository.ListOptions{
			PageSize:  3,
			PageToken: "",
		})
		require.NoError(t, err)
		assert.Len(t, tenants, 3)
		assert.Equal(t, 5, result.TotalCount)
		assert.NotEmpty(t, result.NextPageToken)

		// Get next page
		tenants2, result2, err := repo.List(ctx, repository.ListOptions{
			PageSize:  3,
			PageToken: result.NextPageToken,
		})
		require.NoError(t, err)
		assert.Len(t, tenants2, 2)
		assert.Equal(t, 5, result2.TotalCount)
	})

	t.Run("list with page size limit", func(t *testing.T) {
		tenants, result, err := repo.List(ctx, repository.ListOptions{
			PageSize: 100, // Should be capped at 100
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(tenants), 100)
		assert.Equal(t, 5, result.TotalCount)
	})

	t.Run("empty list", func(t *testing.T) {
		cleanupTestDB(t, pool)
		tenants, result, err := repo.List(ctx, repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Empty(t, tenants)
		assert.Equal(t, 0, result.TotalCount)
	})
}

