package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentunited/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository handles user data access
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateProfile(ctx context.Context, id, displayName, avatarURL string) error
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context) ([]*models.User, error)
}

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *models.User) error {
	var query string
	var err error

	if user.ID != "" {
		// Use provided ID (for bootstrap scenarios)
		query = `
			INSERT INTO users (id, email, display_name, avatar_url, user_type, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING created_at, updated_at
		`
		err = r.db.Pool.QueryRow(
			ctx,
			query,
			user.ID,
			user.Email,
			user.DisplayName,
			user.AvatarURL,
			defaultUserType(user.UserType),
			user.PasswordHash,
			user.CreatedAt,
			user.UpdatedAt,
		).Scan(&user.CreatedAt, &user.UpdatedAt)
	} else {
		// Let database generate ID
		query = `
			INSERT INTO users (email, display_name, avatar_url, user_type, password_hash, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`
		err = r.db.Pool.QueryRow(
			ctx,
			query,
			user.Email,
			user.DisplayName,
			user.AvatarURL,
			defaultUserType(user.UserType),
			user.PasswordHash,
			user.CreatedAt,
			user.UpdatedAt,
		).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	}

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return models.ErrEmailTaken
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email address
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), COALESCE(user_type, 'human'), password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.UserType,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), COALESCE(user_type, 'human'), password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.UserType,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// Update modifies an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET password_hash = $2,
			display_name = $3,
			avatar_url = $4,
			user_type = $5,
			updated_at = $6,
			auth_type = $7,
			central_user_id = NULLIF($8, '')
		WHERE id = $1
	`

	authType := user.AuthType
	if authType == "" {
		authType = "local"
	}
	result, err := r.db.Pool.Exec(ctx, query, user.ID, user.PasswordHash, user.DisplayName, user.AvatarURL, defaultUserType(user.UserType), user.UpdatedAt, authType, user.CentralUserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return models.ErrEmailTaken
		}
		return fmt.Errorf("update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// UpdateProfile modifies display name and avatar URL for an existing user
func (r *PostgresUserRepository) UpdateProfile(ctx context.Context, id, displayName, avatarURL string) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, id, displayName, avatarURL)
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// Count returns the total number of users
func (r *PostgresUserRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int64
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}

	return count, nil
}

// List returns all users in the instance.
func (r *PostgresUserRepository) List(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), COALESCE(user_type, 'human'), password_hash, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.UserType, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func defaultUserType(v string) string {
	if v == "agent" {
		return "agent"
	}
	return "human"
}
