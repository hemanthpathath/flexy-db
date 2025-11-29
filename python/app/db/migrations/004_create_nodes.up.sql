-- Migration: 004_create_nodes.up.sql
-- Create nodes table

CREATE TABLE IF NOT EXISTS nodes (
    id           UUID PRIMARY KEY,
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    node_type_id UUID NOT NULL REFERENCES node_types(id) ON DELETE CASCADE,
    data         JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nodes_tenant_id ON nodes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_nodes_node_type_id ON nodes(node_type_id);
CREATE INDEX IF NOT EXISTS idx_nodes_data ON nodes USING GIN (data);
