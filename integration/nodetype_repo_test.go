package integration

import (
	"context"
	"testing"

	"github.com/hemanthpathath/flex-db/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeTypeRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	tenantRepo := repository.NewPostgresTenantRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	ctx := context.Background()

	tenant := &repository.Tenant{
		Slug: "nodetype-tenant",
		Name: "NodeType Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		nodeType := &repository.NodeType{
			TenantID:    createdTenant.ID,
			Name:        "Task",
			Description: "A task node type",
			Schema:      `{"type": "object", "properties": {"title": {"type": "string"}}}`,
		}

		created, err := nodeTypeRepo.Create(ctx, nodeType)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, createdTenant.ID, created.TenantID)
		assert.Equal(t, "Task", created.Name)
		assert.Contains(t, created.Schema, "properties")
	})

	t.Run("tenant isolation", func(t *testing.T) {
		// Create node type for tenant 1
		nodeType1 := &repository.NodeType{
			TenantID: createdTenant.ID,
			Name:     "Isolated Type",
		}
		created1, err := nodeTypeRepo.Create(ctx, nodeType1)
		require.NoError(t, err)

		// Create another tenant
		tenant2 := &repository.Tenant{
			Slug: "tenant-2",
			Name: "Tenant 2",
		}
		createdTenant2, err := tenantRepo.Create(ctx, tenant2)
		require.NoError(t, err)

		// Try to get node type from tenant 1 using tenant 2's ID (should fail)
		_, err = nodeTypeRepo.GetByID(ctx, createdTenant2.ID, created1.ID)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

