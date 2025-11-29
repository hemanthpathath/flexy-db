# flex-db

A **Database-as-a-Service (DBaaS)** built with Go, gRPC, and PostgreSQL. This service provides a flexible, multi-tenant data storage solution with five core primitives: Tenant, User, NodeType, Node, and Relationship.

## Features

- **Multi-tenant architecture**: All Nodes and Relationships are scoped to a Tenant
- **Flexible data model**: Generic NodeTypes and Nodes with JSONB data storage
- **Graph-like relationships**: Connect Nodes with typed Relationships
- **gRPC API**: Full CRUD operations with pagination support
- **PostgreSQL backend**: Robust, production-ready database

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        gRPC API                             │
│  (TenantService, UserService, NodeTypeService,              │
│   NodeService, RelationshipService)                         │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                           │
│  (Business logic, validation)                               │
├─────────────────────────────────────────────────────────────┤
│                   Repository Layer                          │
│  (PostgreSQL implementations)                               │
├─────────────────────────────────────────────────────────────┤
│                      PostgreSQL                             │
│  (tenants, users, tenant_users, node_types, nodes,         │
│   relationships)                                            │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
flex-db/
├── go/                         # Go implementation
│   ├── api/proto/              # gRPC protobuf definitions
│   │   ├── dbaas.proto
│   │   ├── dbaas.pb.go         # Generated Go code
│   │   └── dbaas_grpc.pb.go    # Generated gRPC code
│   ├── cmd/dbaas-server/       # Main server entry point
│   │   └── main.go
│   ├── docs/                   # Documentation and guides
│   │   ├── SETUP.md            # Local development setup guide
│   │   └── INSOMNIA_GUIDE.md   # Insomnia gRPC testing guide
│   ├── internal/
│   │   ├── db/                 # Database connection and migrations
│   │   │   ├── db.go
│   │   │   └── migrations/     # SQL migration files
│   │   ├── repository/         # Data access layer
│   │   ├── service/            # Business logic layer
│   │   └── grpc/               # gRPC handlers
│   ├── integration/            # Integration tests
│   ├── scripts/                # Utility scripts
│   │   ├── start.sh            # Quick start script
│   │   ├── load-env.sh         # Environment variable loader
│   │   └── regenerate-proto.sh # Regenerate protobuf files
│   ├── go.mod
│   └── go.sum
├── python/                     # Python implementation with JSON-RPC
│   ├── app/                    # Application code
│   │   ├── db/                 # Database connection and migrations
│   │   ├── repository/         # Data access layer
│   │   ├── service/            # Business logic layer
│   │   └── jsonrpc/            # JSON-RPC handlers
│   ├── scripts/                # Utility scripts
│   ├── main.py                 # Main entry point
│   └── requirements.txt        # Python dependencies
├── docker-compose.yml
└── README.md
```

## Prerequisites

### Go Backend
- Go 1.21+
- PostgreSQL 14+
- protoc (Protocol Buffers compiler) - only needed for regenerating proto files

### Python Backend
- Python 3.9+
- PostgreSQL 14+

## Quick Start

### Go Backend (gRPC)

```bash
# 1. Start PostgreSQL
docker-compose up -d

# 2. Set up environment variables
cp .env.example .env.local

# 3. Run the server (handles everything automatically)
cd go && ./scripts/start.sh
```

### Python Backend (JSON-RPC)

```bash
# 1. Start PostgreSQL
docker-compose up -d

# 2. Set up environment variables
cd python && cp .env.example .env.local

# 3. Run the server (handles everything automatically)
./scripts/start.sh
```

**📚 For Go setup instructions, see [go/docs/SETUP.md](go/docs/SETUP.md)**

**📚 For Python setup instructions, see [python/README.md](python/README.md)**

**🧪 For testing Go APIs with Insomnia, see [go/docs/INSOMNIA_GUIDE.md](go/docs/INSOMNIA_GUIDE.md)**

## Documentation

- **[Go Setup Guide](go/docs/SETUP.md)** - Complete Go backend setup instructions
- **[Python README](python/README.md)** - Complete Python backend setup instructions
- **[Insomnia Testing Guide](go/docs/INSOMNIA_GUIDE.md)** - Step-by-step guide for testing gRPC APIs with Insomnia

## API Usage

### Using Insomnia (Recommended)

See the [Insomnia Testing Guide](go/docs/INSOMNIA_GUIDE.md) for detailed instructions on how to set up and test gRPC requests.

### Using grpcurl

Install [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

#### List Available Services

```bash
grpcurl -plaintext localhost:50051 list
```

#### Create a Tenant

```bash
grpcurl -plaintext -d '{"slug": "acme-corp", "name": "Acme Corporation"}' \
  localhost:50051 dbaas.TenantService/CreateTenant
```

#### Get a Tenant

```bash
grpcurl -plaintext -d '{"id": "TENANT_ID"}' \
  localhost:50051 dbaas.TenantService/GetTenant
```

#### List Tenants

```bash
grpcurl -plaintext -d '{"pagination": {"page_size": 10}}' \
  localhost:50051 dbaas.TenantService/ListTenants
```

#### Create a User

```bash
grpcurl -plaintext -d '{"email": "john@example.com", "display_name": "John Doe"}' \
  localhost:50051 dbaas.UserService/CreateUser
```

#### Add User to Tenant

```bash
grpcurl -plaintext -d '{"tenant_id": "TENANT_ID", "user_id": "USER_ID", "role": "admin"}' \
  localhost:50051 dbaas.UserService/AddUserToTenant
```

#### Create a NodeType

```bash
grpcurl -plaintext -d '{
  "tenant_id": "TENANT_ID",
  "name": "Task",
  "description": "A task node type",
  "schema": "{\"type\": \"object\", \"properties\": {\"title\": {\"type\": \"string\"}}}"
}' localhost:50051 dbaas.NodeTypeService/CreateNodeType
```

#### Create a Node

```bash
grpcurl -plaintext -d '{
  "tenant_id": "TENANT_ID",
  "node_type_id": "NODE_TYPE_ID",
  "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
}' localhost:50051 dbaas.NodeService/CreateNode
```

#### Create a Relationship

```bash
grpcurl -plaintext -d '{
  "tenant_id": "TENANT_ID",
  "source_node_id": "SOURCE_NODE_ID",
  "target_node_id": "TARGET_NODE_ID",
  "relationship_type": "depends_on",
  "data": "{\"priority\": 1}"
}' localhost:50051 dbaas.RelationshipService/CreateRelationship
```

### Using evans (Interactive gRPC Client)

Install [evans](https://github.com/ktr0731/evans):

```bash
# macOS
brew install evans

# Linux
go install github.com/ktr0731/evans@latest
```

Connect to the server:

```bash
evans --host localhost --port 50051 -r repl
```

## Data Model

### Tenant
- Primary entity for multi-tenancy
- Contains: id, slug (unique), name, status, timestamps

### User
- Global user entity
- Can be associated with multiple tenants via tenant_users
- Contains: id, email (unique), display_name, timestamps

### NodeType
- Defines the schema for nodes within a tenant
- Contains: id, tenant_id, name, description, schema (JSON), timestamps

### Node
- Actual data entities
- Scoped to tenant and node type
- Contains: id, tenant_id, node_type_id, data (JSONB), timestamps

### Relationship
- Connects two nodes
- Contains: id, tenant_id, source_node_id, target_node_id, relationship_type, data (JSONB), timestamps

## Database Migrations

Migrations are embedded in the binary and run automatically on server startup. The migrations create:

1. `tenants` - Tenant records
2. `users` - User records  
3. `tenant_users` - User-tenant membership
4. `node_types` - Node type definitions
5. `nodes` - Node instances
6. `relationships` - Node relationships

## Development

### Regenerate Protobuf Code

If you modify the protobuf definitions:

```bash
# Use the regenerate script (recommended)
cd go && ./scripts/regenerate-proto.sh

# Or manually
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
cd go && protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/dbaas.proto
```

### Build

```bash
cd go && go build -o dbaas-server ./cmd/dbaas-server
```

### Run Tests

```bash
cd go && go test ./...
```

## License

MIT License
