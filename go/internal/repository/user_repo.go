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

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create creates a new user
func (r *PostgresUserRepository) Create(ctx context.Context, user *User) (*User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, display_name, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.DisplayName, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, email, display_name, created_at, updated_at FROM users WHERE id = $1`

	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *User) (*User, error) {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET email = $2, display_name = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, email, display_name, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query,
		user.ID, user.Email, user.DisplayName, user.UpdatedAt,
	).Scan(&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Delete deletes a user by ID
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// List retrieves users with pagination
func (r *PostgresUserRepository) List(ctx context.Context, opts ListOptions) ([]*User, *ListResult, error) {
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
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count users: %w", err)
	}

	query := `
		SELECT id, email, display_name, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, opts.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Email, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(users)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return users, result, nil
}

// AddToTenant adds a user to a tenant
func (r *PostgresUserRepository) AddToTenant(ctx context.Context, tenantUser *TenantUser) (*TenantUser, error) {
	if tenantUser.Role == "" {
		tenantUser.Role = "member"
	}
	if tenantUser.Status == "" {
		tenantUser.Status = "active"
	}

	query := `
		INSERT INTO tenant_users (tenant_id, user_id, role, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = $3, status = $4
		RETURNING tenant_id, user_id, role, status
	`

	err := r.pool.QueryRow(ctx, query,
		tenantUser.TenantID, tenantUser.UserID, tenantUser.Role, tenantUser.Status,
	).Scan(&tenantUser.TenantID, &tenantUser.UserID, &tenantUser.Role, &tenantUser.Status)

	if err != nil {
		return nil, fmt.Errorf("failed to add user to tenant: %w", err)
	}

	return tenantUser, nil
}

// RemoveFromTenant removes a user from a tenant
func (r *PostgresUserRepository) RemoveFromTenant(ctx context.Context, tenantID, userID string) error {
	query := `DELETE FROM tenant_users WHERE tenant_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// ListTenantUsers lists users in a tenant
func (r *PostgresUserRepository) ListTenantUsers(ctx context.Context, tenantID string, opts ListOptions) ([]*TenantUser, *ListResult, error) {
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
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM tenant_users WHERE tenant_id = $1", tenantID).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count tenant users: %w", err)
	}

	query := `
		SELECT tenant_id, user_id, role, status 
		FROM tenant_users 
		WHERE tenant_id = $1
		ORDER BY user_id
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, tenantID, opts.PageSize, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tenant users: %w", err)
	}
	defer rows.Close()

	var tenantUsers []*TenantUser
	for rows.Next() {
		tu := &TenantUser{}
		if err := rows.Scan(&tu.TenantID, &tu.UserID, &tu.Role, &tu.Status); err != nil {
			return nil, nil, fmt.Errorf("failed to scan tenant user: %w", err)
		}
		tenantUsers = append(tenantUsers, tu)
	}

	result := &ListResult{TotalCount: totalCount}
	nextOffset := offset + len(tenantUsers)
	if nextOffset < totalCount {
		result.NextPageToken = strconv.Itoa(nextOffset)
	}

	return tenantUsers, result, nil
}
