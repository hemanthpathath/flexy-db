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

// PostgresRelationshipRepository implements RelationshipRepository with PostgreSQL
type PostgresRelationshipRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRelationshipRepository creates a new PostgresRelationshipRepository
func NewPostgresRelationshipRepository(pool *pgxpool.Pool) *PostgresRelationshipRepository {
	return &PostgresRelationshipRepository{pool: pool}
}

// Create creates a new relationship
func (r *PostgresRelationshipRepository) Create(ctx context.Context, rel *Relationship) (*Relationship, error) {
	rel.ID = uuid.New().String()
	rel.CreatedAt = time.Now()
	rel.UpdatedAt = time.Now()

	if rel.Data == "" {
		rel.Data = "{}"
	}

	query := `
		INSERT INTO relationships (id, tenant_id, source_node_id, target_node_id, relationship_type, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
		RETURNING id, tenant_id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		rel.ID, rel.TenantID, rel.SourceNodeID, rel.TargetNodeID, rel.RelationshipType, rel.Data, rel.CreatedAt, rel.UpdatedAt,
	).Scan(&rel.ID, &rel.TenantID, &rel.SourceNodeID, &rel.TargetNodeID, &rel.RelationshipType, &rel.Data, &rel.CreatedAt, &rel.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create relationship: %w", err)
	}

	return rel, nil
}

// GetByID retrieves a relationship by ID and tenant ID
func (r *PostgresRelationshipRepository) GetByID(ctx context.Context, tenantID, id string) (*Relationship, error) {
	query := `
		SELECT id, tenant_id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at 
		FROM relationships 
		WHERE id = $1 AND tenant_id = $2
	`

	rel := &Relationship{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&rel.ID, &rel.TenantID, &rel.SourceNodeID, &rel.TargetNodeID, &rel.RelationshipType, &rel.Data, &rel.CreatedAt, &rel.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship: %w", err)
	}

	return rel, nil
}

// Update updates an existing relationship
func (r *PostgresRelationshipRepository) Update(ctx context.Context, rel *Relationship) (*Relationship, error) {
	rel.UpdatedAt = time.Now()

	if rel.Data == "" {
		rel.Data = "{}"
	}

	query := `
		UPDATE relationships 
		SET relationship_type = $3, data = $4::jsonb, updated_at = $5
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		rel.ID, rel.TenantID, rel.RelationshipType, rel.Data, rel.UpdatedAt,
	).Scan(&rel.ID, &rel.TenantID, &rel.SourceNodeID, &rel.TargetNodeID, &rel.RelationshipType, &rel.Data, &rel.CreatedAt, &rel.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update relationship: %w", err)
	}

	return rel, nil
}

// Delete deletes a relationship by ID and tenant ID
func (r *PostgresRelationshipRepository) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM relationships WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete relationship: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves relationships with pagination and optional filtering
func (r *PostgresRelationshipRepository) List(ctx context.Context, tenantID, sourceNodeID, targetNodeID, relType string, opts ListOptions) ([]*Relationship, *ListResult, error) {
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

	// Build query with optional filters
	countQuery := "SELECT COUNT(*) FROM relationships WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if sourceNodeID != "" {
		countQuery += fmt.Sprintf(" AND source_node_id = $%d", argIdx)
		args = append(args, sourceNodeID)
		argIdx++
	}
	if targetNodeID != "" {
		countQuery += fmt.Sprintf(" AND target_node_id = $%d", argIdx)
		args = append(args, targetNodeID)
		argIdx++
	}
	if relType != "" {
		countQuery += fmt.Sprintf(" AND relationship_type = $%d", argIdx)
		args = append(args, relType)
		argIdx++
	}

	var totalCount int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count relationships: %w", err)
	}

	// Build list query
	query := `
		SELECT id, tenant_id, source_node_id, target_node_id, relationship_type, data::text, created_at, updated_at 
		FROM relationships 
		WHERE tenant_id = $1
	`
	listArgs := []interface{}{tenantID}
	listArgIdx := 2

	if sourceNodeID != "" {
		query += fmt.Sprintf(" AND source_node_id = $%d", listArgIdx)
		listArgs = append(listArgs, sourceNodeID)
		listArgIdx++
	}
	if targetNodeID != "" {
		query += fmt.Sprintf(" AND target_node_id = $%d", listArgIdx)
		listArgs = append(listArgs, targetNodeID)
		listArgIdx++
	}
	if relType != "" {
		query += fmt.Sprintf(" AND relationship_type = $%d", listArgIdx)
		listArgs = append(listArgs, relType)
		listArgIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", listArgIdx, listArgIdx+1)
	listArgs = append(listArgs, opts.PageSize, offset)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	defer rows.Close()

	var relationships []*Relationship
	for rows.Next() {
		rel := &Relationship{}
		if err := rows.Scan(&rel.ID, &rel.TenantID, &rel.SourceNodeID, &rel.TargetNodeID, &rel.RelationshipType, &rel.Data, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		relationships = append(relationships, rel)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(relationships)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return relationships, result, nil
}
