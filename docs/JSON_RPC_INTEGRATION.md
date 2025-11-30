# JSON-RPC Integration Guide

This guide provides comprehensive documentation for integrating with the flex-db Python backend using the JSON-RPC 2.0 API.

## Table of Contents

- [Overview](#overview)
- [Endpoint](#endpoint)
- [Request Format](#request-format)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Client Implementations](#client-implementations)
- [Available Methods](#available-methods)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

The flex-db service exposes a JSON-RPC 2.0 API at `/jsonrpc`. This is the **recommended integration method** for backend services as it provides:

- **Type-safe method calls**: Each method has well-defined parameters
- **Consistent error handling**: Standard JSON-RPC error codes
- **Batch requests**: Support for multiple method calls in a single request
- **Notifications**: Fire-and-forget requests without responses
- **Better performance**: Lower overhead compared to REST for programmatic access
- **OpenRPC Documentation**: Auto-generated API specification for interactive docs and code generation
- **Introspection**: `rpc.discover` method for dynamic API discovery

## Endpoint

```
POST http://localhost:5000/jsonrpc
Content-Type: application/json
```

**Production**: Replace `localhost:5000` with your deployed service URL.

## OpenRPC Specification

The service provides an **OpenRPC specification** (similar to OpenAPI for REST APIs) that describes all available methods, parameters, and return types.

### Accessing the Spec

**HTTP Endpoint:**
```
GET http://localhost:5000/openrpc.json
```

**JSON-RPC Introspection:**
```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "rpc.discover",
    "params": {},
    "id": 1
  }'
```

### Using OpenRPC Tools

1. **Interactive Documentation**: 
   - Copy the spec from `/openrpc.json` and paste it into [OpenRPC Playground](https://playground.open-rpc.org/)
   - Or use other OpenRPC-compatible tools
2. **Code Generation**: Generate client code from the spec using OpenRPC code generators
3. **API Validation**: Validate requests/responses against the spec
4. **Testing**: Use OpenRPC tools for automated testing

The OpenRPC spec is **auto-generated from code** using introspection, ensuring it always stays accurate and up-to-date with your codebase.

### Introspection Method

You can also get the OpenRPC spec via JSON-RPC:

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "rpc_discover",
    "params": {},
    "id": 1
  }'
```

**Note**: The method is registered as `rpc_discover` (underscore) but the OpenRPC spec shows it as `rpc.discover` (dot) for standards compliance. Both names work, but `rpc_discover` is the actual method name.

## Request Format

All requests must follow the JSON-RPC 2.0 specification:

```json
{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": {
    "param1": "value1",
    "param2": "value2"
  },
  "id": 1
}
```

### Fields

- **`jsonrpc`** (required): Must be `"2.0"`
- **`method`** (required): The method name to call (e.g., `"create_tenant"`)
- **`params`** (optional): Object containing method parameters
- **`id`** (optional): Request identifier. Use `null` for notifications (no response)

## Response Format

### Success Response

```json
{
  "jsonrpc": "2.0",
  "result": {
    "tenant": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "slug": "acme-corp",
      "name": "Acme Corporation",
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  },
  "id": 1
}
```

### Error Response

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Tenant not found"
  },
  "id": 1
}
```

## Error Handling

### Standard JSON-RPC Error Codes

| Code | Meaning | Description |
|------|---------|-------------|
| `-32700` | Parse error | Invalid JSON was received |
| `-32600` | Invalid Request | The JSON sent is not a valid Request object |
| `-32601` | Method not found | The method does not exist |
| `-32602` | Invalid params | Invalid method parameter(s) |
| `-32603` | Internal error | Internal JSON-RPC error |

### Custom Error Codes

| Code | Meaning | Description |
|------|---------|-------------|
| `-32001` | Not Found | Resource not found (e.g., tenant, user, node) |
| `-32002` | Validation Error | Input validation failed |

### Error Response Example

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Tenant with id '123' not found"
  },
  "id": 1
}
```

## Client Implementations

### Python Client

```python
import requests
import json
from typing import Any, Dict, Optional

class FlexDBClient:
    def __init__(self, base_url: str = "http://localhost:5000"):
        self.base_url = base_url
        self.endpoint = f"{base_url}/jsonrpc"
        self.request_id = 1
    
    def _call(self, method: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Make a JSON-RPC call."""
        payload = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params or {},
            "id": self.request_id
        }
        self.request_id += 1
        
        response = requests.post(
            self.endpoint,
            json=payload,
            headers={"Content-Type": "application/json"}
        )
        response.raise_for_status()
        
        result = response.json()
        
        if "error" in result:
            raise Exception(f"JSON-RPC Error {result['error']['code']}: {result['error']['message']}")
        
        return result.get("result", {})
    
    # Tenant methods
    def create_tenant(self, slug: str, name: str) -> Dict[str, Any]:
        return self._call("create_tenant", {"slug": slug, "name": name})
    
    def get_tenant(self, id: str) -> Dict[str, Any]:
        return self._call("get_tenant", {"id": id})
    
    def update_tenant(self, id: str, slug: str = "", name: str = "", status: str = "") -> Dict[str, Any]:
        return self._call("update_tenant", {"id": id, "slug": slug, "name": name, "status": status})
    
    def delete_tenant(self, id: str) -> Dict[str, Any]:
        return self._call("delete_tenant", {"id": id})
    
    def list_tenants(self, page_size: int = 10, page_token: str = "") -> Dict[str, Any]:
        return self._call("list_tenants", {
            "pagination": {"page_size": page_size, "page_token": page_token}
        })
    
    # User methods
    def create_user(self, email: str, display_name: str) -> Dict[str, Any]:
        return self._call("create_user", {"email": email, "display_name": display_name})
    
    def get_user(self, id: str) -> Dict[str, Any]:
        return self._call("get_user", {"id": id})
    
    def add_user_to_tenant(self, tenant_id: str, user_id: str, role: str = "") -> Dict[str, Any]:
        return self._call("add_user_to_tenant", {
            "tenant_id": tenant_id,
            "user_id": user_id,
            "role": role
        })
    
    # Node methods
    def create_node(self, tenant_id: str, node_type_id: str, data: str = "{}") -> Dict[str, Any]:
        return self._call("create_node", {
            "tenant_id": tenant_id,
            "node_type_id": node_type_id,
            "data": data
        })
    
    def get_node(self, id: str, tenant_id: str) -> Dict[str, Any]:
        return self._call("get_node", {"id": id, "tenant_id": tenant_id})
    
    def list_nodes(self, tenant_id: str, node_type_id: str = "", page_size: int = 10, page_token: str = "") -> Dict[str, Any]:
        return self._call("list_nodes", {
            "tenant_id": tenant_id,
            "node_type_id": node_type_id,
            "pagination": {"page_size": page_size, "page_token": page_token}
        })
    
    # Relationship methods
    def create_relationship(
        self,
        tenant_id: str,
        source_node_id: str,
        target_node_id: str,
        relationship_type: str,
        data: str = "{}"
    ) -> Dict[str, Any]:
        return self._call("create_relationship", {
            "tenant_id": tenant_id,
            "source_node_id": source_node_id,
            "target_node_id": target_node_id,
            "relationship_type": relationship_type,
            "data": data
        })


# Usage example
client = FlexDBClient("http://localhost:5000")

# Create a tenant
result = client.create_tenant("acme-corp", "Acme Corporation")
tenant = result["tenant"]
print(f"Created tenant: {tenant['id']}")

# Create a user
result = client.create_user("john@example.com", "John Doe")
user = result["user"]

# Add user to tenant
result = client.add_user_to_tenant(tenant["id"], user["id"], "admin")
```

### JavaScript/TypeScript Client

```typescript
interface JSONRPCRequest {
  jsonrpc: "2.0";
  method: string;
  params?: Record<string, any>;
  id: number | null;
}

interface JSONRPCResponse {
  jsonrpc: "2.0";
  result?: any;
  error?: {
    code: number;
    message: string;
  };
  id: number | null;
}

class FlexDBClient {
  private baseUrl: string;
  private endpoint: string;
  private requestId: number = 1;

  constructor(baseUrl: string = "http://localhost:5000") {
    this.baseUrl = baseUrl;
    this.endpoint = `${baseUrl}/jsonrpc`;
  }

  private async call(method: string, params?: Record<string, any>): Promise<any> {
    const request: JSONRPCRequest = {
      jsonrpc: "2.0",
      method,
      params: params || {},
      id: this.requestId++,
    };

    const response = await fetch(this.endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result: JSONRPCResponse = await response.json();

    if (result.error) {
      throw new Error(`JSON-RPC Error ${result.error.code}: ${result.error.message}`);
    }

    return result.result;
  }

  // Tenant methods
  async createTenant(slug: string, name: string) {
    return this.call("create_tenant", { slug, name });
  }

  async getTenant(id: string) {
    return this.call("get_tenant", { id });
  }

  async updateTenant(id: string, slug?: string, name?: string, status?: string) {
    return this.call("update_tenant", { id, slug: slug || "", name: name || "", status: status || "" });
  }

  async deleteTenant(id: string) {
    return this.call("delete_tenant", { id });
  }

  async listTenants(pageSize: number = 10, pageToken: string = "") {
    return this.call("list_tenants", {
      pagination: { page_size: pageSize, page_token: pageToken },
    });
  }

  // User methods
  async createUser(email: string, displayName: string) {
    return this.call("create_user", { email, display_name: displayName });
  }

  async getUser(id: string) {
    return this.call("get_user", { id });
  }

  async addUserToTenant(tenantId: string, userId: string, role: string = "") {
    return this.call("add_user_to_tenant", {
      tenant_id: tenantId,
      user_id: userId,
      role,
    });
  }

  // Node methods
  async createNode(tenantId: string, nodeTypeId: string, data: string = "{}") {
    return this.call("create_node", {
      tenant_id: tenantId,
      node_type_id: nodeTypeId,
      data,
    });
  }

  async getNode(id: string, tenantId: string) {
    return this.call("get_node", { id, tenant_id: tenantId });
  }

  async listNodes(tenantId: string, nodeTypeId?: string, pageSize: number = 10, pageToken: string = "") {
    return this.call("list_nodes", {
      tenant_id: tenantId,
      node_type_id: nodeTypeId || "",
      pagination: { page_size: pageSize, page_token: pageToken },
    });
  }

  // Relationship methods
  async createRelationship(
    tenantId: string,
    sourceNodeId: string,
    targetNodeId: string,
    relationshipType: string,
    data: string = "{}"
  ) {
    return this.call("create_relationship", {
      tenant_id: tenantId,
      source_node_id: sourceNodeId,
      target_node_id: targetNodeId,
      relationship_type: relationshipType,
      data,
    });
  }
}

// Usage example
const client = new FlexDBClient("http://localhost:5000");

async function example() {
  // Create a tenant
  const tenantResult = await client.createTenant("acme-corp", "Acme Corporation");
  const tenant = tenantResult.tenant;
  console.log(`Created tenant: ${tenant.id}`);

  // Create a user
  const userResult = await client.createUser("john@example.com", "John Doe");
  const user = userResult.user;

  // Add user to tenant
  await client.addUserToTenant(tenant.id, user.id, "admin");
}
```

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type JSONRPCRequest struct {
    JSONRPC string                 `json:"jsonrpc"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
    ID      int                    `json:"id"`
}

type JSONRPCResponse struct {
    JSONRPC string                 `json:"jsonrpc"`
    Result  map[string]interface{} `json:"result,omitempty"`
    Error   *JSONRPCError          `json:"error,omitempty"`
    ID      int                    `json:"id"`
}

type JSONRPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

type FlexDBClient struct {
    baseURL   string
    endpoint  string
    requestID int
    client    *http.Client
}

func NewFlexDBClient(baseURL string) *FlexDBClient {
    return &FlexDBClient{
        baseURL:   baseURL,
        endpoint:  baseURL + "/jsonrpc",
        requestID: 1,
        client:    &http.Client{},
    }
}

func (c *FlexDBClient) call(method string, params map[string]interface{}) (map[string]interface{}, error) {
    req := JSONRPCRequest{
        JSONRPC: "2.0",
        Method:  method,
        Params:  params,
        ID:      c.requestID,
    }
    c.requestID++

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var jsonResp JSONRPCResponse
    if err := json.Unmarshal(body, &jsonResp); err != nil {
        return nil, err
    }

    if jsonResp.Error != nil {
        return nil, fmt.Errorf("JSON-RPC Error %d: %s", jsonResp.Error.Code, jsonResp.Error.Message)
    }

    return jsonResp.Result, nil
}

// Tenant methods
func (c *FlexDBClient) CreateTenant(slug, name string) (map[string]interface{}, error) {
    return c.call("create_tenant", map[string]interface{}{
        "slug": slug,
        "name": name,
    })
}

func (c *FlexDBClient) GetTenant(id string) (map[string]interface{}, error) {
    return c.call("get_tenant", map[string]interface{}{
        "id": id,
    })
}

func (c *FlexDBClient) ListTenants(pageSize int, pageToken string) (map[string]interface{}, error) {
    return c.call("list_tenants", map[string]interface{}{
        "pagination": map[string]interface{}{
            "page_size":  pageSize,
            "page_token": pageToken,
        },
    })
}

// Usage example
func main() {
    client := NewFlexDBClient("http://localhost:5000")

    result, err := client.CreateTenant("acme-corp", "Acme Corporation")
    if err != nil {
        panic(err)
    }

    tenant := result["tenant"].(map[string]interface{})
    fmt.Printf("Created tenant: %v\n", tenant["id"])
}
```

## Available Methods

### Tenant Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `create_tenant` | Create a new tenant | `slug` (string), `name` (string) |
| `get_tenant` | Get tenant by ID | `id` (string) |
| `update_tenant` | Update tenant | `id` (string), `slug` (string, optional), `name` (string, optional), `status` (string, optional) |
| `delete_tenant` | Delete tenant | `id` (string) |
| `list_tenants` | List tenants with pagination | `pagination` (object, optional) |

### User Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `create_user` | Create a new user | `email` (string), `display_name` (string) |
| `get_user` | Get user by ID | `id` (string) |
| `update_user` | Update user | `id` (string), `email` (string, optional), `display_name` (string, optional) |
| `delete_user` | Delete user | `id` (string) |
| `list_users` | List users with pagination | `pagination` (object, optional) |
| `add_user_to_tenant` | Add user to tenant | `tenant_id` (string), `user_id` (string), `role` (string, optional) |
| `remove_user_from_tenant` | Remove user from tenant | `tenant_id` (string), `user_id` (string) |
| `list_tenant_users` | List users in a tenant | `tenant_id` (string), `pagination` (object, optional) |

### NodeType Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `create_node_type` | Create a new node type | `tenant_id` (string), `name` (string), `description` (string, optional), `schema` (string, optional) |
| `get_node_type` | Get node type by ID | `id` (string), `tenant_id` (string) |
| `update_node_type` | Update node type | `id` (string), `tenant_id` (string), `name` (string, optional), `description` (string, optional), `schema` (string, optional) |
| `delete_node_type` | Delete node type | `id` (string), `tenant_id` (string) |
| `list_node_types` | List node types for a tenant | `tenant_id` (string), `pagination` (object, optional) |

### Node Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `create_node` | Create a new node | `tenant_id` (string), `node_type_id` (string), `data` (string, optional, JSON) |
| `get_node` | Get node by ID | `id` (string), `tenant_id` (string) |
| `update_node` | Update node | `id` (string), `tenant_id` (string), `data` (string, optional, JSON) |
| `delete_node` | Delete node | `id` (string), `tenant_id` (string) |
| `list_nodes` | List nodes for a tenant | `tenant_id` (string), `node_type_id` (string, optional), `pagination` (object, optional) |

### Relationship Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| `create_relationship` | Create a new relationship | `tenant_id` (string), `source_node_id` (string), `target_node_id` (string), `relationship_type` (string), `data` (string, optional, JSON) |
| `get_relationship` | Get relationship by ID | `id` (string), `tenant_id` (string) |
| `update_relationship` | Update relationship | `id` (string), `tenant_id` (string), `relationship_type` (string, optional), `data` (string, optional, JSON) |
| `delete_relationship` | Delete relationship | `id` (string), `tenant_id` (string) |
| `list_relationships` | List relationships for a tenant | `tenant_id` (string), `source_node_id` (string, optional), `target_node_id` (string, optional), `relationship_type` (string, optional), `pagination` (object, optional) |

## Examples

### Complete Workflow Example

```bash
# 1. Create a tenant
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {
      "slug": "acme-corp",
      "name": "Acme Corporation"
    },
    "id": 1
  }'

# Response:
# {
#   "jsonrpc": "2.0",
#   "result": {
#     "tenant": {
#       "id": "550e8400-e29b-41d4-a716-446655440000",
#       "slug": "acme-corp",
#       "name": "Acme Corporation",
#       "status": "active",
#       "created_at": "2024-01-01T12:00:00Z",
#       "updated_at": "2024-01-01T12:00:00Z"
#     }
#   },
#   "id": 1
# }

# 2. Create a node type
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_node_type",
    "params": {
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Task",
      "description": "A task node type",
      "schema": "{\"type\": \"object\", \"properties\": {\"title\": {\"type\": \"string\"}}}"
    },
    "id": 2
  }'

# 3. Create a node
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_node",
    "params": {
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "node_type_id": "660e8400-e29b-41d4-a716-446655440001",
      "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
    },
    "id": 3
  }'

# 4. List nodes with pagination
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_nodes",
    "params": {
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "pagination": {
        "page_size": 10,
        "page_token": ""
      }
    },
    "id": 4
  }'
```

### Batch Requests

JSON-RPC 2.0 supports batch requests (multiple method calls in a single HTTP request):

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '[
    {
      "jsonrpc": "2.0",
      "method": "get_tenant",
      "params": {"id": "550e8400-e29b-41d4-a716-446655440000"},
      "id": 1
    },
    {
      "jsonrpc": "2.0",
      "method": "list_nodes",
      "params": {
        "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
        "pagination": {"page_size": 10}
      },
      "id": 2
    }
  ]'
```

### Notifications (Fire-and-Forget)

Use `"id": null` for notifications (no response expected):

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "update_tenant",
    "params": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "inactive"
    },
    "id": null
  }'
```

## Best Practices

### 1. Error Handling

Always check for errors in the response:

```python
response = client._call("get_tenant", {"id": tenant_id})
if "error" in response:
    error = response["error"]
    if error["code"] == -32001:
        # Handle not found
        print("Tenant not found")
    else:
        # Handle other errors
        print(f"Error: {error['message']}")
```

### 2. Request IDs

Use sequential or UUID-based request IDs to track requests:

```python
import uuid

request_id = str(uuid.uuid4())
```

### 3. Timeouts

Set appropriate HTTP timeouts:

```python
import requests

response = requests.post(
    endpoint,
    json=payload,
    timeout=30  # 30 seconds
)
```

### 4. Retry Logic

Implement retry logic for transient failures:

```python
import time
from requests.exceptions import RequestException

def call_with_retry(method, params, max_retries=3):
    for attempt in range(max_retries):
        try:
            return client._call(method, params)
        except RequestException as e:
            if attempt < max_retries - 1:
                time.sleep(2 ** attempt)  # Exponential backoff
                continue
            raise
```

### 5. Connection Pooling

Reuse HTTP connections for better performance:

```python
import requests

session = requests.Session()
# Reuse session for multiple requests
```

### 6. Pagination

Handle pagination properly:

```python
def list_all_tenants(client):
    all_tenants = []
    page_token = ""
    
    while True:
        result = client.list_tenants(page_size=100, page_token=page_token)
        tenants = result["tenants"]
        all_tenants.extend(tenants)
        
        pagination = result["pagination"]
        page_token = pagination.get("next_page_token", "")
        
        if not page_token:
            break
    
    return all_tenants
```

### 7. Data Validation

Validate JSON data before sending:

```python
import json

def create_node_safe(client, tenant_id, node_type_id, data_dict):
    # Validate and serialize data
    try:
        data_json = json.dumps(data_dict)
        return client.create_node(tenant_id, node_type_id, data_json)
    except json.JSONDecodeError:
        raise ValueError("Invalid JSON data")
```

## Testing

### Using curl

```bash
# Test the endpoint
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tenants",
    "params": {},
    "id": 1
  }'
```

### Using Python requests

```python
import requests

response = requests.post(
    "http://localhost:5000/jsonrpc",
    json={
        "jsonrpc": "2.0",
        "method": "list_tenants",
        "params": {},
        "id": 1
    }
)
print(response.json())
```

## Support

For issues or questions:
- Check the [main README](../README.md) for setup instructions
- Review the [REST API documentation](../README.md#api-usage) for comparison
- Open an issue on GitHub

