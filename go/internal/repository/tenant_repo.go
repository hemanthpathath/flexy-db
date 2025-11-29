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

// ErrNotFound is returned when a resource is not found
var ErrNotFound = errors.New("not found")

// PostgresTenantRepository implements TenantRepository with PostgreSQL
type PostgresTenantRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTenantRepository creates a new PostgresTenantRepository
func NewPostgresTenantRepository(pool *pgxpool.Pool) *PostgresTenantRepository {
	return &PostgresTenantRepository{pool: pool}
}

// Create creates a new tenant
func (r *PostgresTenantRepository) Create(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	tenant.ID = uuid.New().String()
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	if tenant.Status == "" {
		tenant.Status = "active"
	}

	query := `
		INSERT INTO tenants (id, slug, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, slug, name, status, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		tenant.ID, tenant.Slug, tenant.Name, tenant.Status, tenant.CreatedAt, tenant.UpdatedAt,
	).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (r *PostgresTenantRepository) GetByID(ctx context.Context, id string) (*Tenant, error) {
	query := `SELECT id, slug, name, status, created_at, updated_at FROM tenants WHERE id = $1`

	tenant := &Tenant{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// Update updates an existing tenant
func (r *PostgresTenantRepository) Update(ctx context.Context, tenant *Tenant) (*Tenant, error) {
	tenant.UpdatedAt = time.Now()

	query := `
		UPDATE tenants 
		SET slug = $2, name = $3, status = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, slug, name, status, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		tenant.ID, tenant.Slug, tenant.Name, tenant.Status, tenant.UpdatedAt,
	).Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// Delete deletes a tenant by ID
func (r *PostgresTenantRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves tenants with pagination
func (r *PostgresTenantRepository) List(ctx context.Context, opts ListOptions) ([]*Tenant, *ListResult, error) {
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
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tenants").Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count tenants: %w", err)
	}

	query := `
		SELECT id, slug, name, status, created_at, updated_at 
		FROM tenants 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, opts.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		tenant := &Tenant{}
		if err := rows.Scan(&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(tenants)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return tenants, result, nil
}
