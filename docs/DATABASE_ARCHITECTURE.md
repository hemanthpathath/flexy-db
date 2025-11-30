# Database Architecture - Database-Per-Tenant

This document explains how the database architecture evolves as tenants and users are created.

## Overview

The system uses a **database-per-tenant** architecture:
- **Control Database** (`dbaas_control`): Stores tenant metadata and cross-tenant data (users)
- **Tenant Databases** (`dbaas_tenant_{slug}`): One database per tenant for isolated data

## Architecture States

### Stage 0: Initial State (Before First Tenant)

**Databases:**
```
PostgreSQL Server
└── dbaas_control  (Control Database)
    ├── tenants (empty)
    ├── tenant_databases (empty)
    ├── users (empty)
    ├── tenant_users (empty)
    └── tenant_migrations (empty)
```

**Tables in Control Database:**
- `tenants` - Tenant metadata (empty)
- `tenant_databases` - Mapping of tenant_id → database_name (empty)
- `users` - Global user registry (empty)
- `tenant_users` - User-tenant membership (empty)
- `tenant_migrations` - Migration tracking per tenant (empty)

**Tenant Databases:**
- None yet

---

### Stage 1: After Creating First Tenant

**When you create tenant "acme-corp":**

1. **Control Database** gets updated:
   ```
   dbaas_control
   ├── tenants
   │   └── {id: "uuid-1", slug: "acme-corp", name: "Acme Corp", ...}
   └── tenant_databases
       └── {tenant_id: "uuid-1", database_name: "dbaas_tenant_acme_corp"}
   ```

2. **New Tenant Database** is created:
   ```
   dbaas_tenant_acme_corp
   ├── node_types (empty)
   ├── nodes (empty)
   ├── relationships (empty)
   └── schema_migrations (tracks tenant-specific migrations)
   ```

**Full Architecture:**
```
PostgreSQL Server
├── dbaas_control
│   ├── tenants [1 tenant]
│   ├── tenant_databases [1 mapping]
│   ├── users (empty)
│   ├── tenant_users (empty)
│   └── tenant_migrations [tracks migrations for tenant]
│
└── dbaas_tenant_acme_corp
    ├── node_types (empty - no tenant_id column)
    ├── nodes (empty - no tenant_id column)
    └── relationships (empty - no tenant_id column)
```

---

### Stage 2: After Creating Users

**When you create users "john@example.com" and "jane@example.com":**

Users are stored in the **control database** (they're cross-tenant):

```
dbaas_control
├── tenants [1 tenant]
├── tenant_databases [1 mapping]
├── users [2 users]
│   ├── {id: "uuid-u1", email: "john@example.com", display_name: "John Doe"}
│   └── {id: "uuid-u2", email: "jane@example.com", display_name: "Jane Smith"}
└── tenant_users (empty - no memberships yet)
```

**Tenant databases remain unchanged** - users are not stored in tenant databases.

---

### Stage 3: After Adding Users to Tenant

**When you add users to the tenant:**

```
dbaas_control
├── tenants [1 tenant]
├── tenant_databases [1 mapping]
├── users [2 users]
└── tenant_users [2 memberships]
    ├── {tenant_id: "uuid-1", user_id: "uuid-u1", role: "admin"}
    └── {tenant_id: "uuid-1", user_id: "uuid-u2", role: "member"}
```

**Tenant databases remain unchanged** - membership is stored in control DB only.

---

### Stage 4: After Creating Second Tenant

**When you create tenant "tech-startup":**

```
PostgreSQL Server
├── dbaas_control
│   ├── tenants [2 tenants]
│   │   ├── {id: "uuid-1", slug: "acme-corp", ...}
│   │   └── {id: "uuid-2", slug: "tech-startup", ...}
│   ├── tenant_databases [2 mappings]
│   │   ├── {tenant_id: "uuid-1", database_name: "dbaas_tenant_acme_corp"}
│   │   └── {tenant_id: "uuid-2", database_name: "dbaas_tenant_tech_startup"}
│   ├── users [2 users - shared across tenants]
│   └── tenant_users [2 memberships - only for tenant 1]
│
├── dbaas_tenant_acme_corp
│   ├── node_types (maybe some data)
│   ├── nodes (maybe some data)
│   └── relationships (maybe some data)
│
└── dbaas_tenant_tech_startup
    ├── node_types (empty)
    ├── nodes (empty)
    └── relationships (empty)
```

---

### Stage 5: Full Example with Data

**Complete example with both tenants having data and users:**

```
PostgreSQL Server
│
├── dbaas_control (Control Database)
│   ├── tenants
│   │   ├── {id: "t1", slug: "acme-corp", name: "Acme Corp", status: "active"}
│   │   └── {id: "t2", slug: "tech-startup", name: "Tech Startup", status: "active"}
│   │
│   ├── tenant_databases
│   │   ├── {tenant_id: "t1", database_name: "dbaas_tenant_acme_corp"}
│   │   └── {tenant_id: "t2", database_name: "dbaas_tenant_tech_startup"}
│   │
│   ├── users (Cross-tenant - shared by all tenants)
│   │   ├── {id: "u1", email: "john@example.com", display_name: "John Doe"}
│   │   ├── {id: "u2", email: "jane@example.com", display_name: "Jane Smith"}
│   │   └── {id: "u3", email: "bob@techstartup.com", display_name: "Bob Developer"}
│   │
│   └── tenant_users (Cross-tenant memberships)
│       ├── {tenant_id: "t1", user_id: "u1", role: "admin"}
│       ├── {tenant_id: "t1", user_id: "u2", role: "member"}
│       └── {tenant_id: "t2", user_id: "u3", role: "admin"}
│
├── dbaas_tenant_acme_corp (Tenant 1 Database)
│   ├── node_types
│   │   └── {id: "nt1", name: "Task", description: "...", schema: "{}"}
│   ├── nodes
│   │   ├── {id: "n1", node_type_id: "nt1", data: '{"title": "Task 1"}'}
│   │   └── {id: "n2", node_type_id: "nt1", data: '{"title": "Task 2"}'}
│   └── relationships
│       └── {id: "r1", source_node_id: "n1", target_node_id: "n2", type: "depends_on"}
│
└── dbaas_tenant_tech_startup (Tenant 2 Database)
    ├── node_types
    │   └── {id: "nt2", name: "Project", description: "...", schema: "{}"}
    ├── nodes
    │   └── {id: "n3", node_type_id: "nt2", data: '{"name": "Project Alpha"}'}
    └── relationships (empty)
```

## Key Points

### Data Isolation
- **Complete isolation**: Each tenant's data is in a separate database
- **No cross-tenant queries possible**: Can't accidentally query across tenants
- **No tenant_id filtering needed**: Queries are simpler (no WHERE tenant_id = X)

### Cross-Tenant Data
- **Users**: Stored in control database (users can belong to multiple tenants)
- **User-Tenant Membership**: Stored in control database (`tenant_users` table)
- **Tenant Metadata**: Stored in control database (`tenants` table)

### Tenant-Specific Data
- **Node Types**: Each tenant database has its own `node_types`
- **Nodes**: Each tenant database has its own `nodes`
- **Relationships**: Each tenant database has its own `relationships`
- **No tenant_id columns**: Not needed since each database is tenant-scoped

## Database Naming Convention

- **Control DB**: `dbaas_control` (fixed name)
- **Tenant DBs**: `dbaas_tenant_{sanitized_slug}`
  - Example: `acme-corp` → `dbaas_tenant_acme_corp`
  - Example: `tech-startup` → `dbaas_tenant_tech_startup`

## Migration Flow

### Control Database Migrations
- Run once on startup
- Applied to `dbaas_control` database
- Tracked in `schema_migrations` table in control DB

### Tenant Database Migrations
- Run automatically when tenant database is created
- Applied to each tenant database separately
- Tracked in:
  - `schema_migrations` table in each tenant DB
  - `tenant_migrations` table in control DB (for cross-tenant tracking)

## Query Flow Example

### Creating a Node for Tenant "acme-corp":

1. **API receives request**: `POST /tenants/acme-corp-id/nodes`
2. **Resolve tenant database**: Look up `acme-corp-id` → `dbaas_tenant_acme_corp`
3. **Get connection pool**: Get or create pool for `dbaas_tenant_acme_corp`
4. **Execute query**: `INSERT INTO nodes ...` (no tenant_id needed!)
5. **Return result**: Node created in tenant's isolated database

### Listing All Tenants:

1. **API receives request**: `GET /tenants`
2. **Query control database**: `SELECT * FROM tenants`
3. **Return results**: List of all tenants

### Finding Users for a Tenant:

1. **API receives request**: `GET /tenants/acme-corp-id/users`
2. **Query control database**: `SELECT u.* FROM users u JOIN tenant_users tu ON u.id = tu.user_id WHERE tu.tenant_id = 'acme-corp-id'`
3. **Return results**: Users belonging to that tenant

## Visual Summary

```
┌─────────────────────────────────────────────────────────────-┐
│                    PostgreSQL Server                         │
├─────────────────────────────────────────────────────────────-┤
│                                                              │
│  ┌──────────────────────┐                                    │
│  │  dbaas_control       │  (Meta database)                   │
│  │  - tenants           │                                    │
│  │  - tenant_databases  │  (Mapping table)                   │
│  │  - users             │  (Cross-tenant)                    │
│  │  - tenant_users      │  (Memberships)                     │
│  └──────────────────────┘                                    │
│                                                              │
│  ┌──────────────────────┐  ┌──────────────────────┐          │
│  │ dbaas_tenant_acme    │  │ dbaas_tenant_tech    │          │
│  │ - node_types         │  │ - node_types         │          │
│  │ - nodes              │  │ - nodes              │          │
│  │ - relationships      │  │ - relationships      │          │
│  └──────────────────────┘  └──────────────────────┘          │
│                                                              │
│  (Each tenant gets its own isolated database)                │
│                                                              │
└────────────────────────────────────────────────────────────-─┘
```

## Benefits

1. **Complete Data Isolation**: Impossible to accidentally query wrong tenant
2. **Better Performance**: No tenant_id filtering needed in queries
3. **Independent Scaling**: Can move large tenants to separate servers
4. **Easier Backups**: Per-tenant backup/restore
5. **Simpler Queries**: Clean SQL without tenant_id everywhere
6. **Schema Flexibility**: Could customize schema per tenant (advanced use case)

