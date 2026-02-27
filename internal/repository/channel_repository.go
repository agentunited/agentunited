package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentunited/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ChannelRepository handles channel data access
type ChannelRepository interface {
	Create(ctx context.Context, channel *models.Channel) error
	GetByID(ctx context.Context, id string) (*models.Channel, error)
	ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, string, error)
	AddMember(ctx context.Context, channelID, userID, role string) error
}

// PostgresChannelRepository implements ChannelRepository with PostgreSQL
type PostgresChannelRepository struct {
	db *DB
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository(db *DB) ChannelRepository {
	return &PostgresChannelRepository{db: db}
}

// Create inserts a new channel and adds the creator as owner
func (r *PostgresChannelRepository) Create(ctx context.Context, channel *models.Channel) error {
	// Start transaction
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert channel
	query := `
		INSERT INTO channels (name, topic, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		ctx,
		query,
		channel.Name,
		channel.Topic,
		channel.CreatedBy,
		channel.CreatedAt,
		channel.UpdatedAt,
	).Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		// Check for unique constraint violation (duplicate name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return models.ErrChannelNameTaken
		}
		return fmt.Errorf("create channel: %w", err)
	}

	// Add creator as owner in channel_members
	memberQuery := `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err = tx.Exec(ctx, memberQuery, channel.ID, channel.CreatedBy, "owner")
	if err != nil {
		return fmt.Errorf("add creator as owner: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a channel by ID
func (r *PostgresChannelRepository) GetByID(ctx context.Context, id string) (*models.Channel, error) {
	query := `
		SELECT id, name, topic, created_by, created_at, updated_at
		FROM channels
		WHERE id = $1
	`

	var channel models.Channel
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Topic,
		&channel.CreatedBy,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrChannelNotFound
		}
		return nil, fmt.Errorf("get channel by id: %w", err)
	}

	return &channel, nil
}

// ListByUser retrieves all channels where the user is a member
func (r *PostgresChannelRepository) ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	query := `
		SELECT 
			c.id, 
			c.name, 
			c.topic, 
			c.created_by, 
			c.created_at, 
			c.updated_at,
			COUNT(cm2.id) as member_count
		FROM channels c
		INNER JOIN channel_members cm ON c.id = cm.channel_id
		LEFT JOIN channel_members cm2 ON c.id = cm2.channel_id
		WHERE cm.user_id = $1
		GROUP BY c.id, c.name, c.topic, c.created_by, c.created_at, c.updated_at
		ORDER BY c.created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list channels by user: %w", err)
	}
	defer rows.Close()

	var channels []*models.ChannelWithMembers
	for rows.Next() {
		var ch models.ChannelWithMembers
		err := rows.Scan(
			&ch.ID,
			&ch.Name,
			&ch.Topic,
			&ch.CreatedBy,
			&ch.CreatedAt,
			&ch.UpdatedAt,
			&ch.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, &ch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}

	return channels, nil
}

// GetMembers retrieves all members of a channel with their roles
func (r *PostgresChannelRepository) GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error) {
	query := `
		SELECT 
			u.id,
			u.email,
			cm.role
		FROM channel_members cm
		INNER JOIN users u ON cm.user_id = u.id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("get channel members: %w", err)
	}
	defer rows.Close()

	var members []*models.MemberInfo
	for rows.Next() {
		var member models.MemberInfo
		err := rows.Scan(&member.ID, &member.Email, &member.Role)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, &member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate members: %w", err)
	}

	return members, nil
}

// IsMember checks if a user is a member of a channel and returns their role
func (r *PostgresChannelRepository) IsMember(ctx context.Context, channelID, userID string) (bool, string, error) {
	query := `
		SELECT role
		FROM channel_members
		WHERE channel_id = $1 AND user_id = $2
	`

	var role string
	err := r.db.Pool.QueryRow(ctx, query, channelID, userID).Scan(&role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", nil
		}
		return false, "", fmt.Errorf("check membership: %w", err)
	}

	return true, role, nil
}

// AddMember adds a user to a channel with the specified role
func (r *PostgresChannelRepository) AddMember(ctx context.Context, channelID, userID, role string) error {
	query := `
		INSERT INTO channel_members (channel_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := r.db.Pool.Exec(ctx, query, channelID, userID, role)
	if err != nil {
		// Check for unique constraint violation (already a member)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return models.ErrAlreadyChannelMember
		}
		return fmt.Errorf("add member: %w", err)
	}

	return nil
}
