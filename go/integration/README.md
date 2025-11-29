# Integration Tests

Integration tests for the flex-db service layer that test against a real PostgreSQL database.

## Overview

These tests verify:
- ✅ Actual SQL queries work correctly
- ✅ Database constraints (unique, foreign keys)
- ✅ Tenant isolation at database level
- ✅ JSONB operations
- ✅ Pagination logic
- ✅ Repository implementations

## Prerequisites

1. **PostgreSQL** running (can use the same instance as development)
2. **Test Database** created (recommended to use separate test DB)

## Setup

### Option 1: Use Separate Test Database (Recommended)

```bash
# Create test database
createdb dbaas_test

# Or using psql
psql -U postgres
CREATE DATABASE dbaas_test;
\q
```

Set environment variables:
```bash
export TEST_DB_NAME=dbaas_test
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_SSL_MODE=disable
```

### Option 2: Use Existing Database

Tests will use the default database (`dbaas`) if environment variables are not set.

⚠️ **Warning**: Tests truncate tables between runs. Don't use your development database unless you're okay with data loss.

## Running Tests

### Run All Integration Tests

```bash
go test ./integration/... -v
```

### Run with Coverage

```bash
go test ./integration/... -cover
```

### Run Specific Test

```bash
go test ./integration/... -run TestTenantRepository_Create -v
```

### Run with Environment Variables

```bash
TEST_DB_NAME=dbaas_test go test ./integration/... -v
```

## Test Structure

```
integration/
├── test_setup.go              # Database setup and cleanup utilities
├── tenant_repo_test.go        # Tenant repository integration tests
├── user_repo_test.go          # User repository integration tests
├── nodetype_repo_test.go      # NodeType repository integration tests
├── node_repo_test.go          # Node repository integration tests
└── relationship_repo_test.go  # Relationship repository integration tests
```

## What Gets Tested

### Repository Layer
- **CRUD Operations**: Create, Read, Update, Delete
- **Constraints**: Unique constraints, foreign key relationships
- **Tenant Isolation**: Verifies data is properly scoped to tenants
- **Pagination**: List operations with page size and tokens
- **Filtering**: Query filtering (by node type, relationship type, etc.)
- **JSONB Operations**: JSON data storage and retrieval
- **Error Handling**: Not found errors, constraint violations

## Test Isolation

Each test:
1. Sets up a fresh database connection
2. Runs migrations automatically
3. Cleans up (truncates) tables after each test

This ensures tests don't interfere with each other.

## CI/CD Integration

For CI/CD pipelines, you can use Docker to spin up a test database:

```yaml
# Example GitHub Actions
- name: Start PostgreSQL
  run: |
    docker run -d \
      --name test-postgres \
      -e POSTGRES_PASSWORD=postgres \
      -e POSTGRES_DB=dbaas_test \
      -p 5432:5432 \
      postgres:14

- name: Run Integration Tests
  run: |
    export TEST_DB_NAME=dbaas_test
    go test ./integration/... -v
```

## Troubleshooting

### Database Connection Errors

```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# or
psql -U postgres -c "SELECT 1"
```

### Migration Errors

If migrations fail, you might need to reset the test database:

```bash
# Drop and recreate
dropdb dbaas_test
createdb dbaas_test
```

### Tests Failing with Constraint Violations

Make sure cleanup is working. Check that `cleanupTestDB` is being called with `defer`.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEST_DB_HOST` | `localhost` | Database host |
| `TEST_DB_PORT` | `5432` | Database port |
| `TEST_DB_USER` | `postgres` | Database user |
| `TEST_DB_PASSWORD` | `postgres` | Database password |
| `TEST_DB_NAME` | `dbaas_test` | Test database name |
| `TEST_DB_SSL_MODE` | `disable` | SSL mode |

