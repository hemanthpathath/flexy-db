package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresNodeTypeRepository implements NodeTypeRepository with PostgreSQL
type PostgresNodeTypeRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresNodeTypeRepository creates a new PostgresNodeTypeRepository
func NewPostgresNodeTypeRepository(pool *pgxpool.Pool) *PostgresNodeTypeRepository {
	return &PostgresNodeTypeRepository{pool: pool}
}

// Create creates a new node type
func (r *PostgresNodeTypeRepository) Create(ctx context.Context, nodeType *NodeType) (*NodeType, error) {
	nodeType.ID = uuid.New().String()
	nodeType.CreatedAt = time.Now()
	nodeType.UpdatedAt = time.Now()

	query := `
		INSERT INTO node_types (id, tenant_id, name, description, schema, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
		RETURNING id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at
	`

	var schema *string
	if nodeType.Schema != "" {
		schema = &nodeType.Schema
	}

	err := r.pool.QueryRow(ctx, query,
		nodeType.ID, nodeType.TenantID, nodeType.Name, nodeType.Description, schema, nodeType.CreatedAt, nodeType.UpdatedAt,
	).Scan(&nodeType.ID, &nodeType.TenantID, &nodeType.Name, &nodeType.Description, &nodeType.Schema, &nodeType.CreatedAt, &nodeType.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create node type: %w", err)
	}

	return nodeType, nil
}

// GetByID retrieves a node type by ID and tenant ID
func (r *PostgresNodeTypeRepository) GetByID(ctx context.Context, tenantID, id string) (*NodeType, error) {
	query := `
		SELECT id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
		FROM node_types 
		WHERE id = $1 AND tenant_id = $2
	`

	nodeType := &NodeType{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&nodeType.ID, &nodeType.TenantID, &nodeType.Name, &nodeType.Description, &nodeType.Schema, &nodeType.CreatedAt, &nodeType.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get node type: %w", err)
	}

	return nodeType, nil
}

// Update updates an existing node type
func (r *PostgresNodeTypeRepository) Update(ctx context.Context, nodeType *NodeType) (*NodeType, error) {
	nodeType.UpdatedAt = time.Now()

	var schema *string
	if nodeType.Schema != "" {
		schema = &nodeType.Schema
	}

	query := `
		UPDATE node_types 
		SET name = $3, description = $4, schema = $5::jsonb, updated_at = $6
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		nodeType.ID, nodeType.TenantID, nodeType.Name, nodeType.Description, schema, nodeType.UpdatedAt,
	).Scan(&nodeType.ID, &nodeType.TenantID, &nodeType.Name, &nodeType.Description, &nodeType.Schema, &nodeType.CreatedAt, &nodeType.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update node type: %w", err)
	}

	return nodeType, nil
}

// Delete deletes a node type by ID and tenant ID
func (r *PostgresNodeTypeRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM node_types WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete node type: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves node types with pagination
func (r *PostgresNodeTypeRepository) List(ctx context.Context, tenantID string, opts ListOptions) ([]*NodeType, *ListResult, error) {
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100
	}

	offset := 0
	if opts.PageToken != "" {
		var err error
		offset, err = strconv.Atoi(opts.PageToken)
		if err != nil {
			offset = 0
		}
	}

	// Get total count
	var totalCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM node_types WHERE tenant_id = $1", tenantID).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count node types: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, description, COALESCE(schema::text, ''), created_at, updated_at 
		FROM node_types 
		WHERE tenant_id = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list node types: %w", err)
	}
	defer rows.Close()

	var nodeTypes []*NodeType
	for rows.Next() {
		nt := &NodeType{}
		if err := rows.Scan(&nt.ID, &nt.TenantID, &nt.Name, &nt.Description, &nt.Schema, &nt.CreatedAt, &nt.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan node type: %w", err)
		}
		nodeTypes = append(nodeTypes, nt)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(nodeTypes)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return nodeTypes, result, nil
}
