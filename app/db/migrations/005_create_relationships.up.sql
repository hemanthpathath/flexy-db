-- Migration: 005_create_relationships.up.sql
-- Create relationships table

CREATE TABLE IF NOT EXISTS relationships (
    id                UUID PRIMARY KEY,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    target_node_id    UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL,
    data              JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_relationships_tenant_id ON relationships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_relationships_source_node_id ON relationships(source_node_id);
CREATE INDEX IF NOT EXISTS idx_relationships_target_node_id ON relationships(target_node_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_relationships_data ON relationships USING GIN (data);
