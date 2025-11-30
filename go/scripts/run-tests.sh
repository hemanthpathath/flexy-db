#!/bin/bash
#
# flex-db Test Runner Script
#
# This script runs all integration tests for the Go backend
# and API tests for the Python backend.
#
# Environment variables:
#   TEST_DB_HOST - PostgreSQL host (default: localhost)
#   TEST_DB_PORT - PostgreSQL port (default: 5432)
#   TEST_DB_USER - PostgreSQL user (default: postgres)
#   TEST_DB_PASSWORD - PostgreSQL password (default: postgres)
#   TEST_DB_NAME - Test database name (default: dbaas_test)
#   TEST_DB_SSL_MODE - SSL mode (default: disable)
#   PYTHON_TEST_URL - Python backend URL for API tests
#

set -e

echo "=========================================="
echo "flex-db Integration Test Runner"
echo "=========================================="
echo ""

# Default values
TEST_DB_HOST="${TEST_DB_HOST:-localhost}"
TEST_DB_PORT="${TEST_DB_PORT:-5432}"
TEST_DB_USER="${TEST_DB_USER:-postgres}"
TEST_DB_PASSWORD="${TEST_DB_PASSWORD:-postgres}"
TEST_DB_NAME="${TEST_DB_NAME:-dbaas_test}"
TEST_DB_SSL_MODE="${TEST_DB_SSL_MODE:-disable}"
PYTHON_TEST_URL="${PYTHON_TEST_URL:-http://localhost:5000}"

echo "Test Configuration:"
echo "  DB Host: $TEST_DB_HOST"
echo "  DB Port: $TEST_DB_PORT"
echo "  DB Name: $TEST_DB_NAME"
echo "  Python URL: $PYTHON_TEST_URL"
echo ""

# Wait for database to be ready using TCP check
echo "Waiting for database to be ready..."
for i in {1..60}; do
    # Try TCP connection to database port
    if timeout 1 bash -c "echo > /dev/tcp/$TEST_DB_HOST/$TEST_DB_PORT" 2>/dev/null; then
        echo "Database port is open, waiting for it to be ready..."
        sleep 2
        echo "Database is ready!"
        break
    fi
    if [ $i -eq 60 ]; then
        echo "ERROR: Database not ready after 60 seconds"
        exit 1
    fi
    echo "  Waiting... ($i/60)"
    sleep 1
done

echo ""
echo "=========================================="
echo "Running Go Integration Tests"
echo "=========================================="
echo ""

# Export test environment variables
export TEST_DB_HOST TEST_DB_PORT TEST_DB_USER TEST_DB_PASSWORD TEST_DB_NAME TEST_DB_SSL_MODE

# Run Go integration tests with verbose output
cd /app
go test -v -count=1 ./integration/... 2>&1 || {
    echo ""
    echo "ERROR: Go integration tests failed!"
    exit 1
}

echo ""
echo "=========================================="
echo "Running Python Backend API Tests"
echo "=========================================="
echo ""

# Wait for Python backend to be ready
echo "Waiting for Python backend..."
for i in {1..30}; do
    if curl -sf "$PYTHON_TEST_URL/health" > /dev/null 2>&1; then
        echo "Python backend is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "WARNING: Python backend not ready after 30 seconds, skipping API tests"
        exit 0
    fi
    echo "  Waiting... ($i/30)"
    sleep 1
done

echo ""
echo "Testing Python Backend CRUD Operations..."
echo ""

# Helper function for JSON-RPC calls
jsonrpc_call() {
    local method="$1"
    local params="$2"
    local id="${3:-1}"
    
    curl -sf -X POST "$PYTHON_TEST_URL/jsonrpc" \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\": \"2.0\", \"method\": \"$method\", \"params\": $params, \"id\": $id}"
}

# Track test number
TEST_NUM=0

run_test() {
    local name="$1"
    local method="$2"
    local params="$3"
    local expected="$4"
    
    TEST_NUM=$((TEST_NUM + 1))
    echo "$TEST_NUM. Testing $name..."
    
    RESPONSE=$(jsonrpc_call "$method" "$params")
    if echo "$RESPONSE" | grep -q "$expected"; then
        echo "   ✓ $name passed"
    else
        echo "   ✗ $name failed: $RESPONSE"
        exit 1
    fi
}

# ====================
# Health Check
# ====================
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing health check..."
HEALTH_RESPONSE=$(curl -sf "$PYTHON_TEST_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo "   ✓ Health check passed"
else
    echo "   ✗ Health check failed"
    exit 1
fi

# ====================
# Tenant CRUD Tests
# ====================
echo ""
echo "--- Tenant CRUD Tests ---"

# Create tenant
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_tenant..."
TIMESTAMP=$(date +%s)
TENANT_RESPONSE=$(jsonrpc_call "create_tenant" "{\"slug\": \"test-tenant-${TIMESTAMP}\", \"name\": \"Test Tenant\"}")
if echo "$TENANT_RESPONSE" | grep -q '"tenant"'; then
    TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.result.tenant.id')
    echo "   ✓ create_tenant passed: $TENANT_ID"
else
    echo "   ✗ create_tenant failed: $TENANT_RESPONSE"
    exit 1
fi

# Get tenant
run_test "get_tenant" "get_tenant" "{\"id\": \"$TENANT_ID\"}" '"tenant"'

# List tenants
run_test "list_tenants" "list_tenants" '{"pagination": {"page_size": 10}}' '"tenants"'

# Update tenant
run_test "update_tenant" "update_tenant" "{\"id\": \"$TENANT_ID\", \"name\": \"Updated Test Tenant\"}" '"tenant"'

# ====================
# User CRUD Tests
# ====================
echo ""
echo "--- User CRUD Tests ---"

# Create user
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_user..."
USER_TIMESTAMP=$(date +%s)
USER_RESPONSE=$(jsonrpc_call "create_user" "{\"email\": \"test-${USER_TIMESTAMP}@example.com\", \"display_name\": \"Test User\"}")
if echo "$USER_RESPONSE" | grep -q '"user"'; then
    USER_ID=$(echo "$USER_RESPONSE" | jq -r '.result.user.id')
    echo "   ✓ create_user passed: $USER_ID"
else
    echo "   ✗ create_user failed: $USER_RESPONSE"
    exit 1
fi

# Get user
run_test "get_user" "get_user" "{\"id\": \"$USER_ID\"}" '"user"'

# List users
run_test "list_users" "list_users" '{"pagination": {"page_size": 10}}' '"users"'

# Update user
run_test "update_user" "update_user" "{\"id\": \"$USER_ID\", \"display_name\": \"Updated Test User\"}" '"user"'

# Add user to tenant
run_test "add_user_to_tenant" "add_user_to_tenant" "{\"tenant_id\": \"$TENANT_ID\", \"user_id\": \"$USER_ID\", \"role\": \"admin\"}" '"tenant_user"'

# List tenant users
run_test "list_tenant_users" "list_tenant_users" "{\"tenant_id\": \"$TENANT_ID\", \"pagination\": {\"page_size\": 10}}" '"tenant_users"'

# Remove user from tenant
run_test "remove_user_from_tenant" "remove_user_from_tenant" "{\"tenant_id\": \"$TENANT_ID\", \"user_id\": \"$USER_ID\"}" '"result"'

# ====================
# NodeType CRUD Tests
# ====================
echo ""
echo "--- NodeType CRUD Tests ---"

# Create node type
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_node_type..."
NODE_TYPE_RESPONSE=$(jsonrpc_call "create_node_type" "{\"tenant_id\": \"$TENANT_ID\", \"name\": \"TestNodeType\", \"description\": \"A test node type\", \"schema\": \"{}\"}")
if echo "$NODE_TYPE_RESPONSE" | grep -q '"node_type"'; then
    NODE_TYPE_ID=$(echo "$NODE_TYPE_RESPONSE" | jq -r '.result.node_type.id')
    echo "   ✓ create_node_type passed: $NODE_TYPE_ID"
else
    echo "   ✗ create_node_type failed: $NODE_TYPE_RESPONSE"
    exit 1
fi

# Get node type
run_test "get_node_type" "get_node_type" "{\"id\": \"$NODE_TYPE_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"node_type"'

# List node types
run_test "list_node_types" "list_node_types" "{\"tenant_id\": \"$TENANT_ID\", \"pagination\": {\"page_size\": 10}}" '"node_types"'

# Update node type
run_test "update_node_type" "update_node_type" "{\"id\": \"$NODE_TYPE_ID\", \"tenant_id\": \"$TENANT_ID\", \"name\": \"UpdatedNodeType\"}" '"node_type"'

# ====================
# Node CRUD Tests
# ====================
echo ""
echo "--- Node CRUD Tests ---"

# Create node
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_node..."
NODE_RESPONSE=$(jsonrpc_call "create_node" "{\"tenant_id\": \"$TENANT_ID\", \"node_type_id\": \"$NODE_TYPE_ID\", \"data\": \"{\\\"key\\\": \\\"value\\\"}\"}")
if echo "$NODE_RESPONSE" | grep -q '"node"'; then
    NODE_ID=$(echo "$NODE_RESPONSE" | jq -r '.result.node.id')
    echo "   ✓ create_node passed: $NODE_ID"
else
    echo "   ✗ create_node failed: $NODE_RESPONSE"
    exit 1
fi

# Get node
run_test "get_node" "get_node" "{\"id\": \"$NODE_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"node"'

# List nodes
run_test "list_nodes" "list_nodes" "{\"tenant_id\": \"$TENANT_ID\", \"pagination\": {\"page_size\": 10}}" '"nodes"'

# List nodes by node type
run_test "list_nodes (by node_type_id)" "list_nodes" "{\"tenant_id\": \"$TENANT_ID\", \"node_type_id\": \"$NODE_TYPE_ID\", \"pagination\": {\"page_size\": 10}}" '"nodes"'

# Update node
run_test "update_node" "update_node" "{\"id\": \"$NODE_ID\", \"tenant_id\": \"$TENANT_ID\", \"data\": \"{\\\"key\\\": \\\"updated_value\\\"}\"}" '"node"'

# ====================
# Relationship CRUD Tests
# ====================
echo ""
echo "--- Relationship CRUD Tests ---"

# Create a second node for relationships
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_node (second node for relationships)..."
NODE2_RESPONSE=$(jsonrpc_call "create_node" "{\"tenant_id\": \"$TENANT_ID\", \"node_type_id\": \"$NODE_TYPE_ID\", \"data\": \"{\\\"key\\\": \\\"value2\\\"}\"}")
if echo "$NODE2_RESPONSE" | grep -q '"node"'; then
    NODE2_ID=$(echo "$NODE2_RESPONSE" | jq -r '.result.node.id')
    echo "   ✓ create_node (second) passed: $NODE2_ID"
else
    echo "   ✗ create_node (second) failed: $NODE2_RESPONSE"
    exit 1
fi

# Create relationship
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_relationship..."
REL_RESPONSE=$(jsonrpc_call "create_relationship" "{\"tenant_id\": \"$TENANT_ID\", \"source_node_id\": \"$NODE_ID\", \"target_node_id\": \"$NODE2_ID\", \"relationship_type\": \"connected_to\", \"data\": \"{}\"}")
if echo "$REL_RESPONSE" | grep -q '"relationship"'; then
    REL_ID=$(echo "$REL_RESPONSE" | jq -r '.result.relationship.id')
    echo "   ✓ create_relationship passed: $REL_ID"
else
    echo "   ✗ create_relationship failed: $REL_RESPONSE"
    exit 1
fi

# Get relationship
run_test "get_relationship" "get_relationship" "{\"id\": \"$REL_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"relationship"'

# List relationships
run_test "list_relationships" "list_relationships" "{\"tenant_id\": \"$TENANT_ID\", \"pagination\": {\"page_size\": 10}}" '"relationships"'

# List relationships by source node
run_test "list_relationships (by source)" "list_relationships" "{\"tenant_id\": \"$TENANT_ID\", \"source_node_id\": \"$NODE_ID\", \"pagination\": {\"page_size\": 10}}" '"relationships"'

# List relationships by target node
run_test "list_relationships (by target)" "list_relationships" "{\"tenant_id\": \"$TENANT_ID\", \"target_node_id\": \"$NODE2_ID\", \"pagination\": {\"page_size\": 10}}" '"relationships"'

# List relationships by type
run_test "list_relationships (by type)" "list_relationships" "{\"tenant_id\": \"$TENANT_ID\", \"relationship_type\": \"connected_to\", \"pagination\": {\"page_size\": 10}}" '"relationships"'

# Update relationship
run_test "update_relationship" "update_relationship" "{\"id\": \"$REL_ID\", \"tenant_id\": \"$TENANT_ID\", \"relationship_type\": \"linked_to\"}" '"relationship"'

# ====================
# Delete Tests (cleanup)
# ====================
echo ""
echo "--- Delete Tests (cleanup) ---"

# Delete relationship
run_test "delete_relationship" "delete_relationship" "{\"id\": \"$REL_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"result"'

# Delete nodes
run_test "delete_node (second)" "delete_node" "{\"id\": \"$NODE2_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"result"'
run_test "delete_node" "delete_node" "{\"id\": \"$NODE_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"result"'

# Delete node type
run_test "delete_node_type" "delete_node_type" "{\"id\": \"$NODE_TYPE_ID\", \"tenant_id\": \"$TENANT_ID\"}" '"result"'

# Create a second user for deletion test (to avoid affecting the main user)
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_user (for delete test)..."
DEL_USER_TIMESTAMP=$(date +%s)
DEL_USER_RESPONSE=$(jsonrpc_call "create_user" "{\"email\": \"del-${DEL_USER_TIMESTAMP}@example.com\", \"display_name\": \"Delete Test User\"}")
if echo "$DEL_USER_RESPONSE" | grep -q '"user"'; then
    DEL_USER_ID=$(echo "$DEL_USER_RESPONSE" | jq -r '.result.user.id')
    echo "   ✓ create_user (for delete) passed: $DEL_USER_ID"
else
    echo "   ✗ create_user (for delete) failed: $DEL_USER_RESPONSE"
    exit 1
fi

# Delete user
run_test "delete_user" "delete_user" "{\"id\": \"$DEL_USER_ID\"}" '"result"'

# Create a second tenant for deletion test (to avoid affecting the main tenant)
TEST_NUM=$((TEST_NUM + 1))
echo "$TEST_NUM. Testing create_tenant (for delete test)..."
DEL_TENANT_TIMESTAMP=$(date +%s)
DEL_TENANT_RESPONSE=$(jsonrpc_call "create_tenant" "{\"slug\": \"del-tenant-${DEL_TENANT_TIMESTAMP}\", \"name\": \"Delete Test Tenant\"}")
if echo "$DEL_TENANT_RESPONSE" | grep -q '"tenant"'; then
    DEL_TENANT_ID=$(echo "$DEL_TENANT_RESPONSE" | jq -r '.result.tenant.id')
    echo "   ✓ create_tenant (for delete) passed: $DEL_TENANT_ID"
else
    echo "   ✗ create_tenant (for delete) failed: $DEL_TENANT_RESPONSE"
    exit 1
fi

# Delete tenant
run_test "delete_tenant" "delete_tenant" "{\"id\": \"$DEL_TENANT_ID\"}" '"result"'

echo ""
echo "=========================================="
echo "All Tests Passed! ($TEST_NUM tests)"
echo "=========================================="
echo ""

exit 0
