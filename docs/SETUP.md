# Local Development Setup Guide

This guide will help you set up and run flex-db locally, and test the APIs using Insomnia.

> **Note:** All commands in this guide should be run from the project root directory (`flex-db/`).

## Quick Start (TL;DR)

```bash
# 1. Start PostgreSQL
docker-compose up -d

# 2. Set up environment variables
cp .env.example .env.local

# 3. Run the server (this script handles everything automatically)
./scripts/start.sh
```

Then open Insomnia, create a gRPC request, connect to `localhost:50051`, and import `api/proto/dbaas.proto`.

## Prerequisites

- **Go 1.21+** installed
- **Docker** (for PostgreSQL) - recommended for easy setup
- **Insomnia** (for testing gRPC APIs) - version 8.0+ with gRPC support

## Step 1: Set Up PostgreSQL Database

We recommend using Docker for PostgreSQL as it's the easiest setup method.

### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL using docker-compose
docker-compose up -d

# Verify it's running
docker ps | grep flex-db-postgres

# Check logs if needed
docker logs flex-db-postgres
```

The Docker setup will:
- Create a PostgreSQL 14 container named `flex-db-postgres`
- Use default credentials: `postgres/postgres`
- Create database `dbaas`
- Expose PostgreSQL on port `5432`
- Persist data in a Docker volume

### Option B: Using Local PostgreSQL Installation

If you prefer to use a local PostgreSQL installation:

```bash
# Create database
createdb dbaas

# Or using psql
psql -U postgres
CREATE DATABASE dbaas;
\q
```

Make sure to update `.env.local` (created in Step 2) with your PostgreSQL connection details.

## Step 2: Configure Environment Variables

The project uses environment variables for configuration. Create your local environment file:

```bash
cp .env.example .env.local
```

### Environment File Structure

- **`.env.example`** - Template file (committed to repo) with default values
- **`.env.local`** - Your personal local configuration (gitignored, never committed)

### Default Configuration

The `.env.local` file will contain:

```bash
# Local Development Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dbaas
DB_SSL_MODE=disable

# Local gRPC Server Configuration
GRPC_PORT=50051
```

**Customize `.env.local` if:**
- You're using a local PostgreSQL installation with different credentials
- You want to use a different database name
- You need to change the gRPC port

**Note:** 
- If `.env.local` doesn't exist, the server will use default values from the code
- The `start.sh` script automatically loads `.env.local` if it exists
- For manual runs, use `source scripts/load-env.sh` to load variables

## Step 3: Install Dependencies

Download Go dependencies:

```bash
go mod download
```

**Note:** The `start.sh` script will do this automatically if you use it.

## Step 4: Run the Server

You have three options to run the server:

### Option 1: Using the Start Script (Recommended ⭐)

The easiest way - the script handles everything automatically:

```bash
./scripts/start.sh
```

**What the script does:**
1. Loads environment variables from `.env.local` (if exists)
2. Checks if PostgreSQL container is running, starts it if needed
3. Downloads Go dependencies
4. Runs database migrations automatically
5. Starts the gRPC server

### Option 2: Manual Run

If you prefer to run manually:

```bash
# Load environment variables from .env.local
source scripts/load-env.sh

# Run the server
go run ./cmd/dbaas-server
```

### Option 3: Build and Run

For production-like builds:

```bash
# Load environment variables
source scripts/load-env.sh

# Build the binary
go build -o dbaas-server ./cmd/dbaas-server

# Run the binary
./dbaas-server
```

### Expected Output

When the server starts successfully, you should see:

```
Connecting to database...
Connected to database successfully
Running database migrations...
Migration 001_create_tenants already applied, skipping
Migration 002_create_users already applied, skipping
Migration 003_create_node_types already applied, skipping
Migration 004_create_nodes already applied, skipping
Migration 005_create_relationships already applied, skipping
Migrations completed successfully
Starting gRPC server on port 50051...
```

The server will run until you stop it with `Ctrl+C`.

## Step 5: Test with Insomnia

The server has gRPC reflection enabled, which allows Insomnia to automatically discover all available services and methods.

### 5.1 Create a gRPC Request

1. Open **Insomnia**
2. Click **"New Request"** → **"gRPC Request"**
3. In the **URL** field, enter: `localhost:50051`
   - Use `localhost:50051` (default)
   - Or whatever port you set in `GRPC_PORT` in `.env.local`

### 5.2 Import Proto File

1. Click **"Select Proto File"** or **"Use Proto File"**
2. Navigate to the project directory and select: `api/proto/dbaas.proto`
3. Insomnia will parse the proto file and show available services

### 5.3 Select a Service and Method

After importing the proto file, you'll see all available services:

- **`TenantService`** - Manage tenants (create, get, update, delete, list)
- **`UserService`** - Manage users and tenant associations
- **`NodeTypeService`** - Define node schemas for tenants
- **`NodeService`** - Create and manage nodes (data instances)
- **`RelationshipService`** - Create relationships between nodes

1. Select a service from the dropdown (e.g., `TenantService`)
2. Select a method (e.g., `CreateTenant`)
3. Fill in the request body (see examples below)
4. Click **"Send"**

### 5.3 Test API Calls

Here are example requests for each service:

#### Create a Tenant

**Service:** `TenantService.CreateTenant`

**Request Body:**
```json
{
  "slug": "acme-corp",
  "name": "Acme Corporation"
}
```

#### Get a Tenant

**Service:** `TenantService.GetTenant`

**Request Body:**
```json
{
  "id": "TENANT_ID_FROM_CREATE_RESPONSE"
}
```

#### List Tenants

**Service:** `TenantService.ListTenants`

**Request Body:**
```json
{
  "pagination": {
    "page_size": 10
  }
}
```

#### Create a User

**Service:** `UserService.CreateUser`

**Request Body:**
```json
{
  "email": "john@example.com",
  "display_name": "John Doe"
}
```

#### Add User to Tenant

**Service:** `UserService.AddUserToTenant`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "user_id": "USER_ID",
  "role": "admin"
}
```

#### Create a NodeType

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

#### Create a Node

**Service:** `NodeService.CreateNode`

**Request Body:**
```json
{
  "tenant_id": "TENANT_ID",
  "node_type_id": "NODE_TYPE_ID",
  "data": "{\"title\": \"Complete project\", \"priority\": \"high\"}"
}
```

#### Create a Relationship

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

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs flex-db-postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Port Already in Use

If port 50051 is already in use:

```bash
# Find process using the port
lsof -i :50051

# Option 1: Kill the process
kill -9 <PID>

# Option 2: Change the port in .env.local
# Edit .env.local and change GRPC_PORT=50051 to GRPC_PORT=50052
# Then update Insomnia URL to localhost:50052
```

### Migration Errors

If you see migration errors, you might need to reset the database:

**Using Docker:**
```bash
# Stop and remove containers, volumes, and networks
docker-compose down -v

# Start fresh
docker-compose up -d

# Wait a few seconds, then restart the server
./scripts/start.sh
```

**Using Local PostgreSQL:**
```bash
psql -U postgres -d dbaas
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
\q
```

Then restart the server - migrations will run automatically.

### Environment Variables Not Loading

If environment variables aren't being loaded:

```bash
# Check if .env.local exists
ls -la .env.local

# If using start script, it loads automatically
# If running manually, make sure to source the load script:
source scripts/load-env.sh

# Verify variables are loaded
echo $DB_HOST
echo $GRPC_PORT
```

## Alternative: Using grpcurl (Command Line)

If you prefer command line testing instead of Insomnia:

```bash
# Install grpcurl (macOS)
brew install grpcurl

# Or install via Go
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Example Commands

```bash
# List all available services
grpcurl -plaintext localhost:50051 list

# List methods for a specific service
grpcurl -plaintext localhost:50051 list dbaas.TenantService

# Create a tenant
grpcurl -plaintext -d '{"slug": "acme-corp", "name": "Acme Corporation"}' \
  localhost:50051 dbaas.TenantService/CreateTenant

# List tenants
grpcurl -plaintext -d '{"pagination": {"page_size": 10}}' \
  localhost:50051 dbaas.TenantService/ListTenants
```

**Note:** Replace `localhost:50051` with your server URL if using a different port.

## Project Structure

```
flex-db/
├── api/proto/              # gRPC protobuf definitions
│   ├── dbaas.proto         # Proto definition file
│   ├── dbaas.pb.go         # Generated Go code
│   └── dbaas_grpc.pb.go    # Generated gRPC code
├── cmd/dbaas-server/       # Main server entry point
│   └── main.go
├── internal/
│   ├── db/                 # Database connection and migrations
│   │   ├── db.go
│   │   └── migrations/     # SQL migration files
│   ├── repository/         # Data access layer
│   ├── service/            # Business logic layer
│   └── grpc/               # gRPC handlers
├── scripts/
│   ├── start.sh            # Quick start script
│   └── load-env.sh         # Environment variable loader
├── docker-compose.yml      # PostgreSQL Docker setup
├── .env.example            # Environment variable template
├── .env.local              # Your local config (gitignored)
└── docs/                   # Documentation
    ├── SETUP.md            # This file
    └── INSOMNIA_GUIDE.md   # Insomnia testing guide
```

## API Reference

All services support standard CRUD operations:

### TenantService
- `CreateTenant` - Create a new tenant
- `GetTenant` - Get tenant by ID
- `UpdateTenant` - Update tenant details
- `DeleteTenant` - Delete a tenant
- `ListTenants` - List all tenants with pagination

### UserService
- `CreateUser` - Create a new user
- `GetUser` - Get user by ID
- `UpdateUser` - Update user details
- `DeleteUser` - Delete a user
- `ListUsers` - List all users with pagination
- `AddUserToTenant` - Add user to a tenant
- `RemoveUserFromTenant` - Remove user from tenant
- `ListTenantUsers` - List users in a tenant

### NodeTypeService
- `CreateNodeType` - Create a node type for a tenant
- `GetNodeType` - Get node type by ID
- `UpdateNodeType` - Update node type
- `DeleteNodeType` - Delete a node type
- `ListNodeTypes` - List node types for a tenant

### NodeService
- `CreateNode` - Create a new node
- `GetNode` - Get node by ID
- `UpdateNode` - Update node data
- `DeleteNode` - Delete a node
- `ListNodes` - List nodes with optional filtering

### RelationshipService
- `CreateRelationship` - Create relationship between nodes
- `GetRelationship` - Get relationship by ID
- `UpdateRelationship` - Update relationship data
- `DeleteRelationship` - Delete a relationship
- `ListRelationships` - List relationships with filtering

