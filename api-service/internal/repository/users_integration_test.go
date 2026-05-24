package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(ctx context.Context, t *testing.T) *pgxpool.Pool {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate postgres: %s", err)
		}
	})

	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))

	// Create table
	_, err = pool.Exec(ctx, `
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		);
	`)
	require.NoError(t, err)

	return pool
}

func TestUserRepository_Integration(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(ctx, t)
	repo := NewUserRepository(pool)

	t.Run("Create and GetUser", func(t *testing.T) {
		name := "Test Integration"
		email := "integration@test.com"
		passwordHash := "hash123"
		role := "admin"

		// Test Create
		user, err := repo.Create(ctx, name, email, passwordHash, role)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, role, user.Role)

		// Test GetByID
		fetchedUser, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetchedUser.ID)
		assert.Equal(t, name, fetchedUser.Name)

		// Test GetByEmail
		fetchedByEmail, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetchedByEmail.ID)
		assert.Equal(t, email, fetchedByEmail.Email)
		assert.Equal(t, passwordHash, fetchedByEmail.PasswordHash) // we can assert this since get by email returns password
	})

	t.Run("GetByID Not Found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 9999)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
