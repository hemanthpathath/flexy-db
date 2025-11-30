-- Migration: 001_create_control_tables.up.sql
-- Control database: tenant metadata and cross-tenant data

-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tenant database mapping: maps tenant_id to database name
CREATE TABLE IF NOT EXISTS tenant_databases (
    tenant_id     UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    database_name TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status        TEXT NOT NULL DEFAULT 'active'
);

-- Users (cross-tenant: users can belong to multiple tenants)
CREATE TABLE IF NOT EXISTS users (
    id           UUID PRIMARY KEY,
    email        TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tenant-User membership (cross-tenant)
CREATE TABLE IF NOT EXISTS tenant_users (
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         TEXT NOT NULL DEFAULT 'member',
    status       TEXT NOT NULL DEFAULT 'active',
    PRIMARY KEY (tenant_id, user_id)
);

-- Track migrations per tenant database
CREATE TABLE IF NOT EXISTS tenant_migrations (
    tenant_id    UUID NOT NULL,
    version      TEXT NOT NULL,
    applied_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, version)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenant_databases_name ON tenant_databases(database_name);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_id ON tenant_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_users_user_id ON tenant_users(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_migrations_tenant_id ON tenant_migrations(tenant_id);

