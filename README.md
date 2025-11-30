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
- **FastAPI framework**: High-performance async Python web framework

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client                               │
│              (Any JSON-RPC 2.0 compatible client)           │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                      JSON-RPC API                           │
│  (FastAPI + jsonrpcserver)                                  │
│                                                             │
│  Endpoints:                                                 │
│  • POST /jsonrpc      - JSON-RPC 2.0 endpoint              │
│  • GET  /openrpc.json - OpenRPC specification              │
│  • GET  /health       - Health check endpoint              │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                           │
│  • TenantService      - Tenant management                   │
│  • UserService        - User management                     │
│  • NodeTypeService    - Schema definitions                  │
│  • NodeService        - Data entity CRUD                    │
│  • RelationshipService- Node connections                    │
├─────────────────────────────────────────────────────────────┤
│                   Repository Layer                          │
│  (PostgreSQL implementations with asyncpg)                  │
│  • Async database operations                                │
│  • Connection pooling                                       │
│  • Transaction management                                   │
├─────────────────────────────────────────────────────────────┤
│                      PostgreSQL 14                          │
│  Tables: tenants, users, tenant_users, node_types,          │
│          nodes, relationships                               │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
flex-db/
├── app/                        # Application code
│   ├── __init__.py
│   ├── config.py               # Configuration management
│   ├── api/                    # API dependencies and models
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
│   ├── setup_local.sh          # Local environment setup
│   ├── start.sh                # Start the server
│   └── test_basic_operations.sh# Basic API tests
├── main.py                     # Main entry point
├── requirements.txt            # Python dependencies
├── .env.example                # Environment variable template
├── Dockerfile                  # Docker image definition
├── docker-compose.yml          # Docker Compose configuration
├── Makefile                    # Development commands
└── README.md                   # This file
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Python 3.11 |
| Web Framework | FastAPI |
| RPC Protocol | JSON-RPC 2.0 (jsonrpcserver) |
| Database | PostgreSQL 14 |
| DB Driver | asyncpg (async) |
| Server | Uvicorn |
| Containerization | Docker & Docker Compose |

## Prerequisites

- **Docker** and **Docker Compose** (recommended for development)
- **Python 3.9+** (for local development without Docker)
- **PostgreSQL 14+** (if running locally without Docker)

## Quick Start

### Option 1: Docker (Recommended)

The fastest way to get started is using Docker:

```bash
# Clone the repository
git clone https://github.com/hemanthpathath/flex-db.git
cd flex-db

# Start the development environment
make setup-dev
```

Once started, the following services are available:

| Service | URL |
|---------|-----|
| JSON-RPC API | http://localhost:5000/jsonrpc |
| OpenRPC Spec | http://localhost:5000/openrpc.json |
| Health Check | http://localhost:5000/health |
| PostgreSQL | localhost:5432 |

### Option 2: Local Development

For local development without Docker:

```bash
# 1. Start PostgreSQL (using Docker)
docker compose --profile dev up postgres -d

# 2. Create environment file
cp .env.example .env.local

# 3. Create and activate virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# 4. Install dependencies
pip install -r requirements.txt

# 5. Run the server
python main.py
```

## Development Workflow

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make setup-dev` | Build and start development environment |
| `make test-all` | Run all tests in isolation |
| `make stop` | Stop all running containers |
| `make clean` | Remove containers, volumes, and images |
| `make build` | Build Docker images |
| `make logs` | Show logs from running containers |
| `make status` | Show container status |

### Development Mode with Auto-Reload

```bash
# Using environment variable
RELOAD=true python main.py

# Or using uvicorn directly
uvicorn main:app --reload --host 0.0.0.0 --port 5000
```

### Running Tests

```bash
# Run tests in Docker (recommended)
make test-all

# Run tests locally
python -m pytest
```

## API Usage

### JSON-RPC 2.0 Endpoint

All API operations use the JSON-RPC 2.0 protocol at `POST /jsonrpc`.

#### Create a Tenant

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

#### Create a User

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_user",
    "params": {"email": "user@example.com", "display_name": "John Doe"},
    "id": 2
  }'
```

#### Create a NodeType

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_node_type",
    "params": {
      "tenant_id": "TENANT_ID",
      "name": "Article",
      "description": "Blog article",
      "schema": {"title": "string", "content": "string"}
    },
    "id": 3
  }'
```

#### Create a Node

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_node",
    "params": {
      "tenant_id": "TENANT_ID",
      "node_type_id": "NODE_TYPE_ID",
      "data": {"title": "Hello World", "content": "My first article"}
    },
    "id": 4
  }'
```

#### Create a Relationship

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_relationship",
    "params": {
      "tenant_id": "TENANT_ID",
      "source_node_id": "SOURCE_ID",
      "target_node_id": "TARGET_ID",
      "relationship_type": "references"
    },
    "id": 5
  }'
```

### Available Methods

| Category | Methods |
|----------|---------|
| Tenant | `create_tenant`, `get_tenant`, `list_tenants`, `update_tenant`, `delete_tenant` |
| User | `create_user`, `get_user`, `list_users`, `update_user`, `delete_user`, `add_user_to_tenant` |
| NodeType | `create_node_type`, `get_node_type`, `list_node_types`, `update_node_type`, `delete_node_type` |
| Node | `create_node`, `get_node`, `list_nodes`, `update_node`, `delete_node` |
| Relationship | `create_relationship`, `get_relationship`, `list_relationships`, `delete_relationship` |

For complete API documentation, see the [OpenRPC specification](http://localhost:5000/openrpc.json) or the [JSON-RPC Integration Guide](docs/JSON_RPC_INTEGRATION.md).

## Data Model

### Entity Relationship Diagram

```
┌─────────────┐       ┌──────────────┐       ┌─────────────┐
│   Tenant    │───────│ tenant_users │───────│    User     │
│             │  1:N  │              │  N:1  │             │
│ - id        │       │ - tenant_id  │       │ - id        │
│ - slug      │       │ - user_id    │       │ - email     │
│ - name      │       │ - role       │       │ - display   │
│ - status    │       └──────────────┘       │   _name     │
└──────┬──────┘                              └─────────────┘
       │ 1:N
       │
┌──────▼──────┐       ┌──────────────────┐
│  NodeType   │       │   Relationship   │
│             │       │                  │
│ - id        │       │ - id             │
│ - tenant_id │       │ - tenant_id      │
│ - name      │       │ - source_node_id │
│ - schema    │       │ - target_node_id │
└──────┬──────┘       │ - type           │
       │ 1:N          │ - data           │
       │              └────────▲─────────┘
┌──────▼──────┐                │
│    Node     │────────────────┘
│             │  N:N (via Relationship)
│ - id        │
│ - tenant_id │
│ - type_id   │
│ - data      │
└─────────────┘
```

### Entities

| Entity | Description |
|--------|-------------|
| **Tenant** | Organization/workspace that owns data. All nodes and relationships are tenant-scoped. |
| **User** | Global user that can belong to multiple tenants with different roles. |
| **NodeType** | Schema definition for nodes within a tenant (e.g., "Article", "Comment"). |
| **Node** | Actual data entity with JSONB data, conforming to a NodeType schema. |
| **Relationship** | Typed connection between two nodes with optional JSONB metadata. |

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `dbaas` |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `JSONRPC_HOST` | Server host | `0.0.0.0` |
| `JSONRPC_PORT` | Server port | `5000` |
| `RELOAD` | Enable auto-reload | `false` |

## Database Migrations

Migrations run automatically on server startup. The following tables are created:

1. `tenants` - Tenant records
2. `users` - User records  
3. `tenant_users` - User-tenant membership with roles
4. `node_types` - Node type/schema definitions
5. `nodes` - Node instances with JSONB data
6. `relationships` - Node relationships with JSONB metadata

## Documentation

| Document | Description |
|----------|-------------|
| [Local Setup Guide](docs/LOCAL_SETUP.md) | Detailed local development setup |
| [JSON-RPC Integration](docs/JSON_RPC_INTEGRATION.md) | Complete API reference and examples |
| [Database Architecture](docs/DATABASE_ARCHITECTURE.md) | Database schema and design decisions |

## Docker Configuration

### Development Profile

```bash
# Start development environment
docker compose --profile dev up -d

# Services started:
# - postgres (PostgreSQL 14)
# - flex-db (Python application)
```

### Test Profile

```bash
# Run tests in isolation
docker compose --profile test up --abort-on-container-exit

# Services started:
# - postgres-test (PostgreSQL with tmpfs for speed)
# - flex-db-test (Test instance)
# - test-runner (Executes tests)
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.
