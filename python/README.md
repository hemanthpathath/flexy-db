# flex-db Python Backend

A **Database-as-a-Service (DBaaS)** implemented in Python with JSON-RPC. This is a Python implementation that provides feature parity with the Go backend.

## Features

- **Multi-tenant architecture**: All Nodes and Relationships are scoped to a Tenant
- **Flexible data model**: Generic NodeTypes and Nodes with JSONB data storage
- **Graph-like relationships**: Connect Nodes with typed Relationships
- **JSON-RPC API**: Full CRUD operations with pagination support
- **OpenRPC Specification**: Interactive API documentation and discovery
- **PostgreSQL backend**: Robust, production-ready database
- **Async/await**: Built with asyncio for high-performance I/O

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
│       ├── server.py       # FastAPI router for JSON-RPC
│       └── openrpc.py      # OpenRPC specification generator
├── docs/
│   └── JSON_RPC_INTEGRATION.md  # Comprehensive integration guide
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

The server will start on `localhost:5000` (JSON-RPC).

**📚 For detailed setup instructions, see the sections below.**

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `dbaas` | Database name |
| `DB_SSL_MODE` | `disable` | SSL mode |
| `JSONRPC_HOST` | `0.0.0.0` | JSON-RPC server host |
| `JSONRPC_PORT` | `5000` | JSON-RPC server port |

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

# Server Configuration
JSONRPC_HOST=0.0.0.0
JSONRPC_PORT=5000


# Development Options
RELOAD=false
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
2024-01-01 12:00:00 - INFO - Starting flex-db server on 0.0.0.0:5000...
2024-01-01 12:00:00 - INFO - JSON-RPC endpoint: http://0.0.0.0:5000/jsonrpc
2024-01-01 12:00:00 - INFO - OpenRPC spec: http://0.0.0.0:5000/openrpc.json
2024-01-01 12:00:00 - INFO - Health check: http://0.0.0.0:5000/health
```

## API Usage

The server exposes a JSON-RPC 2.0 API at `http://localhost:5000/jsonrpc`.

### OpenRPC Documentation

The service provides **OpenRPC** specification (similar to OpenAPI for REST) for interactive documentation and API discovery:

- **OpenRPC Spec**: `http://localhost:5000/openrpc.json`
- **Introspection Method**: Call `rpc.discover` via JSON-RPC to get the spec programmatically

You can use OpenRPC tooling to:
- View interactive documentation
- Generate client code in various languages
- Validate API calls
- Discover available methods dynamically

**📚 For comprehensive JSON-RPC integration guide, see [docs/JSON_RPC_INTEGRATION.md](docs/JSON_RPC_INTEGRATION.md)**

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

#### Get a Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "get_tenant",
    "params": {"id": "550e8400-e29b-41d4-a716-446655440000"},
    "id": 2
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
    "id": 3
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
    "id": 4
  }'
```

#### Add User to Tenant

```bash
curl -X POST http://localhost:5000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "add_user_to_tenant",
    "params": {
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "660e8400-e29b-41d4-a716-446655440001",
      "role": "admin"
    },
    "id": 5
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
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Task",
      "description": "A task node type",
      "schema": "{\"type\": \"object\", \"properties\": {\"title\": {\"type\": \"string\"}}}"
    },
    "id": 6
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
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "node_type_id": "770e8400-e29b-41d4-a716-446655440002",
      "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
    },
    "id": 7
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
      "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
      "source_node_id": "880e8400-e29b-41d4-a716-446655440003",
      "target_node_id": "990e8400-e29b-41d4-a716-446655440004",
      "relationship_type": "depends_on",
      "data": "{\"priority\": 1}"
    },
    "id": 8
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
            "id": id
        }
    )
    return response.json()

# Create a tenant
result = call_jsonrpc("create_tenant", {"slug": "acme-corp", "name": "Acme Corporation"})
print(result)
```

For more examples and client implementations, see [docs/JSON_RPC_INTEGRATION.md](docs/JSON_RPC_INTEGRATION.md).

## Docker Setup

### Using Docker Compose

```bash
# From python/ directory
docker-compose up -d
```

This will:
1. Start PostgreSQL container
2. Build and start the Python backend container
3. Run database migrations automatically
4. Expose the service on `http://localhost:5001` (host port 5001 maps to container port 5000)

### Using Docker directly

```bash
# Build the image
docker build -t flex-db-python .

# Run the container
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
- **Commit SHA**: `ghcr.io/<owner>/<repo>/flex-db:sha-<sha>` - Unique tag for each commit (7 character short SHA)
- **Branch name**: `ghcr.io/<owner>/<repo>/flex-db:main` - For push events
- **PR number**: `ghcr.io/<owner>/<repo>/flex-db:pr-<number>` - For pull request events

#### Pulling the Image from GHCR

To pull the image locally:

```bash
# Authenticate to GHCR (required for private repositories)
echo $GITHUB_TOKEN | docker login ghcr.io -u <username> --password-stdin

# Pull the latest image from main branch
docker pull ghcr.io/<owner>/<repo>/flex-db:main

# Pull a specific commit version (use short SHA with sha- prefix)
docker pull ghcr.io/<owner>/<repo>/flex-db:sha-<short-sha>
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

## JSON-RPC Integration

For comprehensive documentation on integrating with the JSON-RPC API, including:
- Complete API reference
- Client implementations (Python, JavaScript, Go)
- Code examples
- Best practices
- Error handling

**📚 See [docs/JSON_RPC_INTEGRATION.md](docs/JSON_RPC_INTEGRATION.md)**

## Testing

### Run Tests

```bash
python -m pytest
```

### Health Check

```bash
curl http://localhost:5000/health
```

Expected response:

```json
{"status": "ok"}
```

## Comparison with Go Backend

| Feature | Go Backend | Python Backend |
|---------|-----------|----------------|
| Protocol | gRPC | JSON-RPC |
| Language | Go | Python |
| Database Driver | pgx | asyncpg |
| API Definition | Protocol Buffers | JSON-RPC methods |
| Documentation | Protocol Buffers | OpenRPC specification |
| Performance | High (compiled) | High (async I/O) |

Both implementations provide identical functionality and can share the same database.

## License

MIT License
