package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hemanthpathath/flex-db/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB creates a connection to the test database and runs migrations
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	// Load test database configuration from environment or use defaults
	// If TEST_DB_NAME is not set, fall back to regular dbaas database
	// (tests will clean up, so be careful in production!)
	defaultDBName := "dbaas"
	if os.Getenv("TEST_DB_NAME") == "" {
		// If using default database, warn user
		t.Logf("⚠️  Using default database '%s' for tests. Set TEST_DB_NAME to use a separate test database.", defaultDBName)
	}

	cfg := db.Config{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnvInt("TEST_DB_PORT", 5432),
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", defaultDBName),
		SSLMode:  getEnv("TEST_DB_SSL_MODE", "disable"),
	}

	// Connect to test database
	pool, err := db.Connect(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return pool
}

// cleanupTestDB truncates all tables to ensure clean state between tests
func cleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Truncate all tables in reverse dependency order to avoid foreign key violations
	tables := []string{
		"relationships",
		"nodes",
		"node_types",
		"tenant_users",
		"users",
		"tenants",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := pool.Exec(ctx, query); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	// Reset sequences if they exist (optional, depends on your schema)
	resetSequences := []string{
		"ALTER SEQUENCE IF EXISTS tenants_id_seq RESTART WITH 1",
		"ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1",
	}

	for _, query := range resetSequences {
		pool.Exec(ctx, query) // Ignore errors for sequences that don't exist
	}
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns environment variable as int or default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Note: Tests should be run with a test database to avoid affecting development data
// Set TEST_DB_NAME environment variable to use a separate test database
// Example: TEST_DB_NAME=dbaas_test go test ./integration/...

