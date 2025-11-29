# flex-db Python Backend

A **Database-as-a-Service (DBaaS)** implemented in Python with JSON-RPC. This is a Python implementation that provides feature parity with the Go backend.

## Features

- **Multi-tenant architecture**: All Nodes and Relationships are scoped to a Tenant
- **Flexible data model**: Generic NodeTypes and Nodes with JSONB data storage
- **Graph-like relationships**: Connect Nodes with typed Relationships
- **JSON-RPC API**: Full CRUD operations with pagination support
- **REST API Wrapper**: FastAPI-based REST facade with Swagger UI documentation
- **PostgreSQL backend**: Robust, production-ready database
- **Async/await**: Built with asyncio for high-performance I/O

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    REST API (FastAPI)                       │
│  Optional: Swagger UI at /docs, ReDoc at /redoc             │
├─────────────────────────────────────────────────────────────┤
│                      JSON-RPC API                           │
│  (TenantService, UserService, NodeTypeService,              │
│   NodeService, RelationshipService)                         │
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
python/
├── app/
│   ├── __init__.py
│   ├── config.py           # Configuration management
│   ├── db/
│   │   ├── __init__.py
│   │   ├── database.py     # Database connection and migrations
│   │   └── migrations/     # SQL migration files
│   ├── repository/
│   │   ├── __init__.py
│   │   ├── errors.py       # Repository errors
│   │   ├── models.py       # Data models
│   │   ├── tenant_repo.py
│   │   ├── user_repo.py
│   │   ├── nodetype_repo.py
│   │   ├── node_repo.py
│   │   └── relationship_repo.py
│   ├── service/
│   │   ├── __init__.py
│   │   ├── tenant_service.py
│   │   ├── user_service.py
│   │   ├── nodetype_service.py
│   │   ├── node_service.py
│   │   └── relationship_service.py
│   └── jsonrpc/
│       ├── __init__.py
│       ├── errors.py       # Error mapping
│       ├── handlers.py     # JSON-RPC method handlers
│       └── server.py       # aiohttp server setup
├── rest_wrapper/           # REST API facade (NEW)
│   ├── __init__.py
│   ├── main.py             # FastAPI application
│   ├── config.py           # REST wrapper configuration
│   ├── client.py           # JSON-RPC client
│   ├── models.py           # Pydantic request/response models
│   ├── routers/            # REST endpoint routers
│   └── tests/              # REST wrapper tests
├── scripts/
│   └── start.sh            # Quick start script
├── main.py                 # Main entry point
├── requirements.txt        # Python dependencies
├── .env.example            # Environment variable template
└── README.md               # This file
```

## Prerequisites

- Python 3.9+
- PostgreSQL 14+ (or Docker for easy setup)

## Quick Start

```bash
# 1. Start PostgreSQL (from repository root)
docker-compose up -d

# 2. Set up environment variables
cp .env.example .env.local

# 3. Run the server (handles everything automatically)
./scripts/start.sh
```

## Docker Setup

### Building the Docker Image

```bash
# From the python directory
docker build -t flex-db-python .
```

### Running with Docker Compose

The easiest way to run the Python backend with Docker is using docker-compose:

```bash
# From the python directory
docker-compose up -d
```

This will start both the PostgreSQL database and the Python backend together.

### Running Standalone (with existing PostgreSQL)

If you already have PostgreSQL running:

```bash
# Run the container connecting to host PostgreSQL
docker run -d \
  --name flex-db-python \
  -p 5000:5000 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=dbaas \
  flex-db-python
```

### Environment Variables

The following environment variables can be configured:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `dbaas` | PostgreSQL database name |
| `DB_SSL_MODE` | `disable` | PostgreSQL SSL mode |
| `JSONRPC_HOST` | `0.0.0.0` | JSON-RPC server host |
| `JSONRPC_PORT` | `5000` | JSON-RPC server port |

### Verifying the Container

```bash
# Check container status
docker ps

# Check health endpoint
curl http://localhost:5000/health

# View logs
docker logs flex-db-python
```

### CI/CD: GitHub Actions Workflow

This repository includes a GitHub Actions workflow that automatically builds and pushes the Docker image to GitHub Container Registry (GHCR).

#### Workflow Triggers

The workflow is triggered on:
- **Push to main branch**: Builds and pushes the image to GHCR
- **Pull requests to main**: Builds the image (without pushing) to validate changes

The workflow only runs when changes are made to:
- Files in the `python/` directory
- The workflow file itself (`.github/workflows/python-docker-build.yml`)

#### Image Tags

The workflow automatically tags images with:
- **Commit SHA**: `ghcr.io/<owner>/<repo>/flex-db:<sha>` - Unique tag for each commit
- **Branch name**: `ghcr.io/<owner>/<repo>/flex-db:main` - For push events
- **PR number**: `ghcr.io/<owner>/<repo>/flex-db:pr-<number>` - For pull request events

#### Pulling the Image from GHCR

To pull the image locally:

```bash
# Authenticate to GHCR (required for private repositories)
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Pull the latest image from main branch
docker pull ghcr.io/<owner>/<repo>/flex-db:main

# Pull a specific commit version
docker pull ghcr.io/<owner>/<repo>/flex-db:<commit-sha>
```

Replace `<owner>/<repo>` with the actual repository path (e.g., `hemanthpathath/flex-db`).

#### Running the GHCR Image

```bash
# Run the container from GHCR
docker run -d \
  --name flex-db-python \
  -p 5000:5000 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=dbaas \
  ghcr.io/<owner>/<repo>/flex-db:main
```

#### Workflow Features

- **Docker Buildx**: Efficient multi-platform builds
- **Layer Caching**: GitHub Actions cache for faster builds
- **Vulnerability Scanning**: Trivy scanner runs on pushed images and uploads results to the GitHub Security tab

## Manual Setup

### Step 1: Set Up PostgreSQL

Using Docker (recommended):

```bash
# From repository root
docker-compose up -d
```

Or using local PostgreSQL:

```bash
createdb dbaas
```

### Step 2: Configure Environment Variables

```bash
cp .env.example .env.local
```

Edit `.env.local` if needed:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dbaas
DB_SSL_MODE=disable

# JSON-RPC Server Configuration
JSONRPC_HOST=0.0.0.0
JSONRPC_PORT=5000
```

### Step 3: Set Up Virtual Environment

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Step 4: Run the Server

```bash
python main.py
```

Expected output:

```
2024-01-01 12:00:00 - INFO - Loading environment from .env.local
2024-01-01 12:00:00 - INFO - Connecting to database...
2024-01-01 12:00:00 - INFO - Connected to database successfully
2024-01-01 12:00:00 - INFO - Running database migrations...
2024-01-01 12:00:00 - INFO - Migrations completed successfully
2024-01-01 12:00:00 - INFO - Starting JSON-RPC server on 0.0.0.0:5000...
2024-01-01 12:00:00 - INFO - JSON-RPC endpoint: http://0.0.0.0:5000/jsonrpc
2024-01-01 12:00:00 - INFO - Health check endpoint: http://0.0.0.0:5000/health
```

## API Usage

The server exposes a JSON-RPC 2.0 API at `http://localhost:5000/jsonrpc`.

### Using curl

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

Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "tenant": {
      "id": "uuid-here",
      "slug": "acme-corp",
      "name": "Acme Corporation",
      "status": "active",
      "created_at": "2024-01-01T12:00:00",
      "updated_at": "2024-01-01T12:00:00"
    }
  },
  "id": 1
}
```

#### Get a Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "get_tenant",
    "params": {"id": "TENANT_ID"},
    "id": 1
  }'
```

#### List Tenants

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tenants",
    "params": {"pagination": {"page_size": 10}},
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
    "params": {"email": "john@example.com", "display_name": "John Doe"},
    "id": 1
  }'
```

#### Add User to Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "add_user_to_tenant",
    "params": {"tenant_id": "TENANT_ID", "user_id": "USER_ID", "role": "admin"},
    "id": 1
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
      "name": "Task",
      "description": "A task node type",
      "schema": "{\"type\": \"object\", \"properties\": {\"title\": {\"type\": \"string\"}}}"
    },
    "id": 1
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
      "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
    },
    "id": 1
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
      "source_node_id": "SOURCE_NODE_ID",
      "target_node_id": "TARGET_NODE_ID",
      "relationship_type": "depends_on",
      "data": "{\"priority\": 1}"
    },
    "id": 1
  }'
```

### Using Python

```python
import requests

def call_jsonrpc(method, params, id=1):
    response = requests.post(
        "http://localhost:5000/jsonrpc",
        json={
            "jsonrpc": "2.0",
            "method": method,
            "params": params,
            "id": id,
        },
    )
    return response.json()

# Create a tenant
result = call_jsonrpc("create_tenant", {"slug": "acme-corp", "name": "Acme Corporation"})
print(result)
```

## API Reference

### Tenant Service

| Method | Parameters | Description |
|--------|-----------|-------------|
| `create_tenant` | `slug`, `name` | Create a new tenant |
| `get_tenant` | `id` | Get tenant by ID |
| `update_tenant` | `id`, `slug?`, `name?`, `status?` | Update tenant |
| `delete_tenant` | `id` | Delete tenant |
| `list_tenants` | `pagination?` | List tenants with pagination |

### User Service

| Method | Parameters | Description |
|--------|-----------|-------------|
| `create_user` | `email`, `display_name` | Create a new user |
| `get_user` | `id` | Get user by ID |
| `update_user` | `id`, `email?`, `display_name?` | Update user |
| `delete_user` | `id` | Delete user |
| `list_users` | `pagination?` | List users with pagination |
| `add_user_to_tenant` | `tenant_id`, `user_id`, `role?` | Add user to tenant |
| `remove_user_from_tenant` | `tenant_id`, `user_id` | Remove user from tenant |
| `list_tenant_users` | `tenant_id`, `pagination?` | List users in tenant |

### NodeType Service

| Method | Parameters | Description |
|--------|-----------|-------------|
| `create_node_type` | `tenant_id`, `name`, `description?`, `schema?` | Create node type |
| `get_node_type` | `id`, `tenant_id` | Get node type by ID |
| `update_node_type` | `id`, `tenant_id`, `name?`, `description?`, `schema?` | Update node type |
| `delete_node_type` | `id`, `tenant_id` | Delete node type |
| `list_node_types` | `tenant_id`, `pagination?` | List node types |

### Node Service

| Method | Parameters | Description |
|--------|-----------|-------------|
| `create_node` | `tenant_id`, `node_type_id`, `data?` | Create node |
| `get_node` | `id`, `tenant_id` | Get node by ID |
| `update_node` | `id`, `tenant_id`, `data?` | Update node |
| `delete_node` | `id`, `tenant_id` | Delete node |
| `list_nodes` | `tenant_id`, `node_type_id?`, `pagination?` | List nodes |

### Relationship Service

| Method | Parameters | Description |
|--------|-----------|-------------|
| `create_relationship` | `tenant_id`, `source_node_id`, `target_node_id`, `relationship_type`, `data?` | Create relationship |
| `get_relationship` | `id`, `tenant_id` | Get relationship by ID |
| `update_relationship` | `id`, `tenant_id`, `relationship_type?`, `data?` | Update relationship |
| `delete_relationship` | `id`, `tenant_id` | Delete relationship |
| `list_relationships` | `tenant_id`, `source_node_id?`, `target_node_id?`, `relationship_type?`, `pagination?` | List relationships |

## Error Codes

| Code | Description |
|------|-------------|
| `-32700` | Parse error |
| `-32600` | Invalid request |
| `-32601` | Method not found |
| `-32602` | Invalid params (validation error) |
| `-32603` | Internal error |
| `-32001` | Not found (custom) |

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

## Comparison with Go Backend

| Feature | Go Backend | Python Backend |
|---------|-----------|----------------|
| Protocol | gRPC | JSON-RPC |
| Port (default) | 50051 | 5000 |
| Database | PostgreSQL (pgx) | PostgreSQL (asyncpg) |
| Async | Goroutines | asyncio |
| API Definition | Protocol Buffers | JSON-RPC methods |

Both backends provide identical functionality and share the same database schema.

## REST API Wrapper

The Python backend includes an optional REST API wrapper built with FastAPI that provides:

- **RESTful endpoints** for all JSON-RPC methods
- **Swagger UI** at `/docs` for interactive API exploration
- **ReDoc** at `/redoc` for alternative API documentation
- **OpenAPI schema** at `/openapi.json`
- **Pydantic validation** for all request/response models

### Running the REST Wrapper

```bash
# First, start the JSON-RPC backend
python main.py

# In another terminal, start the REST wrapper
pip install -r rest_wrapper/requirements.txt
PYTHONPATH=. uvicorn rest_wrapper.main:app --host 0.0.0.0 --port 8000
```

### REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/tenants` | Create a tenant |
| GET | `/tenants/{id}` | Get a tenant |
| PUT | `/tenants/{id}` | Update a tenant |
| DELETE | `/tenants/{id}` | Delete a tenant |
| GET | `/tenants` | List tenants |
| POST | `/users` | Create a user |
| GET | `/users/{id}` | Get a user |
| PUT | `/users/{id}` | Update a user |
| DELETE | `/users/{id}` | Delete a user |
| GET | `/users` | List users |
| POST | `/tenants/{tenant_id}/users` | Add user to tenant |
| DELETE | `/tenants/{tenant_id}/users/{user_id}` | Remove user from tenant |
| GET | `/tenants/{tenant_id}/users` | List tenant users |
| POST | `/tenants/{tenant_id}/node-types` | Create a node type |
| GET | `/tenants/{tenant_id}/node-types/{id}` | Get a node type |
| PUT | `/tenants/{tenant_id}/node-types/{id}` | Update a node type |
| DELETE | `/tenants/{tenant_id}/node-types/{id}` | Delete a node type |
| GET | `/tenants/{tenant_id}/node-types` | List node types |
| POST | `/tenants/{tenant_id}/nodes` | Create a node |
| GET | `/tenants/{tenant_id}/nodes/{id}` | Get a node |
| PUT | `/tenants/{tenant_id}/nodes/{id}` | Update a node |
| DELETE | `/tenants/{tenant_id}/nodes/{id}` | Delete a node |
| GET | `/tenants/{tenant_id}/nodes` | List nodes |
| POST | `/tenants/{tenant_id}/relationships` | Create a relationship |
| GET | `/tenants/{tenant_id}/relationships/{id}` | Get a relationship |
| PUT | `/tenants/{tenant_id}/relationships/{id}` | Update a relationship |
| DELETE | `/tenants/{tenant_id}/relationships/{id}` | Delete a relationship |
| GET | `/tenants/{tenant_id}/relationships` | List relationships |

For more details, see [rest_wrapper/README.md](rest_wrapper/README.md).

## License

MIT License
