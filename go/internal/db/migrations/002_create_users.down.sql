-- Migration: 002_create_users.down.sql
-- Drop users and tenant_users tables

DROP TABLE IF EXISTS tenant_users CASCADE;
DROP TABLE IF EXISTS users CASCADE;
