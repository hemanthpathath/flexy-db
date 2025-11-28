package integration

import (
	"context"
	"testing"

	"github.com/hemanthpathath/flex-db/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelationshipRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	tenantRepo := repository.NewPostgresTenantRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	relRepo := repository.NewPostgresRelationshipRepository(pool)
	ctx := context.Background()

	// Setup: Create tenant, node type, and nodes
	tenant := &repository.Tenant{
		Slug: "rel-tenant",
		Name: "Relationship Tenant",
	}
	createdTenant, err := tenantRepo.Create(ctx, tenant)
	require.NoError(t, err)

	nodeType := &repository.NodeType{
		TenantID: createdTenant.ID,
		Name:     "Task",
	}
	createdNodeType, err := nodeTypeRepo.Create(ctx, nodeType)
	require.NoError(t, err)

	sourceNode := &repository.Node{
		TenantID:   createdTenant.ID,
		NodeTypeID: createdNodeType.ID,
		Data:       `{"title": "Source Task"}`,
	}
	createdSource, err := nodeRepo.Create(ctx, sourceNode)
	require.NoError(t, err)

	targetNode := &repository.Node{
		TenantID:   createdTenant.ID,
		NodeTypeID: createdNodeType.ID,
		Data:       `{"title": "Target Task"}`,
	}
	createdTarget, err := nodeRepo.Create(ctx, targetNode)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		rel := &repository.Relationship{
			TenantID:         createdTenant.ID,
			SourceNodeID:     createdSource.ID,
			TargetNodeID:     createdTarget.ID,
			RelationshipType: "depends_on",
			Data:             `{"priority": 1}`,
		}

		created, err := relRepo.Create(ctx, rel)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, createdSource.ID, created.SourceNodeID)
		assert.Equal(t, createdTarget.ID, created.TargetNodeID)
		assert.Equal(t, "depends_on", created.RelationshipType)
		assert.Contains(t, created.Data, "priority")
	})

	t.Run("list relationships with filters", func(t *testing.T) {
		// Create another relationship
		rel2 := &repository.Relationship{
			TenantID:         createdTenant.ID,
			SourceNodeID:     createdTarget.ID,
			TargetNodeID:     createdSource.ID,
			RelationshipType: "related_to",
			Data:             `{"type": "reverse"}`,
		}
		_, err := relRepo.Create(ctx, rel2)
		require.NoError(t, err)

		// Filter by source node
		rels, _, err := relRepo.List(ctx, createdTenant.ID, createdSource.ID, "", "", repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rels), 1)
		assert.Equal(t, createdSource.ID, rels[0].SourceNodeID)

		// Filter by relationship type
		rels2, _, err := relRepo.List(ctx, createdTenant.ID, "", "", "depends_on", repository.ListOptions{
			PageSize: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rels2), 1)
		assert.Equal(t, "depends_on", rels2[0].RelationshipType)
	})
}

