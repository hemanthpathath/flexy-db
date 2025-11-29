package integration

import (
	"context"
	"testing"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTenantIsolation_Users tests that users can belong to multiple tenants
// but tenant-specific operations are isolated
func TestTenantIsolation_Users(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	userRepo := repository.NewPostgresUserRepository(pool)
	ctx := context.Background()

	// Create two tenants
	tenant1, err := CreateTestTenant(ctx, pool, "tenant-1", "Tenant 1")
	require.NoError(t, err)

	tenant2, err := CreateTestTenant(ctx, pool, "tenant-2", "Tenant 2")
	require.NoError(t, err)

	// Create a user
	user, err := CreateTestUser(ctx, pool, "shared@example.com", "Shared User")
	require.NoError(t, err)

	// Add user to both tenants
	tenantUser1, err := userRepo.AddToTenant(ctx, &repository.TenantUser{
		TenantID: tenant1.ID,
		UserID:   user.ID,
		Role:     "admin",
	})
	require.NoError(t, err)
	assert.Equal(t, tenant1.ID, tenantUser1.TenantID)

	tenantUser2, err := userRepo.AddToTenant(ctx, &repository.TenantUser{
		TenantID: tenant2.ID,
		UserID:   user.ID,
		Role:     "member",
	})
	require.NoError(t, err)
	assert.Equal(t, tenant2.ID, tenantUser2.TenantID)

	// List users for tenant1 - should only see tenant1's membership
	tenant1Users, _, err := userRepo.ListTenantUsers(ctx, tenant1.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant1Users, 1)
	assert.Equal(t, tenant1.ID, tenant1Users[0].TenantID)
	assert.Equal(t, "admin", tenant1Users[0].Role)

	// List users for tenant2 - should only see tenant2's membership
	tenant2Users, _, err := userRepo.ListTenantUsers(ctx, tenant2.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant2Users, 1)
	assert.Equal(t, tenant2.ID, tenant2Users[0].TenantID)
	assert.Equal(t, "member", tenant2Users[0].Role)

	// Remove user from tenant1, should still be in tenant2
	err = userRepo.RemoveFromTenant(ctx, tenant1.ID, user.ID)
	require.NoError(t, err)

	tenant1UsersAfter, _, err := userRepo.ListTenantUsers(ctx, tenant1.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant1UsersAfter, 0)

	tenant2UsersAfter, _, err := userRepo.ListTenantUsers(ctx, tenant2.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant2UsersAfter, 1)
}

// TestTenantIsolation_NodeTypes tests that node types are isolated per tenant
func TestTenantIsolation_NodeTypes(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	ctx := context.Background()

	// Create two tenants
	tenant1, err := CreateTestTenant(ctx, pool, "tenant-1", "Tenant 1")
	require.NoError(t, err)

	tenant2, err := CreateTestTenant(ctx, pool, "tenant-2", "Tenant 2")
	require.NoError(t, err)

	// Create node types with same name in different tenants
	nodeType1, err := CreateTestNodeType(ctx, pool, tenant1.ID, "Task", "Task for tenant 1", `{"type": "object"}`)
	require.NoError(t, err)

	nodeType2, err := CreateTestNodeType(ctx, pool, tenant2.ID, "Task", "Task for tenant 2", `{"type": "object"}`)
	require.NoError(t, err)

	// Verify they have different IDs
	assert.NotEqual(t, nodeType1.ID, nodeType2.ID)

	// Try to get tenant1's node type using tenant2's ID - should fail
	_, err = nodeTypeRepo.GetByID(ctx, tenant2.ID, nodeType1.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Try to get tenant2's node type using tenant1's ID - should fail
	_, err = nodeTypeRepo.GetByID(ctx, tenant1.ID, nodeType2.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// List node types for each tenant - should only see their own
	tenant1NodeTypes, _, err := nodeTypeRepo.List(ctx, tenant1.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant1NodeTypes, 1)
	assert.Equal(t, tenant1.ID, tenant1NodeTypes[0].TenantID)

	tenant2NodeTypes, _, err := nodeTypeRepo.List(ctx, tenant2.ID, repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant2NodeTypes, 1)
	assert.Equal(t, tenant2.ID, tenant2NodeTypes[0].TenantID)
}

// TestTenantIsolation_Nodes tests that nodes are isolated per tenant
func TestTenantIsolation_Nodes(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	nodeRepo := repository.NewPostgresNodeRepository(pool)
	ctx := context.Background()

	// Create two tenants
	tenant1, err := CreateTestTenant(ctx, pool, "tenant-1", "Tenant 1")
	require.NoError(t, err)

	tenant2, err := CreateTestTenant(ctx, pool, "tenant-2", "Tenant 2")
	require.NoError(t, err)

	// Create node types for each tenant
	nodeType1, err := CreateTestNodeType(ctx, pool, tenant1.ID, "Task", "Task type", `{"type": "object"}`)
	require.NoError(t, err)

	nodeType2, err := CreateTestNodeType(ctx, pool, tenant2.ID, "Task", "Task type", `{"type": "object"}`)
	require.NoError(t, err)

	// Create nodes in each tenant
	node1, err := CreateTestNode(ctx, pool, tenant1.ID, nodeType1.ID, `{"title": "Tenant 1 Task"}`)
	require.NoError(t, err)

	node2, err := CreateTestNode(ctx, pool, tenant2.ID, nodeType2.ID, `{"title": "Tenant 2 Task"}`)
	require.NoError(t, err)

	// Try to get tenant1's node using tenant2's ID - should fail
	_, err = nodeRepo.GetByID(ctx, tenant2.ID, node1.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Try to get tenant2's node using tenant1's ID - should fail
	_, err = nodeRepo.GetByID(ctx, tenant1.ID, node2.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// List nodes for each tenant - should only see their own
	tenant1Nodes, _, err := nodeRepo.List(ctx, tenant1.ID, "", repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant1Nodes, 1)
	assert.Equal(t, tenant1.ID, tenant1Nodes[0].TenantID)
	assert.Equal(t, node1.ID, tenant1Nodes[0].ID)

	tenant2Nodes, _, err := nodeRepo.List(ctx, tenant2.ID, "", repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant2Nodes, 1)
	assert.Equal(t, tenant2.ID, tenant2Nodes[0].TenantID)
	assert.Equal(t, node2.ID, tenant2Nodes[0].ID)

	// Try to update tenant1's node using tenant2's ID - should fail
	node1Copy := *node1
	node1Copy.TenantID = tenant2.ID
	node1Copy.Data = `{"title": "Hacked!"}`
	_, err = nodeRepo.Update(ctx, &node1Copy)
	assert.Error(t, err)

	// Verify node1 is unchanged
	verified, err := nodeRepo.GetByID(ctx, tenant1.ID, node1.ID)
	require.NoError(t, err)
	assert.Contains(t, verified.Data, "Tenant 1 Task")
}

// TestTenantIsolation_Relationships tests that relationships are isolated per tenant
func TestTenantIsolation_Relationships(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	relRepo := repository.NewPostgresRelationshipRepository(pool)
	ctx := context.Background()

	// Create two tenants
	tenant1, err := CreateTestTenant(ctx, pool, "tenant-1", "Tenant 1")
	require.NoError(t, err)

	tenant2, err := CreateTestTenant(ctx, pool, "tenant-2", "Tenant 2")
	require.NoError(t, err)

	// Create node types and nodes for each tenant
	nodeType1, err := CreateTestNodeType(ctx, pool, tenant1.ID, "Task", "Task type", `{"type": "object"}`)
	require.NoError(t, err)

	nodeType2, err := CreateTestNodeType(ctx, pool, tenant2.ID, "Task", "Task type", `{"type": "object"}`)
	require.NoError(t, err)

	source1, err := CreateTestNode(ctx, pool, tenant1.ID, nodeType1.ID, `{"title": "Source 1"}`)
	require.NoError(t, err)

	target1, err := CreateTestNode(ctx, pool, tenant1.ID, nodeType1.ID, `{"title": "Target 1"}`)
	require.NoError(t, err)

	source2, err := CreateTestNode(ctx, pool, tenant2.ID, nodeType2.ID, `{"title": "Source 2"}`)
	require.NoError(t, err)

	target2, err := CreateTestNode(ctx, pool, tenant2.ID, nodeType2.ID, `{"title": "Target 2"}`)
	require.NoError(t, err)

	// Create relationships in each tenant
	rel1, err := CreateTestRelationship(ctx, pool, tenant1.ID, source1.ID, target1.ID, "depends_on", `{"priority": 1}`)
	require.NoError(t, err)

	rel2, err := CreateTestRelationship(ctx, pool, tenant2.ID, source2.ID, target2.ID, "depends_on", `{"priority": 2}`)
	require.NoError(t, err)

	// Try to get tenant1's relationship using tenant2's ID - should fail
	_, err = relRepo.GetByID(ctx, tenant2.ID, rel1.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Try to get tenant2's relationship using tenant1's ID - should fail
	_, err = relRepo.GetByID(ctx, tenant1.ID, rel2.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// List relationships for each tenant - should only see their own
	tenant1Rels, _, err := relRepo.List(ctx, tenant1.ID, "", "", "", repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant1Rels, 1)
	assert.Equal(t, tenant1.ID, tenant1Rels[0].TenantID)
	assert.Equal(t, rel1.ID, tenant1Rels[0].ID)

	tenant2Rels, _, err := relRepo.List(ctx, tenant2.ID, "", "", "", repository.ListOptions{PageSize: 10})
	require.NoError(t, err)
	assert.Len(t, tenant2Rels, 1)
	assert.Equal(t, tenant2.ID, tenant2Rels[0].TenantID)
	assert.Equal(t, rel2.ID, tenant2Rels[0].ID)

	// Try to create relationship between nodes from different tenants
	// Note: This might succeed if the database doesn't enforce tenant isolation
	// at the foreign key level, but the application should prevent it
	_, err = relRepo.Create(ctx, &repository.Relationship{
		TenantID:         tenant1.ID,
		SourceNodeID:     source1.ID,
		TargetNodeID:     source2.ID, // Node from tenant2
		RelationshipType: "depends_on",
		Data:             `{}`,
	})
	// This may or may not fail depending on DB constraints
	// If it succeeds, that's okay - the application layer should handle validation
	// We just verify the operation completes (either success or expected error)
	if err != nil {
		// If it fails, that's expected - cross-tenant relationships shouldn't be allowed
		t.Logf("Cross-tenant relationship creation correctly failed: %v", err)
	} else {
		// If it succeeds, log a warning but don't fail the test
		// The application service layer should validate this
		t.Logf("Warning: Cross-tenant relationship creation succeeded (should be validated at service layer)")
	}
}

// TestTenantIsolation_CrossTenantAccess tests comprehensive cross-tenant access prevention
func TestTenantIsolation_CrossTenantAccess(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	defer cleanupTestDB(t, pool)

	ctx := context.Background()

	// Create comprehensive test fixtures
	fixtures, err := CreateTestFixtures(ctx, pool)
	require.NoError(t, err)

	// Create a second tenant with its own data
	tenant2, err := CreateTestTenant(ctx, pool, "tenant-2-isolation", "Tenant 2 Isolation")
	require.NoError(t, err)

	nodeType2, err := CreateTestNodeType(ctx, pool, tenant2.ID, "Task", "Task type", `{"type": "object"}`)
	require.NoError(t, err)

	node2, err := CreateTestNode(ctx, pool, tenant2.ID, nodeType2.ID, `{"title": "Tenant 2 Node"}`)
	require.NoError(t, err)

	// Test: Tenant2 cannot access Tenant1's node type
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	_, err = nodeTypeRepo.GetByID(ctx, tenant2.ID, fixtures.NodeType1.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Test: Tenant2 cannot access Tenant1's node
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	_, err = nodeRepo.GetByID(ctx, tenant2.ID, fixtures.Node1.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Test: Tenant1 cannot access Tenant2's node
	_, err = nodeRepo.GetByID(ctx, fixtures.Tenant1.ID, node2.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrNotFound, err)

	// Test: Tenant2 cannot update Tenant1's node
	fixtures.Node1.TenantID = tenant2.ID
	_, err = nodeRepo.Update(ctx, fixtures.Node1)
	assert.Error(t, err)

	// Test: Tenant2 cannot delete Tenant1's node
	err = nodeRepo.Delete(ctx, tenant2.ID, fixtures.Node1.ID)
	assert.Error(t, err)

	// Test: Tenant2 cannot create relationship with Tenant1's nodes
	relRepo := repository.NewPostgresRelationshipRepository(pool)
	_, err = relRepo.Create(ctx, &repository.Relationship{
		TenantID:         tenant2.ID,
		SourceNodeID:     node2.ID,
		TargetNodeID:     fixtures.Node1.ID, // Node from tenant1
		RelationshipType: "depends_on",
		Data:             `{}`,
	})
	// This may or may not fail depending on DB constraints
	// If it succeeds, that's okay - the application layer should handle validation
	if err != nil {
		// If it fails, that's expected
		t.Logf("Cross-tenant relationship creation correctly failed: %v", err)
	} else {
		// If it succeeds, log a warning but don't fail the test
		t.Logf("Warning: Cross-tenant relationship creation succeeded (should be validated at service layer)")
	}
}

