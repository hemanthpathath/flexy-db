package integration

import (
	"context"
	"fmt"

	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestFixtures holds commonly used test data
type TestFixtures struct {
	Tenant1    *repository.Tenant
	Tenant2    *repository.Tenant
	User1      *repository.User
	User2      *repository.User
	NodeType1  *repository.NodeType
	NodeType2  *repository.NodeType
	Node1      *repository.Node
	Node2      *repository.Node
	Relationship1 *repository.Relationship
}

// CreateTestFixtures creates a complete set of test fixtures for integration tests
func CreateTestFixtures(ctx context.Context, pool *pgxpool.Pool) (*TestFixtures, error) {
	tenantRepo := repository.NewPostgresTenantRepository(pool)
	userRepo := repository.NewPostgresUserRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	relRepo := repository.NewPostgresRelationshipRepository(pool)

	fixtures := &TestFixtures{}

	// Create tenants
	tenant1, err := tenantRepo.Create(ctx, &repository.Tenant{
		Slug: "test-tenant-1",
		Name: "Test Tenant 1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant1: %w", err)
	}
	fixtures.Tenant1 = tenant1

	tenant2, err := tenantRepo.Create(ctx, &repository.Tenant{
		Slug: "test-tenant-2",
		Name: "Test Tenant 2",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant2: %w", err)
	}
	fixtures.Tenant2 = tenant2

	// Create users
	user1, err := userRepo.Create(ctx, &repository.User{
		Email:       "user1@example.com",
		DisplayName: "User One",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user1: %w", err)
	}
	fixtures.User1 = user1

	user2, err := userRepo.Create(ctx, &repository.User{
		Email:       "user2@example.com",
		DisplayName: "User Two",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user2: %w", err)
	}
	fixtures.User2 = user2

	// Create node types
	nodeType1, err := nodeTypeRepo.Create(ctx, &repository.NodeType{
		TenantID:    tenant1.ID,
		Name:        "Task",
		Description: "A task node type",
		Schema:      `{"type": "object", "properties": {"title": {"type": "string"}}}`,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nodeType1: %w", err)
	}
	fixtures.NodeType1 = nodeType1

	nodeType2, err := nodeTypeRepo.Create(ctx, &repository.NodeType{
		TenantID:    tenant1.ID,
		Name:        "Note",
		Description: "A note node type",
		Schema:      `{"type": "object"}`,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nodeType2: %w", err)
	}
	fixtures.NodeType2 = nodeType2

	// Create nodes
	node1, err := nodeRepo.Create(ctx, &repository.Node{
		TenantID:   tenant1.ID,
		NodeTypeID: nodeType1.ID,
		Data:       `{"title": "Test Task 1", "priority": "high"}`,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create node1: %w", err)
	}
	fixtures.Node1 = node1

	node2, err := nodeRepo.Create(ctx, &repository.Node{
		TenantID:   tenant1.ID,
		NodeTypeID: nodeType1.ID,
		Data:       `{"title": "Test Task 2", "priority": "low"}`,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create node2: %w", err)
	}
	fixtures.Node2 = node2

	// Create relationship
	rel1, err := relRepo.Create(ctx, &repository.Relationship{
		TenantID:         tenant1.ID,
		SourceNodeID:     node1.ID,
		TargetNodeID:     node2.ID,
		RelationshipType: "depends_on",
		Data:             `{"priority": 1}`,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create relationship1: %w", err)
	}
	fixtures.Relationship1 = rel1

	return fixtures, nil
}

// CreateTestTenant creates a test tenant
func CreateTestTenant(ctx context.Context, pool *pgxpool.Pool, slug, name string) (*repository.Tenant, error) {
	repo := repository.NewPostgresTenantRepository(pool)
	return repo.Create(ctx, &repository.Tenant{
		Slug: slug,
		Name: name,
	})
}

// CreateTestUser creates a test user
func CreateTestUser(ctx context.Context, pool *pgxpool.Pool, email, displayName string) (*repository.User, error) {
	repo := repository.NewPostgresUserRepository(pool)
	return repo.Create(ctx, &repository.User{
		Email:       email,
		DisplayName: displayName,
	})
}

// CreateTestNodeType creates a test node type
func CreateTestNodeType(ctx context.Context, pool *pgxpool.Pool, tenantID, name, description, schema string) (*repository.NodeType, error) {
	repo := repository.NewPostgresNodeTypeRepository(pool)
	return repo.Create(ctx, &repository.NodeType{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		Schema:      schema,
	})
}

// CreateTestNode creates a test node
func CreateTestNode(ctx context.Context, pool *pgxpool.Pool, tenantID, nodeTypeID, data string) (*repository.Node, error) {
	repo := repository.NewPostgresNodeRepository(pool)
	return repo.Create(ctx, &repository.Node{
		TenantID:   tenantID,
		NodeTypeID: nodeTypeID,
		Data:       data,
	})
}

// CreateTestRelationship creates a test relationship
func CreateTestRelationship(ctx context.Context, pool *pgxpool.Pool, tenantID, sourceNodeID, targetNodeID, relType, data string) (*repository.Relationship, error) {
	repo := repository.NewPostgresRelationshipRepository(pool)
	return repo.Create(ctx, &repository.Relationship{
		TenantID:         tenantID,
		SourceNodeID:     sourceNodeID,
		TargetNodeID:     targetNodeID,
		RelationshipType: relType,
		Data:             data,
	})
}

