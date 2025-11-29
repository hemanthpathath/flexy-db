-- Migration: 003_create_node_types.up.sql
-- Create node_types table

CREATE TABLE IF NOT EXISTS node_types (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    schema      JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_node_types_tenant_id ON node_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_node_types_name ON node_types(name);
