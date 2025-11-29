# Insomnia gRPC Testing Guide

Quick reference for testing flex-db APIs with Insomnia.

> **Note:** All commands in this guide should be run from the project root directory (`flex-db/`).

## Initial Setup

### 1. Create gRPC Request
- Click **"Send a request"** or the **+** button
- Select **"gRPC Request"** as the request type

### 2. Configure Server URL
- **URL:** `localhost:50051`
- (Or the port specified in your `.env.local` file)

### 3. Import Proto File
- Click **"Select Proto File"** or **"Use Proto File"**
- Navigate to: `go/api/proto/dbaas.proto`
- Select the file

After importing, you'll see all available services in the dropdown.

## Available Services

### TenantService
Manage tenants (organizations).

### UserService
Manage users and their tenant associations.

### NodeTypeService
Define node schemas for tenants.

### NodeService
Create and manage data nodes.

### RelationshipService
Create relationships between nodes.

## Example Requests

### Create a Tenant

**Service:** `TenantService.CreateTenant`

**Request Body:**
```json
{
  "slug": "acme-corp",
  "name": "Acme Corporation"
}
```

**Response:** Returns the created tenant with ID, timestamps, etc.

---

### Get a Tenant

**Service:** `TenantService.GetTenant`

**Request Body:**
```json
{
  "id": "TENANT_ID_FROM_CREATE_RESPONSE"
}
```

---

### List Tenants

**Service:** `TenantService.ListTenants`

**Request Body:**
```json
{
  "pagination": {
    "page_size": 10
  }
}
```

---

### Create a User

**Service:** `UserService.CreateUser`

**Request Body:**
```json
{
  "email": "john@example.com",
  "display_name": "John Doe"
}
```

---

### Add User to Tenant

**Service:** `UserService.AddUserToTenant`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "user_id": "USER_ID",
  "role": "admin"
}
```

**Note:** Save the IDs from previous responses to use here.

---

### Create a NodeType

**Service:** `NodeTypeService.CreateNodeType`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "name": "Task",
  "description": "A task node type",
  "schema": "{\"type\": \"object\", \"properties\": {\"title\": {\"type\": \"string\"}}}"
}
```

---

### Create a Node

**Service:** `NodeService.CreateNode`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "node_type_id": "NODE_TYPE_ID",
  "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
}
```

---

### Create a Relationship

**Service:** `RelationshipService.CreateRelationship`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "source_node_id": "SOURCE_NODE_ID",
  "target_node_id": "TARGET_NODE_ID",
  "relationship_type": "depends_on",
  "data": "{\"priority\": 1}"
}
```

## Testing Workflow

1. **Start the server:**
   ```bash
   cd go && ./scripts/start.sh
   ```

2. **Create a Tenant** - Get your first tenant ID

3. **Create a User** - Get a user ID

4. **Add User to Tenant** - Associate user with tenant

5. **Create a NodeType** - Define a schema for your data

6. **Create Nodes** - Add data instances

7. **Create Relationships** - Connect nodes together

## Troubleshooting

### Connection Issues

- Make sure the server is running (`cd go && ./scripts/start.sh`)
- Verify the port matches (default is 50051)
- Check the server logs for errors

### Proto File Not Loading

- Ensure `go/api/proto/dbaas.proto` exists
- Make sure the proto file is properly formatted
- Try regenerating: `cd go && ./scripts/regenerate-proto.sh`

### Methods Not Showing

- The server has gRPC reflection enabled, so methods should auto-discover
- Try re-importing the proto file
- Restart Insomnia if needed

