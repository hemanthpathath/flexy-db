package grpc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	pb "github.com/hemanthpathath/flex-db/go/api/proto"
	"github.com/hemanthpathath/flex-db/go/internal/db"
	grpchandlers "github.com/hemanthpathath/flex-db/go/internal/grpc"
	"github.com/hemanthpathath/flex-db/go/internal/repository"
	"github.com/hemanthpathath/flex-db/go/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// setupTestServer creates a test gRPC server with real database
func setupTestServer(t *testing.T) (pb.TenantServiceClient, pb.UserServiceClient, pb.NodeTypeServiceClient, pb.NodeServiceClient, pb.RelationshipServiceClient, func()) {
	t.Helper()

	ctx := context.Background()

	// Setup test database
	cfg := db.Config{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnvInt("TEST_DB_PORT", 5432),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", "dbaas"),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}

	pool, err := db.Connect(ctx, cfg)
	require.NoError(t, err)

	// Run migrations
	err = db.RunMigrations(ctx, pool)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		cleanupTestDB(t, pool)
		pool.Close()
	}

	// Initialize repositories
	tenantRepo := repository.NewPostgresTenantRepository(pool)
	userRepo := repository.NewPostgresUserRepository(pool)
	nodeTypeRepo := repository.NewPostgresNodeTypeRepository(pool)
	nodeRepo := repository.NewPostgresNodeRepository(pool)
	relationshipRepo := repository.NewPostgresRelationshipRepository(pool)

	// Initialize services
	tenantSvc := service.NewTenantService(tenantRepo)
	userSvc := service.NewUserService(userRepo)
	nodeTypeSvc := service.NewNodeTypeService(nodeTypeRepo)
	nodeSvc := service.NewNodeService(nodeRepo, nodeTypeRepo)
	relationshipSvc := service.NewRelationshipService(relationshipRepo, nodeRepo)

	// Initialize gRPC handlers
	tenantHandler := grpchandlers.NewTenantHandler(tenantSvc)
	userHandler := grpchandlers.NewUserHandler(userSvc)
	nodeTypeHandler := grpchandlers.NewNodeTypeHandler(nodeTypeSvc)
	nodeHandler := grpchandlers.NewNodeHandler(nodeSvc)
	relationshipHandler := grpchandlers.NewRelationshipHandler(relationshipSvc)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterTenantServiceServer(grpcServer, tenantHandler)
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterNodeTypeServiceServer(grpcServer, nodeTypeHandler)
	pb.RegisterNodeServiceServer(grpcServer, nodeHandler)
	pb.RegisterRelationshipServiceServer(grpcServer, relationshipHandler)

	// Start server on random port
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Create clients
	// Use grpc.Dial instead of grpc.NewClient (not available in this gRPC version)
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	tenantClient := pb.NewTenantServiceClient(conn)
	userClient := pb.NewUserServiceClient(conn)
	nodeTypeClient := pb.NewNodeTypeServiceClient(conn)
	nodeClient := pb.NewNodeServiceClient(conn)
	relationshipClient := pb.NewRelationshipServiceClient(conn)

	// Enhanced cleanup
	cleanupWithConn := func() {
		conn.Close()
		grpcServer.Stop()
		cleanup()
	}

	return tenantClient, userClient, nodeTypeClient, nodeClient, relationshipClient, cleanupWithConn
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func cleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"relationships",
		"nodes",
		"node_types",
		"tenant_users",
		"users",
		"tenants",
	}
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		pool.Exec(ctx, query)
	}
}

// TestTenantService_E2E tests the full tenant service through gRPC
func TestTenantService_E2E(t *testing.T) {
	tenantClient, _, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("create tenant", func(t *testing.T) {
		resp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
			Slug: "e2e-tenant",
			Name: "E2E Test Tenant",
		})
		require.NoError(t, err)
		require.NotNil(t, resp.Tenant)
		assert.Equal(t, "e2e-tenant", resp.Tenant.Slug)
		assert.Equal(t, "E2E Test Tenant", resp.Tenant.Name)
		assert.Equal(t, "active", resp.Tenant.Status)
	})

	t.Run("get tenant", func(t *testing.T) {
		// Create first
		createResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
			Slug: "get-tenant",
			Name: "Get Tenant",
		})
		require.NoError(t, err)

		// Get it
		getResp, err := tenantClient.GetTenant(ctx, &pb.GetTenantRequest{
			Id: createResp.Tenant.Id,
		})
		require.NoError(t, err)
		assert.Equal(t, createResp.Tenant.Id, getResp.Tenant.Id)
		assert.Equal(t, "get-tenant", getResp.Tenant.Slug)
	})

	t.Run("get tenant not found", func(t *testing.T) {
		// Use a valid UUID format that doesn't exist
		_, err := tenantClient.GetTenant(ctx, &pb.GetTenantRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		})
		require.Error(t, err)
		// Check for either "not found" or "InvalidArgument" (for invalid UUID)
		assert.True(t, 
			contains(err.Error(), "not found") || 
			contains(err.Error(), "InvalidArgument") ||
			contains(err.Error(), "invalid"),
			"Expected 'not found' or 'InvalidArgument' error, got: %v", err)
	})

	t.Run("list tenants with pagination", func(t *testing.T) {
		// Create multiple tenants
		for i := 0; i < 5; i++ {
			_, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
				Slug: fmt.Sprintf("list-tenant-%d", i),
				Name: fmt.Sprintf("List Tenant %d", i),
			})
			require.NoError(t, err)
		}

		// List with pagination
		listResp, err := tenantClient.ListTenants(ctx, &pb.ListTenantsRequest{
			Pagination: &pb.Pagination{
				PageSize: 3,
			},
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Tenants), 3)
		assert.NotNil(t, listResp.Pagination)
	})
}

// TestUserService_E2E tests the full user service through gRPC
func TestUserService_E2E(t *testing.T) {
	_, userClient, _, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("create user", func(t *testing.T) {
		resp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Email:       "e2e@example.com",
			DisplayName: "E2E User",
		})
		require.NoError(t, err)
		require.NotNil(t, resp.User)
		assert.Equal(t, "e2e@example.com", resp.User.Email)
		assert.Equal(t, "E2E User", resp.User.DisplayName)
	})

	t.Run("create user validation error", func(t *testing.T) {
		_, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Email:       "",
			DisplayName: "No Email",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("get user", func(t *testing.T) {
		createResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Email:       "get@example.com",
			DisplayName: "Get User",
		})
		require.NoError(t, err)

		getResp, err := userClient.GetUser(ctx, &pb.GetUserRequest{
			Id: createResp.User.Id,
		})
		require.NoError(t, err)
		assert.Equal(t, createResp.User.Id, getResp.User.Id)
	})

	t.Run("add user to tenant", func(t *testing.T) {
		// Create tenant and user
		tenantClient, _, _, _, _, _ := setupTestServer(t)
		tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
			Slug: "user-tenant",
			Name: "User Tenant",
		})
		require.NoError(t, err)

		userResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Email:       "tenant-user@example.com",
			DisplayName: "Tenant User",
		})
		require.NoError(t, err)

		// Add user to tenant
		addResp, err := userClient.AddUserToTenant(ctx, &pb.AddUserToTenantRequest{
			TenantId: tenantResp.Tenant.Id,
			UserId:   userResp.User.Id,
			Role:     "admin",
		})
		require.NoError(t, err)
		assert.Equal(t, tenantResp.Tenant.Id, addResp.TenantUser.TenantId)
		assert.Equal(t, userResp.User.Id, addResp.TenantUser.UserId)
		assert.Equal(t, "admin", addResp.TenantUser.Role)
	})
}

// TestNodeTypeService_E2E tests the full node type service through gRPC
func TestNodeTypeService_E2E(t *testing.T) {
	tenantClient, _, nodeTypeClient, _, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant first
	tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
		Slug: "nodetype-tenant",
		Name: "NodeType Tenant",
	})
	require.NoError(t, err)

	t.Run("create node type", func(t *testing.T) {
		resp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
			TenantId:    tenantResp.Tenant.Id,
			Name:        "Task",
			Description: "A task node type",
			Schema:      `{"type": "object", "properties": {"title": {"type": "string"}}}`,
		})
		require.NoError(t, err)
		require.NotNil(t, resp.NodeType)
		assert.Equal(t, "Task", resp.NodeType.Name)
		assert.Equal(t, tenantResp.Tenant.Id, resp.NodeType.TenantId)
	})

	t.Run("get node type", func(t *testing.T) {
		createResp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
			TenantId: tenantResp.Tenant.Id,
			Name:     "Note",
			Schema:   `{"type": "object"}`,
		})
		require.NoError(t, err)

		getResp, err := nodeTypeClient.GetNodeType(ctx, &pb.GetNodeTypeRequest{
			TenantId: tenantResp.Tenant.Id,
			Id:       createResp.NodeType.Id,
		})
		require.NoError(t, err)
		assert.Equal(t, createResp.NodeType.Id, getResp.NodeType.Id)
		assert.Equal(t, "Note", getResp.NodeType.Name)
	})

	t.Run("get node type wrong tenant", func(t *testing.T) {
		// Create another tenant
		tenant2Resp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
			Slug: "tenant-2",
			Name: "Tenant 2",
		})
		require.NoError(t, err)

		// Create node type in tenant1
		createResp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
			TenantId: tenantResp.Tenant.Id,
			Name:     "Isolated",
			Schema:   `{"type": "object"}`,
		})
		require.NoError(t, err)

		// Try to get it using tenant2's ID - should fail
		_, err = nodeTypeClient.GetNodeType(ctx, &pb.GetNodeTypeRequest{
			TenantId: tenant2Resp.Tenant.Id,
			Id:       createResp.NodeType.Id,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestNodeService_E2E tests the full node service through gRPC
func TestNodeService_E2E(t *testing.T) {
	tenantClient, _, nodeTypeClient, nodeClient, _, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Create tenant and node type
	tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
		Slug: "node-tenant",
		Name: "Node Tenant",
	})
	require.NoError(t, err)

	nodeTypeResp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
		TenantId: tenantResp.Tenant.Id,
		Name:     "Task",
		Schema:   `{"type": "object"}`,
	})
	require.NoError(t, err)

	t.Run("create node", func(t *testing.T) {
		resp, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
			TenantId:   tenantResp.Tenant.Id,
			NodeTypeId: nodeTypeResp.NodeType.Id,
			Data:       `{"title": "Test Task", "priority": "high"}`,
		})
		require.NoError(t, err)
		require.NotNil(t, resp.Node)
		assert.Equal(t, tenantResp.Tenant.Id, resp.Node.TenantId)
		assert.Equal(t, nodeTypeResp.NodeType.Id, resp.Node.NodeTypeId)
		assert.Contains(t, resp.Node.Data, "Test Task")
	})

	t.Run("list nodes with pagination", func(t *testing.T) {
		// Ensure we have a valid tenant and node type for this test
		// (tenantResp and nodeTypeResp are from parent test setup)
		require.NotEmpty(t, tenantResp.Tenant.Id)
		require.NotEmpty(t, nodeTypeResp.NodeType.Id)

		// Create multiple nodes
		for i := 0; i < 5; i++ {
			_, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
				TenantId:   tenantResp.Tenant.Id,
				NodeTypeId: nodeTypeResp.NodeType.Id,
				Data:       fmt.Sprintf(`{"title": "Task %d"}`, i),
			})
			require.NoError(t, err, "Failed to create node %d: tenant=%s, nodeType=%s", i, tenantResp.Tenant.Id, nodeTypeResp.NodeType.Id)
		}

		listResp, err := nodeClient.ListNodes(ctx, &pb.ListNodesRequest{
			TenantId: tenantResp.Tenant.Id,
			Pagination: &pb.Pagination{
				PageSize: 3,
			},
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Nodes), 3)
		assert.NotNil(t, listResp.Pagination)
	})
}

// TestRelationshipService_E2E tests the full relationship service through gRPC
func TestRelationshipService_E2E(t *testing.T) {
	tenantClient, _, nodeTypeClient, nodeClient, relationshipClient, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: Create tenant, node type, and nodes
	tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
		Slug: "rel-tenant",
		Name: "Relationship Tenant",
	})
	require.NoError(t, err)

	nodeTypeResp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
		TenantId: tenantResp.Tenant.Id,
		Name:     "Task",
		Schema:   `{"type": "object"}`,
	})
	require.NoError(t, err)

	sourceResp, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
		TenantId:   tenantResp.Tenant.Id,
		NodeTypeId: nodeTypeResp.NodeType.Id,
		Data:       `{"title": "Source Task"}`,
	})
	require.NoError(t, err)

	targetResp, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
		TenantId:   tenantResp.Tenant.Id,
		NodeTypeId: nodeTypeResp.NodeType.Id,
		Data:       `{"title": "Target Task"}`,
	})
	require.NoError(t, err)

	t.Run("create relationship", func(t *testing.T) {
		resp, err := relationshipClient.CreateRelationship(ctx, &pb.CreateRelationshipRequest{
			TenantId:         tenantResp.Tenant.Id,
			SourceNodeId:     sourceResp.Node.Id,
			TargetNodeId:     targetResp.Node.Id,
			RelationshipType: "depends_on",
			Data:             `{"priority": 1}`,
		})
		require.NoError(t, err)
		require.NotNil(t, resp.Relationship)
		assert.Equal(t, sourceResp.Node.Id, resp.Relationship.SourceNodeId)
		assert.Equal(t, targetResp.Node.Id, resp.Relationship.TargetNodeId)
		assert.Equal(t, "depends_on", resp.Relationship.RelationshipType)
	})

	t.Run("list relationships", func(t *testing.T) {
		listResp, err := relationshipClient.ListRelationships(ctx, &pb.ListRelationshipsRequest{
			TenantId: tenantResp.Tenant.Id,
			Pagination: &pb.Pagination{
				PageSize: 10,
			},
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listResp.Relationships), 1)
	})
}

// TestE2E_CompleteWorkflow tests a complete user journey end-to-end
func TestE2E_CompleteWorkflow(t *testing.T) {
	tenantClient, userClient, nodeTypeClient, nodeClient, relationshipClient, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Step 1: Create tenant
	tenantResp, err := tenantClient.CreateTenant(ctx, &pb.CreateTenantRequest{
		Slug: "workflow-tenant",
		Name: "Workflow Tenant",
	})
	require.NoError(t, err)

	// Step 2: Create user
	userResp, err := userClient.CreateUser(ctx, &pb.CreateUserRequest{
		Email:       "workflow@example.com",
		DisplayName: "Workflow User",
	})
	require.NoError(t, err)

	// Step 3: Add user to tenant
	_, err = userClient.AddUserToTenant(ctx, &pb.AddUserToTenantRequest{
		TenantId: tenantResp.Tenant.Id,
		UserId:   userResp.User.Id,
		Role:     "admin",
	})
	require.NoError(t, err)

	// Step 4: Create node type
	nodeTypeResp, err := nodeTypeClient.CreateNodeType(ctx, &pb.CreateNodeTypeRequest{
		TenantId:    tenantResp.Tenant.Id,
		Name:        "Project",
		Description: "A project node",
		Schema:      `{"type": "object", "properties": {"name": {"type": "string"}}}`,
	})
	require.NoError(t, err)

	// Step 5: Create nodes
	project1Resp, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
		TenantId:   tenantResp.Tenant.Id,
		NodeTypeId: nodeTypeResp.NodeType.Id,
		Data:       `{"name": "Project Alpha"}`,
	})
	require.NoError(t, err)

	project2Resp, err := nodeClient.CreateNode(ctx, &pb.CreateNodeRequest{
		TenantId:   tenantResp.Tenant.Id,
		NodeTypeId: nodeTypeResp.NodeType.Id,
		Data:       `{"name": "Project Beta"}`,
	})
	require.NoError(t, err)

	// Step 6: Create relationship
	relResp, err := relationshipClient.CreateRelationship(ctx, &pb.CreateRelationshipRequest{
		TenantId:         tenantResp.Tenant.Id,
		SourceNodeId:     project1Resp.Node.Id,
		TargetNodeId:     project2Resp.Node.Id,
		RelationshipType: "depends_on",
		Data:             `{"priority": "high"}`,
	})
	require.NoError(t, err)

	// Step 7: Verify everything
	assert.NotEmpty(t, tenantResp.Tenant.Id)
	assert.NotEmpty(t, userResp.User.Id)
	assert.NotEmpty(t, nodeTypeResp.NodeType.Id)
	assert.NotEmpty(t, project1Resp.Node.Id)
	assert.NotEmpty(t, project2Resp.Node.Id)
	assert.NotEmpty(t, relResp.Relationship.Id)
}

