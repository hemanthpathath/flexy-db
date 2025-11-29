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

// PostgresNodeRepository implements NodeRepository with PostgreSQL
type PostgresNodeRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresNodeRepository creates a new PostgresNodeRepository
func NewPostgresNodeRepository(pool *pgxpool.Pool) *PostgresNodeRepository {
	return &PostgresNodeRepository{pool: pool}
}

// Create creates a new node
func (r *PostgresNodeRepository) Create(ctx context.Context, node *Node) (*Node, error) {
	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	if node.Data == "" {
		node.Data = "{}"
	}

	query := `
		INSERT INTO nodes (id, tenant_id, node_type_id, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6)
		RETURNING id, tenant_id, node_type_id, data::text, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		node.ID, node.TenantID, node.NodeTypeID, node.Data, node.CreatedAt, node.UpdatedAt,
	).Scan(&node.ID, &node.TenantID, &node.NodeTypeID, &node.Data, &node.CreatedAt, &node.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	return node, nil
}

// GetByID retrieves a node by ID and tenant ID
func (r *PostgresNodeRepository) GetByID(ctx context.Context, tenantID, id string) (*Node, error) {
	query := `
		SELECT id, tenant_id, node_type_id, data::text, created_at, updated_at 
		FROM nodes 
		WHERE id = $1 AND tenant_id = $2
	`

	node := &Node{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&node.ID, &node.TenantID, &node.NodeTypeID, &node.Data, &node.CreatedAt, &node.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return node, nil
}

// Update updates an existing node
func (r *PostgresNodeRepository) Update(ctx context.Context, node *Node) (*Node, error) {
	node.UpdatedAt = time.Now()

	if node.Data == "" {
		node.Data = "{}"
	}

	query := `
		UPDATE nodes 
		SET data = $3::jsonb, updated_at = $4
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, node_type_id, data::text, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		node.ID, node.TenantID, node.Data, node.UpdatedAt,
	).Scan(&node.ID, &node.TenantID, &node.NodeTypeID, &node.Data, &node.CreatedAt, &node.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	return node, nil
}

// Delete deletes a node by ID and tenant ID
func (r *PostgresNodeRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM nodes WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves nodes with pagination and optional filtering
func (r *PostgresNodeRepository) List(ctx context.Context, tenantID, nodeTypeID string, opts ListOptions) ([]*Node, *ListResult, error) {
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

	// Build query with optional node_type_id filter
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM nodes WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	if nodeTypeID != "" {
		countQuery += " AND node_type_id = $2"
		args = append(args, nodeTypeID)
	}

	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count nodes: %w", err)
	}

	query := `
		SELECT id, tenant_id, node_type_id, data::text, created_at, updated_at 
		FROM nodes 
		WHERE tenant_id = $1
	`
	listArgs := []interface{}{tenantID}
	argIdx := 2

	if nodeTypeID != "" {
		query += fmt.Sprintf(" AND node_type_id = $%d", argIdx)
		listArgs = append(listArgs, nodeTypeID)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	listArgs = append(listArgs, opts.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		node := &Node{}
		if err := rows.Scan(&node.ID, &node.TenantID, &node.NodeTypeID, &node.Data, &node.CreatedAt, &node.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan node: %w", err)
		}
		nodes = append(nodes, node)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(nodes)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return nodes, result, nil
}
