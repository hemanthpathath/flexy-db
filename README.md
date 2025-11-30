# flex-db

A **Database-as-a-Service (DBaaS)** implemented in Python with a JSON-RPC API. This service provides a flexible, multi-tenant data storage solution with five core primitives: Tenant, User, NodeType, Node, and Relationship.

## Features

- **Multi-tenant architecture**: All Nodes and Relationships are scoped to a Tenant
- **Flexible data model**: Generic NodeTypes and Nodes with JSONB data storage
- **Graph-like relationships**: Connect Nodes with typed Relationships
- **JSON-RPC 2.0 API**: Standards-compliant RPC protocol
- **OpenRPC specification**: Machine-readable API documentation
- **PostgreSQL backend**: Robust, production-ready database
- **Docker support**: Development and testing workflows via Docker Compose

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      JSON-RPC API                           │
│  (TenantService, UserService, NodeTypeService,              │
│   NodeService, RelationshipService)                         │
│  • /jsonrpc - JSON-RPC 2.0 endpoint                        │
│  • /openrpc.json - OpenRPC specification                    │
│  • /health - Health check endpoint                         │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                           │
│  (Business logic, validation)                               │
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
├── app/                        # Application code
│   ├── __init__.py
│   ├── api/                    # API dependencies
│   ├── config.py               # Configuration management
│   ├── db/                     # Database connection and migrations
│   ├── jsonrpc/                # JSON-RPC handlers and OpenRPC
│   │   ├── handlers.py         # JSON-RPC method handlers
│   │   ├── server.py           # FastAPI router for JSON-RPC
│   │   └── openrpc.py          # OpenRPC specification generator
│   ├── repository/             # Data access layer
│   └── service/                # Business logic layer
├── docs/                       # Documentation
│   ├── DATABASE_ARCHITECTURE.md
│   ├── JSON_RPC_INTEGRATION.md
│   └── LOCAL_SETUP.md
├── scripts/                    # Utility scripts
│   ├── setup_local.sh
│   ├── start.sh
│   └── test_basic_operations.sh
├── main.py                     # Main entry point (FastAPI - JSON-RPC)
├── requirements.txt            # Python dependencies
├── .env.example                # Environment variable template
├── Dockerfile                  # Docker image
├── docker-compose.yml          # Docker Compose config
├── Makefile                    # Development and testing commands
└── README.md                   # This file
```

## Prerequisites

- Python 3.9+ (for local development)
- Docker and Docker Compose (for containerized development)
- PostgreSQL 14+ (if running locally without Docker)

## Quick Start

### Using Docker (Recommended)

```bash
# Start development environment
make setup-dev

# The service will be available at:
# - JSON-RPC API: http://localhost:5000/jsonrpc
# - OpenRPC Spec: http://localhost:5000/openrpc.json
# - Health Check: http://localhost:5000/health

# Run tests
make test-all

# Stop containers
make stop

# Clean up
make clean
```

### Local Development

```bash
# 1. Start PostgreSQL
docker compose --profile dev up postgres -d

# 2. Set up environment variables
cp .env.example .env.local

# 3. Install dependencies
pip install -r requirements.txt

# 4. Run the server
python main.py
```

The service will start on `localhost:5000` with JSON-RPC API:
- **JSON-RPC**: `http://localhost:5000/jsonrpc`
- **OpenRPC Spec**: `http://localhost:5000/openrpc.json`
- **Health Check**: `http://localhost:5000/health`

## Documentation

- **[Local Setup Guide](docs/LOCAL_SETUP.md)** - Complete local development setup
- **[JSON-RPC Integration Guide](docs/JSON_RPC_INTEGRATION.md)** - Comprehensive API integration guide
- **[Database Architecture](docs/DATABASE_ARCHITECTURE.md)** - Database schema and design

## API Usage (JSON-RPC)

The API provides a JSON-RPC 2.0 endpoint at `http://localhost:5000/jsonrpc`.

### Create a Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {"slug": "acme-corp", "name": "Acme Corporation"},
    "id": 1
  }'
```

### Get a Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "get_tenant",
    "params": {"id": "TENANT_ID"},
    "id": 2
  }'
```

### List Tenants

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tenants",
    "params": {"pagination": {"page_size": 10}},
    "id": 3
  }'
```

### OpenRPC Specification

The service provides an OpenRPC specification (similar to OpenAPI for REST):
- **OpenRPC Spec**: http://localhost:5000/openrpc.json
- **Integration Guide**: [docs/JSON_RPC_INTEGRATION.md](docs/JSON_RPC_INTEGRATION.md)

For complete API reference, client implementations, and examples, see [docs/JSON_RPC_INTEGRATION.md](docs/JSON_RPC_INTEGRATION.md).

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

Migrations run automatically on server startup:

1. `tenants` - Tenant records
2. `users` - User records  
3. `tenant_users` - User-tenant membership
4. `node_types` - Node type definitions
5. `nodes` - Node instances
6. `relationships` - Node relationships

## Development

### Run Tests

```bash
python -m pytest
```

### Development Mode

```bash
# JSON-RPC service with auto-reload
RELOAD=true python main.py

# Or using uvicorn directly
uvicorn main:app --reload --host 0.0.0.0 --port 5000
```

## License

MIT License
