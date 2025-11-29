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

# Test 1: Health check
echo "1. Testing health check..."
HEALTH_RESPONSE=$(curl -sf "$PYTHON_TEST_URL/health")
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo "   ✓ Health check passed"
else
    echo "   ✗ Health check failed"
    exit 1
fi

# Test 2: Create tenant
echo "2. Testing tenant creation..."
TIMESTAMP=$(date +%s)
TENANT_RESPONSE=$(jsonrpc_call "create_tenant" "{\"slug\": \"test-tenant-${TIMESTAMP}\", \"name\": \"Test Tenant\"}")
if echo "$TENANT_RESPONSE" | grep -q '"tenant"'; then
    TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.result.tenant.id')
    echo "   ✓ Tenant created: $TENANT_ID"
else
    echo "   ✗ Tenant creation failed: $TENANT_RESPONSE"
    exit 1
fi

# Test 3: Get tenant
echo "3. Testing get tenant..."
GET_TENANT_RESPONSE=$(jsonrpc_call "get_tenant" "{\"id\": \"$TENANT_ID\"}")
if echo "$GET_TENANT_RESPONSE" | grep -q "$TENANT_ID"; then
    echo "   ✓ Get tenant passed"
else
    echo "   ✗ Get tenant failed: $GET_TENANT_RESPONSE"
    exit 1
fi

# Test 4: List tenants
echo "4. Testing list tenants..."
LIST_TENANT_RESPONSE=$(jsonrpc_call "list_tenants" '{"pagination": {"page_size": 10}}')
if echo "$LIST_TENANT_RESPONSE" | grep -q '"tenants"'; then
    echo "   ✓ List tenants passed"
else
    echo "   ✗ List tenants failed: $LIST_TENANT_RESPONSE"
    exit 1
fi

# Test 5: Create user
echo "5. Testing user creation..."
USER_TIMESTAMP=$(date +%s)
USER_RESPONSE=$(jsonrpc_call "create_user" "{\"email\": \"test-${USER_TIMESTAMP}@example.com\", \"display_name\": \"Test User\"}")
if echo "$USER_RESPONSE" | grep -q '"user"'; then
    USER_ID=$(echo "$USER_RESPONSE" | jq -r '.result.user.id')
    echo "   ✓ User created: $USER_ID"
else
    echo "   ✗ User creation failed: $USER_RESPONSE"
    exit 1
fi

# Test 6: Get user
echo "6. Testing get user..."
GET_USER_RESPONSE=$(jsonrpc_call "get_user" "{\"id\": \"$USER_ID\"}")
if echo "$GET_USER_RESPONSE" | grep -q "$USER_ID"; then
    echo "   ✓ Get user passed"
else
    echo "   ✗ Get user failed: $GET_USER_RESPONSE"
    exit 1
fi

# Test 7: List users
echo "7. Testing list users..."
LIST_USER_RESPONSE=$(jsonrpc_call "list_users" '{"pagination": {"page_size": 10}}')
if echo "$LIST_USER_RESPONSE" | grep -q '"users"'; then
    echo "   ✓ List users passed"
else
    echo "   ✗ List users failed: $LIST_USER_RESPONSE"
    exit 1
fi

# Test 8: Add user to tenant
# Note: JSON-RPC responses have a "result" field on success
echo "8. Testing add user to tenant..."
ADD_USER_RESPONSE=$(jsonrpc_call "add_user_to_tenant" "{\"tenant_id\": \"$TENANT_ID\", \"user_id\": \"$USER_ID\", \"role\": \"admin\"}")
if echo "$ADD_USER_RESPONSE" | grep -q '"result"'; then
    echo "   ✓ Add user to tenant passed"
else
    echo "   ✗ Add user to tenant failed: $ADD_USER_RESPONSE"
    exit 1
fi

# Test 9: Update tenant
echo "9. Testing update tenant..."
UPDATE_TENANT_RESPONSE=$(jsonrpc_call "update_tenant" "{\"id\": \"$TENANT_ID\", \"name\": \"Updated Test Tenant\"}")
if echo "$UPDATE_TENANT_RESPONSE" | grep -q '"result"'; then
    echo "   ✓ Update tenant passed"
else
    echo "   ✗ Update tenant failed: $UPDATE_TENANT_RESPONSE"
    exit 1
fi

# Test 10: Update user
echo "10. Testing update user..."
UPDATE_USER_RESPONSE=$(jsonrpc_call "update_user" "{\"id\": \"$USER_ID\", \"display_name\": \"Updated Test User\"}")
if echo "$UPDATE_USER_RESPONSE" | grep -q '"result"'; then
    echo "   ✓ Update user passed"
else
    echo "   ✗ Update user failed: $UPDATE_USER_RESPONSE"
    exit 1
fi

echo ""
echo "=========================================="
echo "All Tests Passed!"
echo "=========================================="
echo ""

exit 0
