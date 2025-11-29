# flex-db

A **Database-as-a-Service (DBaaS)** with multiple implementation options. This service provides a flexible, multi-tenant data storage solution with five core primitives: Tenant, User, NodeType, Node, and Relationship.

## Features

- **Multi-tenant architecture**: All Nodes and Relationships are scoped to a Tenant
- **Flexible data model**: Generic NodeTypes and Nodes with JSONB data storage
- **Graph-like relationships**: Connect Nodes with typed Relationships
- **Multiple API options**: Choose between gRPC (Go) or JSON-RPC/REST (Python)
- **PostgreSQL backend**: Robust, production-ready database
- **Feature parity**: Both implementations provide identical functionality

## Implementations

This project provides two complete implementations that share the same database schema:

1. **Go Backend** (`/go`): gRPC API built with Go
   - Protocol: gRPC (Protocol Buffers)
   - Port: 50051 (default)
   - Database: PostgreSQL with pgx driver

2. **Python Backend** (`/python`): Unified FastAPI service with JSON-RPC and REST
   - Protocols: JSON-RPC 2.0 and REST API
   - Port: 5000 (default) - both protocols on the same port
   - Framework: FastAPI with uvicorn
   - Database: PostgreSQL with asyncpg driver
   - Features: Swagger UI, ReDoc, OpenAPI documentation
   - Architecture: REST endpoints call services directly (no JSON-RPC overhead)

Both implementations provide full CRUD operations with pagination support and share the same database schema.

## Architecture

### Go Backend (gRPC)

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
│  (PostgreSQL implementations with pgx)                     │
├─────────────────────────────────────────────────────────────┤
│                      PostgreSQL                             │
│  (tenants, users, tenant_users, node_types, nodes,         │
│   relationships)                                            │
└─────────────────────────────────────────────────────────────┘
```

### Python Backend (Unified FastAPI Service)

```
┌─────────────────────────────────────────────────────────────┐
│              Unified FastAPI Service (Port 5000)            │
│  ┌──────────────────────┐  ┌──────────────────────┐         │
│  │   REST API          │  │   JSON-RPC API       │         │
│  │   • /tenants       │  │   • /jsonrpc        │         │
│  │   • /users         │  │                      │         │
│  │   • /nodes         │  │                      │         │
│  │   • /relationships │  │                      │         │
│  │   • /docs (Swagger)│  │                      │         │
│  │   • /redoc         │  │                      │         │
│  └──────────┬─────────┘  └──────────┬──────────┘         │
│             │                       │                      │
│             └───────────┬───────────┘                      │
│                         │                                  │
├─────────────────────────┼──────────────────────────────────┤
│                     Service Layer                           │
│  (TenantService, UserService, NodeTypeService,              │
│   NodeService, RelationshipService)                        │
├─────────────────────────────────────────────────────────────┤
│                   Repository Layer                          │
│  (PostgreSQL implementations with asyncpg)                  │
├─────────────────────────────────────────────────────────────┤
│                      PostgreSQL                             │
│  (tenants, users, tenant_users, node_types, nodes,         │
│   relationships)                                            │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
flex-db/
├── go/                         # Go implementation (gRPC)
│   ├── api/proto/              # gRPC protobuf definitions
│   │   ├── dbaas.proto
│   │   ├── dbaas.pb.go         # Generated Go code
│   │   └── dbaas_grpc.pb.go    # Generated gRPC code
│   ├── cmd/dbaas-server/       # Main server entry point
│   │   └── main.go
│   ├── docs/                   # Documentation and guides
│   │   ├── SETUP.md            # Local development setup guide
│   │   ├── INSOMNIA_GUIDE.md   # Insomnia gRPC testing guide
│   │   ├── CLIENT_INTEGRATION.md
│   │   └── README.md
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
├── python/                     # Python implementation (JSON-RPC + REST)
│   ├── app/                    # Application code
│   │   ├── config.py           # Configuration management
│   │   ├── db/                 # Database connection and migrations
│   │   ├── repository/         # Data access layer
│   │   ├── service/            # Business logic layer
│   │   ├── jsonrpc/            # JSON-RPC handlers
│   │   └── api/                # REST API (unified with main service)
│   │       ├── routers/        # REST endpoint routers
│   │       ├── models.py       # Pydantic request/response models
│   │       └── errors.py       # Error handling utilities
│   ├── scripts/                # Utility scripts
│   ├── main.py                 # Main entry point (FastAPI - JSON-RPC & REST)
│   ├── requirements.txt        # Python dependencies
│   ├── .env.example            # Environment variable template
│   ├── Dockerfile              # Docker image
│   └── docker-compose.yml      # Docker Compose config
├── docker-compose.yml          # Shared PostgreSQL setup
└── README.md                   # This file
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
cd go && cp .env.example .env.local

# 3. Run the server (handles everything automatically)
./scripts/start.sh
```

The server will start on `localhost:50051` (gRPC).

**📚 For detailed Go setup instructions, see [go/docs/SETUP.md](go/docs/SETUP.md)**

**🧪 For testing Go APIs with Insomnia, see [go/docs/INSOMNIA_GUIDE.md](go/docs/INSOMNIA_GUIDE.md)**

### Python Backend (JSON-RPC + REST)

```bash
# 1. Start PostgreSQL
docker-compose up -d

# 2. Set up environment variables
cd python && cp .env.example .env.local

# 3. Run the unified server (handles everything automatically)
./scripts/start.sh
# Or: python main.py
```

The unified service will start on `localhost:5000` with both JSON-RPC and REST APIs:
- **JSON-RPC**: `http://localhost:5000/jsonrpc`
- **REST API**: `http://localhost:5000/tenants`, `/users`, `/nodes`, etc.
- **Swagger UI**: `http://localhost:5000/docs`
- **ReDoc**: `http://localhost:5000/redoc`
- **Health Check**: `http://localhost:5000/health`

**📚 For detailed Python setup instructions, see [python/README.md](python/README.md)**

## Documentation

### Go Backend
- **[Go Setup Guide](go/docs/SETUP.md)** - Complete Go backend setup instructions
- **[Insomnia Testing Guide](go/docs/INSOMNIA_GUIDE.md)** - Step-by-step guide for testing gRPC APIs with Insomnia
- **[Client Integration Guide](go/docs/CLIENT_INTEGRATION.md)** - Guide for integrating Go clients

### Python Backend
- **[Python README](python/README.md)** - Complete Python backend setup instructions

## API Usage

### Go Backend (gRPC)

#### Using Insomnia (Recommended)

See the [Insomnia Testing Guide](go/docs/INSOMNIA_GUIDE.md) for detailed instructions on how to set up and test gRPC requests.

#### Using grpcurl

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

### Python Backend (JSON-RPC + REST)

The Python backend provides both JSON-RPC and REST APIs on the same port (5000).

#### JSON-RPC API

The JSON-RPC 2.0 API is available at `http://localhost:5000/jsonrpc`:

```bash
# Create a tenant
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {"slug": "acme-corp", "name": "Acme Corporation"},
    "id": 1
  }'
```

#### REST API

The REST API provides RESTful endpoints that call services directly:

```bash
# Create a tenant
curl -X POST http://localhost:5000/tenants \
  -H "Content-Type: application/json" \
  -d '{"slug": "acme-corp", "name": "Acme Corporation"}'

# Get a tenant
curl http://localhost:5000/tenants/{tenant_id}

# List tenants
curl "http://localhost:5000/tenants?page_size=10"
```

#### Interactive API Documentation

Both JSON-RPC and REST are documented in the unified OpenAPI specification:
- **Swagger UI**: http://localhost:5000/docs
- **ReDoc**: http://localhost:5000/redoc
- **OpenAPI Schema**: http://localhost:5000/openapi.json

For more examples and the complete API reference, see [python/README.md](python/README.md).

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

Both implementations use the same database schema. Migrations run automatically on server startup:

1. `tenants` - Tenant records
2. `users` - User records  
3. `tenant_users` - User-tenant membership
4. `node_types` - Node type definitions
5. `nodes` - Node instances
6. `relationships` - Node relationships

**Note**: Both Go and Python implementations can share the same database. The migrations are identical and idempotent.

## Development

### Go Backend

#### Regenerate Protobuf Code

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

#### Build

```bash
cd go && go build -o dbaas-server ./cmd/dbaas-server
```

#### Run Tests

```bash
cd go && go test ./...
```

### Python Backend

#### Run Tests

```bash
cd python && python -m pytest
```

#### Development Mode

```bash
# Unified service with auto-reload (both JSON-RPC and REST)
cd python && RELOAD=true python main.py

# Or using uvicorn directly
cd python && uvicorn main:app --reload --host 0.0.0.0 --port 5000
```

### Running Both Implementations

You can run both Go and Python backends simultaneously as they use the same database schema:

```bash
# Terminal 1: Start PostgreSQL
docker-compose up -d

# Terminal 2: Start Go backend (gRPC on port 50051)
cd go && ./scripts/start.sh

# Terminal 3: Start Python backend (JSON-RPC + REST on port 5000)
cd python && ./scripts/start.sh
```

This allows you to test and compare both implementations side-by-side:
- **Go (gRPC)**: `localhost:50051`
- **Python (JSON-RPC)**: `localhost:5000/jsonrpc`
- **Python (REST)**: `localhost:5000/tenants`, `/users`, etc.
- **Python (Docs)**: `localhost:5000/docs`

## License

MIT License
