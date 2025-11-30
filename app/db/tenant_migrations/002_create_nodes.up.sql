-- Migration: 002_create_nodes.up.sql
-- Create nodes table for tenant database
-- Note: No tenant_id column needed - each tenant has its own database

CREATE TABLE IF NOT EXISTS nodes (
    id           UUID PRIMARY KEY,
    node_type_id UUID NOT NULL REFERENCES node_types(id) ON DELETE CASCADE,
    data         JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nodes_node_type_id ON nodes(node_type_id);
CREATE INDEX IF NOT EXISTS idx_nodes_data ON nodes USING GIN (data);

