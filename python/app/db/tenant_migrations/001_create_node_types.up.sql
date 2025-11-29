-- Migration: 001_create_node_types.up.sql
-- Create node_types table for tenant database
-- Note: No tenant_id column needed - each tenant has its own database

CREATE TABLE IF NOT EXISTS node_types (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT,
    schema      JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name)
);

CREATE INDEX IF NOT EXISTS idx_node_types_name ON node_types(name);

