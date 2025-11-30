# Local Setup and Testing Guide

Complete guide for setting up and running flex-db locally.

## Quick Start

### Option 1: Docker Compose (Recommended)

The easiest way to run everything:

```bash
cd python
docker compose up
```

That's it! This will:
- Start PostgreSQL database
- Build and start the Python application
- Auto-create databases and run migrations

**Access the API:**
- JSON-RPC: http://localhost:5001/jsonrpc
- Health: http://localhost:5001/health
- OpenRPC: http://localhost:5001/openrpc.json

**Note:** Port 5001 is used because port 5000 is typically in use on macOS.

### Option 2: Local Python Setup

If you prefer to run Python directly on your machine:

```bash
cd python
./scripts/setup_local.sh
source venv/bin/activate
python main.py
```

The setup script will:
- Create Python virtual environment
- Install dependencies
- Create `.env.local` configuration
- Start PostgreSQL in Docker
- Verify setup

**Access the API:** http://localhost:5000/jsonrpc

---

## Prerequisites

- Python 3.9+ (for local Python setup)
- Docker and Docker Compose (for PostgreSQL)
- `curl` or any HTTP client for testing

---

## Detailed Setup Instructions

### Using Docker Compose

#### Step 1: Start Everything

```bash
cd python
docker compose up
```

To run in background:
```bash
docker compose up -d
```

#### Step 2: Verify It's Running

```bash
# Health check
curl http://localhost:5001/health

# View logs
docker compose logs -f
```

#### Step 3: Stop Services

```bash
docker compose down
```

**Clean up everything (including data):**
```bash
docker compose down -v
```

### Using Local Python

#### Step 1: Set Up Environment

Run the automated setup script:

```bash
cd python
./scripts/setup_local.sh
```

This creates:
- Python virtual environment
- `.env.local` configuration file
- Starts PostgreSQL

#### Step 2: Start the Server

```bash
source venv/bin/activate
python main.py
```

Or use the start script:

```bash
./scripts/start.sh
```

#### Step 3: Verify Setup

```bash
# Health check
curl http://localhost:5000/health
```

---

## Testing the Application

### Quick Test Script

Use the automated test script:

```bash
./scripts/test_basic_operations.sh
```

This will test:
- Health check
- Tenant creation
- User creation
- Listing tenants and users

### Manual Testing

#### Test 1: Health Check

```bash
curl http://localhost:5001/health
```

Expected: `{"status":"ok"}`

#### Test 2: Create a Tenant

```bash
curl -X POST http://localhost:5001/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_tenant",
    "params": {
      "slug": "acme-corp",
      "name": "Acme Corporation"
    },
    "id": 1
  }' | python3 -m json.tool
```

**Note:** Save the tenant `id` from the response - you'll need it for tenant-scoped operations.

#### Test 3: List All Tenants

```bash
curl -X POST http://localhost:5001/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_tenants",
    "params": {
      "pagination": {"page_size": 10}
    },
    "id": 2
  }' | python3 -m json.tool
```

#### Test 4: Create a User

```bash
curl -X POST http://localhost:5001/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "create_user",
    "params": {
      "email": "john.doe@example.com",
      "display_name": "John Doe"
    },
    "id": 3
  }' | python3 -m json.tool
```

#### Test 5: List All Users

```bash
curl -X POST http://localhost:5001/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "list_users",
    "params": {
      "pagination": {"page_size": 10}
    },
    "id": 4
  }' | python3 -m json.tool
```

---

## Verifying Tenant Database Creation

One of the key features of this architecture is that each tenant gets its own database automatically. Here's how to verify:

### Step 1: Check Current Databases

Before creating a tenant:

```bash
docker exec flex-db-python-postgres psql -U postgres -c "SELECT datname FROM pg_database WHERE datname LIKE 'dbaas%' ORDER BY datname;"
```

**Expected:** You should see at least `dbaas_control`.

### Step 2: Create a Tenant

Create a tenant via API (see Test 2 above).

### Step 3: Verify Database Was Created

Check databases again:

```bash
docker exec flex-db-python-postgres psql -U postgres -c "SELECT datname FROM pg_database WHERE datname LIKE 'dbaas%' ORDER BY datname;"
```

**Expected:** You should now see:
- `dbaas_control` (control database)
- `dbaas_tenant_acme_corp` (new tenant database!)

### Step 4: Verify Tenant Mapping

Check that the tenant is mapped to its database:

```bash
docker exec flex-db-python-postgres psql -U postgres -d dbaas_control -c "SELECT t.slug, t.name, td.database_name FROM tenants t JOIN tenant_databases td ON t.id = td.tenant_id ORDER BY t.slug;"
```

### Step 5: Verify Tenant Database Schema

Check that the tenant database has the correct tables:

```bash
docker exec flex-db-python-postgres psql -U postgres -d dbaas_tenant_acme_corp -c "\dt"
```

**Expected Tables:**
- `node_types`
- `nodes`
- `relationships`
- `schema_migrations`

### What Happens Automatically

When you create a tenant, the system automatically:

1. ✅ Creates tenant record in `dbaas_control.tenants`
2. ✅ Creates PostgreSQL database named `dbaas_tenant_{slug}`
3. ✅ Records mapping in `dbaas_control.tenant_databases`
4. ✅ Runs migrations on the new tenant database
5. ✅ Creates all required tables

**Check Application Logs:**

```bash
docker compose logs flex-db-python | grep -A 5 "Creating tenant database"
```

---

## Verify Database Setup

You can verify that the databases were created correctly:

```bash
# Connect to PostgreSQL
docker exec -it flex-db-python-postgres psql -U postgres

# List all databases
\l

# Check control database tables
\c dbaas_control
\dt

# Check tenants table
SELECT * FROM tenants;

# Check users table
SELECT * FROM users;

# Check tenant_databases mapping
SELECT * FROM tenant_databases;

# Exit
\q
```

---

## Troubleshooting

### Permission Denied

If you get "Permission denied" when running scripts:

```bash
chmod +x scripts/*.sh
```

### Port 5000/5001 Already in Use

**For Docker Compose:**
- Docker uses port 5001 by default (configured in `docker-compose.yml`)
- If 5001 is also in use, edit `docker-compose.yml` and change the port mapping

**For Local Python:**
- Change `JSONRPC_PORT` in `.env.local` to a different port
- Or stop the process using port 5000

### Cannot Connect to Database

- Ensure PostgreSQL container is running: `docker ps`
- Check if port 5432 is already in use
- Verify environment variables in `.env.local`

### Control Database Not Found

- The application should auto-create it, but you can manually create it:
  ```bash
  docker exec -it flex-db-python-postgres psql -U postgres -c "CREATE DATABASE dbaas_control;"
  ```

### Tenant Database Creation Fails

- Check PostgreSQL logs: `docker logs flex-db-python-postgres`
- Ensure the user has permission to create databases

### PostgreSQL Won't Start

```bash
# Check if port 5432 is in use
lsof -i :5432

# Stop any existing PostgreSQL
docker compose down

# Start again
docker compose up -d postgres
```

### Python Version Issues

Make sure you have Python 3.9+:

```bash
python3 --version
```

---

## Next Steps

Once you've verified tenant and user creation works:

1. **Test tenant-scoped operations** (nodes, relationships, node types)
   - These require a `tenant_id` parameter
   - Each tenant has its own isolated database

2. **Explore the OpenRPC spec**:
   ```bash
   curl http://localhost:5001/openrpc.json > openrpc.json
   ```

3. **Use the interactive OpenRPC playground**:
   - Visit https://playground.open-rpc.org/
   - Paste the contents of `openrpc.json`

4. **Read the architecture documentation**:
   - See [DATABASE_ARCHITECTURE.md](DATABASE_ARCHITECTURE.md) for details

5. **API Documentation**:
   - See [JSON_RPC_INTEGRATION.md](JSON_RPC_INTEGRATION.md) for complete API docs

---

## Clean Up

To stop and remove everything:

```bash
# Stop the application (Ctrl+C if running in foreground)

# Stop and remove containers
docker compose down

# Remove database volume (WARNING: deletes all data)
docker compose down -v
```

---

## Quick Reference Commands

**Docker Compose:**
```bash
docker compose up              # Start everything
docker compose up -d           # Start in background
docker compose down            # Stop everything
docker compose logs -f         # View logs
docker compose ps              # Check status
```

**Database Verification:**
```bash
# List all databases
docker exec flex-db-python-postgres psql -U postgres -c "\l" | grep dbaas

# List all tenants
docker exec flex-db-python-postgres psql -U postgres -d dbaas_control -c "SELECT slug, name FROM tenants;"

# List tenant-to-database mappings
docker exec flex-db-python-postgres psql -U postgres -d dbaas_control -c "SELECT t.slug, td.database_name FROM tenants t JOIN tenant_databases td ON t.id = td.tenant_id;"
```

