#!/bin/bash

# Quick test script for basic operations
# Make sure the server is running on http://localhost:5001 (Docker) or http://localhost:5000 (local)

# Check if port 5001 is responding (Docker), otherwise try 5000 (local)
if curl -s http://localhost:5001/health > /dev/null 2>&1; then
    BASE_URL="http://localhost:5001/jsonrpc"
    HEALTH_URL="http://localhost:5001/health"
    PORT=5001
elif curl -s http://localhost:5000/health > /dev/null 2>&1; then
    BASE_URL="http://localhost:5000/jsonrpc"
    HEALTH_URL="http://localhost:5000/health"
    PORT=5000
else
    echo "Error: Server not responding on port 5001 (Docker) or 5000 (local)"
    exit 1
fi

echo "Testing server on port $PORT"

echo "=== Testing flex-db Basic Operations ==="
echo ""

# Test 1: Health Check
echo "1. Testing health check..."
curl -s $HEALTH_URL | jq '.'
echo ""
echo "---"
echo ""

# Test 2: Create Tenant 1
echo "2. Creating tenant: acme-corp..."
TENANT1_RESPONSE=$(curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {
      "slug": "acme-corp",
      "name": "Acme Corporation"
    },
    "id": 1
  }')
echo "$TENANT1_RESPONSE" | jq '.'
TENANT1_ID=$(echo "$TENANT1_RESPONSE" | jq -r '.result.tenant.id')
echo "Tenant 1 ID: $TENANT1_ID"
echo ""
echo "---"
echo ""

# Test 3: Create Tenant 2
echo "3. Creating tenant: tech-startup..."
TENANT2_RESPONSE=$(curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {
      "slug": "tech-startup",
      "name": "Tech Startup Inc"
    },
    "id": 2
  }')
echo "$TENANT2_RESPONSE" | jq '.'
TENANT2_ID=$(echo "$TENANT2_RESPONSE" | jq -r '.result.tenant.id')
echo "Tenant 2 ID: $TENANT2_ID"
echo ""
echo "---"
echo ""

# Test 4: List All Tenants
echo "4. Listing all tenants..."
curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tenants",
    "params": {
      "pagination": {
        "page_size": 10
      }
    },
    "id": 3
  }' | jq '.'
echo ""
echo "---"
echo ""

# Test 5: Create User 1
echo "5. Creating user: john.doe@example.com..."
USER1_RESPONSE=$(curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_user",
    "params": {
      "email": "john.doe@example.com",
      "display_name": "John Doe"
    },
    "id": 4
  }')
echo "$USER1_RESPONSE" | jq '.'
USER1_ID=$(echo "$USER1_RESPONSE" | jq -r '.result.user.id')
echo "User 1 ID: $USER1_ID"
echo ""
echo "---"
echo ""

# Test 6: Create User 2
echo "6. Creating user: jane.smith@example.com..."
USER2_RESPONSE=$(curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_user",
    "params": {
      "email": "jane.smith@example.com",
      "display_name": "Jane Smith"
    },
    "id": 5
  }')
echo "$USER2_RESPONSE" | jq '.'
USER2_ID=$(echo "$USER2_RESPONSE" | jq -r '.result.user.id')
echo "User 2 ID: $USER2_ID"
echo ""
echo "---"
echo ""

# Test 7: List All Users
echo "7. Listing all users..."
curl -s -X POST $BASE_URL \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_users",
    "params": {
      "pagination": {
        "page_size": 10
      }
    },
    "id": 6
  }' | jq '.'
echo ""
echo "---"
echo ""

# Test 8: Add User 1 to Tenant 1
if [ "$TENANT1_ID" != "null" ] && [ "$USER1_ID" != "null" ]; then
  echo "8. Adding user 1 to tenant 1..."
  curl -s -X POST $BASE_URL \
    -H "Content-Type: application/json" \
    -d "{
      \"jsonrpc\": \"2.0\",
      \"method\": \"add_user_to_tenant\",
      \"params\": {
        \"tenant_id\": \"$TENANT1_ID\",
        \"user_id\": \"$USER1_ID\",
        \"role\": \"admin\"
      },
      \"id\": 7
    }" | jq '.'
  echo ""
  echo "---"
  echo ""
fi

# Test 9: Get Tenant 1
if [ "$TENANT1_ID" != "null" ]; then
  echo "9. Getting tenant 1 by ID..."
  curl -s -X POST $BASE_URL \
    -H "Content-Type: application/json" \
    -d "{
      \"jsonrpc\": \"2.0\",
      \"method\": \"get_tenant\",
      \"params\": {
        \"id\": \"$TENANT1_ID\"
      },
      \"id\": 8
    }" | jq '.'
  echo ""
  echo "---"
  echo ""
fi

# Test 10: Get User 1
if [ "$USER1_ID" != "null" ]; then
  echo "10. Getting user 1 by ID..."
  curl -s -X POST $BASE_URL \
    -H "Content-Type: application/json" \
    -d "{
      \"jsonrpc\": \"2.0\",
      \"method\": \"get_user\",
      \"params\": {
        \"id\": \"$USER1_ID\"
      },
      \"id\": 9
    }" | jq '.'
  echo ""
fi

echo ""
echo "=== Tests Complete ==="
echo ""
echo "Summary:"
echo "- Tenant 1 ID: $TENANT1_ID"
echo "- Tenant 2 ID: $TENANT2_ID"
echo "- User 1 ID: $USER1_ID"
echo "- User 2 ID: $USER2_ID"

