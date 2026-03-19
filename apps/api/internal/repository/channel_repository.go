package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ChannelRepository handles channel data access
type ChannelRepository interface {
	Create(ctx context.Context, channel *models.Channel) error
	GetByID(ctx context.Context, id string) (*models.Channel, error)
	Update(ctx context.Context, channelID, name, topic string) (*models.Channel, error)
	Delete(ctx context.Context, channelID string) error
	ListByUser(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	GetMembers(ctx context.Context, channelID string) ([]*models.MemberInfo, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, string, error)
	AddMember(ctx context.Context, channelID, userID, role string) error
	RemoveMember(ctx context.Context, channelID, userID string) error
	ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error)
	GetOrCreateDMChannel(ctx context.Context, user1ID, user2ID string) (*models.Channel, error)
	MarkChannelRead(ctx context.Context, userID, channelID string) error
	GetUnreadCounts(ctx context.Context, userID string) (map[string]int, error)
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
	var query string
	if channel.ID != "" {
		// Use provided ID (for bootstrap scenarios)
		query = `
			INSERT INTO channels (id, name, topic, type, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING created_at, updated_at
		`
		err = tx.QueryRow(
			ctx,
			query,
			channel.ID,
			channel.Name,
			channel.Topic,
			channel.Type,
			channel.CreatedBy,
			channel.CreatedAt,
			channel.UpdatedAt,
		).Scan(&channel.CreatedAt, &channel.UpdatedAt)
	} else {
		// Let database generate ID
		query = `
			INSERT INTO channels (name, topic, type, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at
		`
		err = tx.QueryRow(
			ctx,
			query,
			channel.Name,
			channel.Topic,
			channel.Type,
			channel.CreatedBy,
			channel.CreatedAt,
			channel.UpdatedAt,
		).Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)
	}

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
		SELECT id, name, topic, type, created_by, created_at, updated_at
		FROM channels
		WHERE id = $1
	`

	var channel models.Channel
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Topic,
		&channel.Type,
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
			c.type, 
			c.created_by, 
			c.created_at, 
			c.updated_at,
			COUNT(cm2.id) as member_count,
			COALESCE((
				SELECT COUNT(*)
				FROM messages m
				LEFT JOIN user_channel_reads ucr ON ucr.user_id = $1 AND ucr.channel_id = m.channel_id
				WHERE m.channel_id = c.id
				  AND m.author_id <> $1
				  AND m.created_at > COALESCE(ucr.last_read_at, to_timestamp(0))
			), 0) as unread_count
		FROM channels c
		INNER JOIN channel_members cm ON c.id = cm.channel_id
		LEFT JOIN channel_members cm2 ON c.id = cm2.channel_id
		WHERE cm.user_id = $1 AND c.type = 'channel'
		GROUP BY c.id, c.name, c.topic, c.type, c.created_by, c.created_at, c.updated_at
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
			&ch.Type,
			&ch.CreatedBy,
			&ch.CreatedAt,
			&ch.UpdatedAt,
			&ch.MemberCount,
			&ch.UnreadCount,
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
			COALESCE(NULLIF(u.display_name, ''), split_part(u.email, '@', 1), u.email) AS display_name,
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

// Update modifies a channel's name and topic
func (r *PostgresChannelRepository) Update(ctx context.Context, channelID, name, topic string) (*models.Channel, error) {
	query := `
		UPDATE channels 
		SET name = $2, topic = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, topic, type, created_by, created_at, updated_at
	`

	var channel models.Channel
	err := r.db.Pool.QueryRow(ctx, query, channelID, name, topic).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Topic,
		&channel.Type,
		&channel.CreatedBy,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrChannelNotFound
		}
		// Check for unique constraint violation (duplicate name)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, models.ErrChannelNameTaken
		}
		return nil, fmt.Errorf("update channel: %w", err)
	}

	return &channel, nil
}

// Delete removes a channel and all associated data
func (r *PostgresChannelRepository) Delete(ctx context.Context, channelID string) error {
	// Start transaction
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete channel (cascade will delete members and messages)
	result, err := tx.Exec(ctx, "DELETE FROM channels WHERE id = $1", channelID)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrChannelNotFound
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// RemoveMember removes a user from a channel
func (r *PostgresChannelRepository) RemoveMember(ctx context.Context, channelID, userID string) error {
	query := `
		DELETE FROM channel_members 
		WHERE channel_id = $1 AND user_id = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, channelID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrNotChannelMember
	}

	return nil
}

// ListDMChannels retrieves all DM channels for a user
func (r *PostgresChannelRepository) ListDMChannels(ctx context.Context, userID string) ([]*models.ChannelWithMembers, error) {
	query := `
		SELECT 
			c.id, 
			c.name, 
			c.topic,
			c.type, 
			c.created_by, 
			c.created_at, 
			c.updated_at,
			COUNT(cm2.id) as member_count,
			COALESCE(NULLIF(other_user.display_name, ''), split_part(other_user.email, '@', 1), other_agent.display_name, '') as other_participant,
			COALESCE((
				SELECT COUNT(*)
				FROM messages m
				LEFT JOIN user_channel_reads ucr ON ucr.user_id = $1 AND ucr.channel_id = m.channel_id
				WHERE m.channel_id = c.id
				  AND m.author_id <> $1
				  AND m.created_at > COALESCE(ucr.last_read_at, to_timestamp(0))
			), 0) as unread_count
		FROM channels c
		INNER JOIN channel_members cm ON c.id = cm.channel_id
		LEFT JOIN channel_members cm2 ON c.id = cm2.channel_id
		LEFT JOIN channel_members cm_other ON c.id = cm_other.channel_id AND cm_other.user_id != $1
		LEFT JOIN users other_user ON cm_other.user_id = other_user.id
		LEFT JOIN agents other_agent ON cm_other.user_id = other_agent.owner_id
		WHERE cm.user_id = $1 AND c.type = 'dm'
		GROUP BY c.id, c.name, c.topic, c.type, c.created_by, c.created_at, c.updated_at, other_user.display_name, other_user.email, other_agent.display_name
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list dm channels: %w", err)
	}
	defer rows.Close()

	var channels []*models.ChannelWithMembers
	for rows.Next() {
		var ch models.ChannelWithMembers
		var otherParticipant string
		err := rows.Scan(
			&ch.ID,
			&ch.Name,
			&ch.Topic,
			&ch.Type,
			&ch.CreatedBy,
			&ch.CreatedAt,
			&ch.UpdatedAt,
			&ch.MemberCount,
			&otherParticipant,
			&ch.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dm channel: %w", err)
		}
		if otherParticipant != "" {
			ch.Name = otherParticipant
		}
		channels = append(channels, &ch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dm channels: %w", err)
	}

	return channels, nil
}

// GetOrCreateDMChannel gets existing DM channel between two users or creates one
func (r *PostgresChannelRepository) GetOrCreateDMChannel(ctx context.Context, user1ID, user2ID string) (*models.Channel, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	findQuery := `
		SELECT c.id, c.name, c.topic, c.type, c.created_by, c.created_at, c.updated_at
		FROM channels c
		WHERE c.type = 'dm'
		AND EXISTS (SELECT 1 FROM channel_members cm1 WHERE cm1.channel_id = c.id AND cm1.user_id = $1)
		AND EXISTS (SELECT 1 FROM channel_members cm2 WHERE cm2.channel_id = c.id AND cm2.user_id = $2)
		AND (SELECT COUNT(*) FROM channel_members cm WHERE cm.channel_id = c.id) = 2
		LIMIT 1
	`

	var channel models.Channel
	err = tx.QueryRow(ctx, findQuery, user1ID, user2ID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Topic,
		&channel.Type,
		&channel.CreatedBy,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit transaction: %w", err)
		}
		return &channel, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("find existing dm channel: %w", err)
	}

	now := time.Now()
	channel = models.Channel{
		Name:      fmt.Sprintf("dm-%s-%s", user1ID[:8], user2ID[:8]),
		Topic:     "",
		Type:      "dm",
		CreatedBy: user1ID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	insertQuery := `
		INSERT INTO channels (name, topic, type, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(ctx, insertQuery, channel.Name, channel.Topic, channel.Type, channel.CreatedBy, channel.CreatedAt, channel.UpdatedAt).
		Scan(&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create dm channel: %w", err)
	}

	memberQuery := `INSERT INTO channel_members (channel_id, user_id, role, joined_at) VALUES ($1, $2, $3, NOW())`
	if _, err := tx.Exec(ctx, memberQuery, channel.ID, user1ID, "member"); err != nil {
		return nil, fmt.Errorf("add first member to dm channel: %w", err)
	}
	if _, err := tx.Exec(ctx, memberQuery, channel.ID, user2ID, "member"); err != nil {
		return nil, fmt.Errorf("add second member to dm channel: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return &channel, nil
}

// MarkChannelRead upserts the last_read_at timestamp for a user/channel pair.
func (r *PostgresChannelRepository) MarkChannelRead(ctx context.Context, userID, channelID string) error {
	query := `
		INSERT INTO user_channel_reads (user_id, channel_id, last_read_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, channel_id)
		DO UPDATE SET last_read_at = EXCLUDED.last_read_at
	`
	if _, err := r.db.Pool.Exec(ctx, query, userID, channelID); err != nil {
		return fmt.Errorf("mark channel read: %w", err)
	}
	return nil
}

// GetUnreadCounts returns per-channel unread counts for the given user.
func (r *PostgresChannelRepository) GetUnreadCounts(ctx context.Context, userID string) (map[string]int, error) {
	query := `
		SELECT
			m.channel_id,
			COUNT(*)::int AS unread_count
		FROM messages m
		INNER JOIN channel_members cm ON cm.channel_id = m.channel_id AND cm.user_id = $1
		LEFT JOIN user_channel_reads ucr ON ucr.user_id = $1 AND ucr.channel_id = m.channel_id
		WHERE m.author_id <> $1
		  AND m.created_at > COALESCE(ucr.last_read_at, to_timestamp(0))
		GROUP BY m.channel_id
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get unread counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var channelID string
		var count int
		if err := rows.Scan(&channelID, &count); err != nil {
			return nil, fmt.Errorf("scan unread count: %w", err)
		}
		counts[channelID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate unread counts: %w", err)
	}

	return counts, nil
}
