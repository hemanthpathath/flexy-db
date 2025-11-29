package integration

import (
	"context"
	"testing"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	tenantRepo := repository.NewPostgresTenantRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	ctx := context.Background()

	// Setup: Create tenant and node type
	tenant := &repository.Tenant{
		Slug: "test-tenant",
		Name: "Test Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	nodeType := &repository.NodeType{
		TenantID:    createdTenant.ID,
		Name:        "Task",
		Description: "A task node",
		Schema:      `{"type": "object"}`,
	}
	createdNodeType, err := nodeTypeRepo.Create(ctx, nodeType)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		node := &repository.Node{
			TenantID:   createdTenant.ID,
			NodeTypeID: createdNodeType.ID,
			Data:       `{"title": "Complete project", "priority": "high"}`,
		}

		created, err := nodeRepo.Create(ctx, node)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, createdTenant.ID, created.TenantID)
		assert.Equal(t, createdNodeType.ID, created.NodeTypeID)
		assert.Contains(t, created.Data, "Complete project")
	})

	t.Run("create with JSONB data", func(t *testing.T) {
		complexData := `{"title": "Task 2", "assignee": {"name": "John", "id": "123"}, "tags": ["urgent", "backend"]}`
		node := &repository.Node{
			TenantID:   createdTenant.ID,
			NodeTypeID: createdNodeType.ID,
			Data:       complexData,
		}

		created, err := nodeRepo.Create(ctx, node)
		require.NoError(t, err)
		assert.Contains(t, created.Data, "Task 2")
		assert.Contains(t, created.Data, "assignee")
	})

	t.Run("tenant isolation", func(t *testing.T) {
		// Create another tenant
		tenant2 := &repository.Tenant{
			Slug: "tenant-2",
			Name: "Tenant 2",
		}
		createdTenant2, err := tenantRepo.Create(ctx, tenant2)
		require.NoError(t, err)

		// Try to get node from tenant 1 using tenant 2's ID (should fail)
		node := &repository.Node{
			TenantID:   createdTenant.ID,
			NodeTypeID: createdNodeType.ID,
			Data:       `{"title": "Isolated node"}`,
		}
		createdNode, err := nodeRepo.Create(ctx, node)
		require.NoError(t, err)

		// Try to access with wrong tenant ID
		_, err = nodeRepo.GetByID(ctx, createdTenant2.ID, createdNode.ID)
		assert.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestNodeRepository_List(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	tenantRepo := repository.NewPostgresTenantRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	ctx := context.Background()

	// Setup
	tenant := &repository.Tenant{
		Slug: "filter-tenant",
		Name: "Filter Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	nodeType1 := &repository.NodeType{
		TenantID: createdTenant.ID,
		Name:     "Task",
	}
	createdNodeType1, err := nodeTypeRepo.Create(ctx, nodeType1)
	require.NoError(t, err)

	nodeType2 := &repository.NodeType{
		TenantID: createdTenant.ID,
		Name:     "Note",
	}
	createdNodeType2, err := nodeTypeRepo.Create(ctx, nodeType2)
	require.NoError(t, err)

	// Create nodes of different types
	for i := 0; i < 3; i++ {
		node := &repository.Node{
			TenantID:   createdTenant.ID,
			NodeTypeID: createdNodeType1.ID,
			Data:       `{"title": "Task ` + string(rune('A'+i)) + `"}`,
		}
		_, err := nodeRepo.Create(ctx, node)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		node := &repository.Node{
			TenantID:   createdTenant.ID,
			NodeTypeID: createdNodeType2.ID,
			Data:       `{"title": "Note ` + string(rune('1'+i)) + `"}`,
		}
		_, err := nodeRepo.Create(ctx, node)
		require.NoError(t, err)
	}

	t.Run("list all nodes", func(t *testing.T) {
		nodes, result, err := nodeRepo.List(ctx, createdTenant.ID, "", repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, nodes, 5)
		assert.Equal(t, 5, result.TotalCount)
	})

	t.Run("filter by node type", func(t *testing.T) {
		nodes, result, err := nodeRepo.List(ctx, createdTenant.ID, createdNodeType1.ID, repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.Len(t, nodes, 3)
		assert.Equal(t, 3, result.TotalCount)
	})
}

