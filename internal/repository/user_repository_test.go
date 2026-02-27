package repository

import (
	"context"
	"testing"
	"time"

	"github.com/agentunited/backend/internal/config"
	"github.com/agentunited/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *DB {
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		Database: "agentunited_test",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	db, err := NewDB(ctx, cfg)
	require.NoError(t, err, "Failed to connect to test database")

	// Drop all tables and recreate schema (clean slate for tests)
	_, err = db.Pool.Exec(ctx, "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;")
	require.NoError(t, err, "Failed to reset schema")

	// Run migrations
	err = db.RunMigrations(ctx, "../../migrations")
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T, db *DB) {
	ctx := context.Background()
	// Truncate all tables in dependency order
	_, err := db.Pool.Exec(ctx, "TRUNCATE TABLE channel_members, channels, users CASCADE")
	require.NoError(t, err, "Failed to truncate tables")
	db.Close()
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify ID was assigned
	assert.NotEmpty(t, user.ID)

	// Verify timestamps were set
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create first user
	user1 := &models.User{
		Email:        "duplicate@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same email
	user2 := &models.User{
		Email:        "duplicate@example.com",
		PasswordHash: "differenthash456",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = repo.Create(ctx, user2)

	// Should return ErrEmailTaken
	assert.ErrorIs(t, err, models.ErrEmailTaken)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	originalUser := &models.User{
		Email:        "getbyemail@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Retrieve by email
	retrievedUser, err := repo.GetByEmail(ctx, "getbyemail@example.com")
	require.NoError(t, err)

	// Verify fields match
	assert.Equal(t, originalUser.ID, retrievedUser.ID)
	assert.Equal(t, originalUser.Email, retrievedUser.Email)
	assert.Equal(t, originalUser.PasswordHash, retrievedUser.PasswordHash)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get non-existent user
	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")

	// Should return ErrUserNotFound
	assert.ErrorIs(t, err, models.ErrUserNotFound)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	originalUser := &models.User{
		Email:        "getbyid@example.com",
		PasswordHash: "hashedpassword123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Retrieve by ID
	retrievedUser, err := repo.GetByID(ctx, originalUser.ID)
	require.NoError(t, err)

	// Verify fields match
	assert.Equal(t, originalUser.ID, retrievedUser.ID)
	assert.Equal(t, originalUser.Email, retrievedUser.Email)
	assert.Equal(t, originalUser.PasswordHash, retrievedUser.PasswordHash)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Try to get non-existent user
	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")

	// Should return ErrUserNotFound
	assert.ErrorIs(t, err, models.ErrUserNotFound)
}
