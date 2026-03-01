package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentunited/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// MessageRepository handles message data access
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	GetByChannel(ctx context.Context, channelID string, limit int, before string) ([]*models.Message, bool, error)
}

// PostgresMessageRepository implements MessageRepository with PostgreSQL
type PostgresMessageRepository struct {
	db *DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *DB) MessageRepository {
	return &PostgresMessageRepository{db: db}
}

// Create inserts a new message
func (r *PostgresMessageRepository) Create(ctx context.Context, message *models.Message) error {
	query := `
		INSERT INTO messages (channel_id, author_id, author_type, text, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		message.ChannelID,
		message.AuthorID,
		message.AuthorType,
		message.Text,
		message.CreatedAt,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

// GetByChannel retrieves messages for a channel with cursor-based pagination
// Returns: messages (newest first), hasMore flag, error
func (r *PostgresMessageRepository) GetByChannel(ctx context.Context, channelID string, limit int, before string) ([]*models.Message, bool, error) {
	// Fetch one extra to determine if there are more messages
	queryLimit := limit + 1

	var query string
	var args []interface{}

	if before == "" {
		// No cursor - get latest messages
		query = `
			SELECT 
				m.id,
				m.channel_id,
				m.author_id,
				m.author_type,
				m.text,
				m.created_at,
				CASE 
					WHEN m.author_type = 'agent' THEN COALESCE(a.display_name, a.name, '')
					ELSE COALESCE(u.email, '')
				END as author_email
			FROM messages m
			LEFT JOIN users u ON m.author_id = u.id AND m.author_type = 'user'
			LEFT JOIN agents a ON m.author_id = a.id AND m.author_type = 'agent'
			WHERE m.channel_id = $1
			ORDER BY m.created_at DESC
			LIMIT $2
		`
		args = []interface{}{channelID, queryLimit}
	} else {
		// With cursor - get messages before the specified message
		query = `
			SELECT 
				m.id,
				m.channel_id,
				m.author_id,
				m.author_type,
				m.text,
				m.created_at,
				CASE 
					WHEN m.author_type = 'agent' THEN COALESCE(a.display_name, a.name, '')
					ELSE COALESCE(u.email, '')
				END as author_email
			FROM messages m
			LEFT JOIN users u ON m.author_id = u.id AND m.author_type = 'user'
			LEFT JOIN agents a ON m.author_id = a.id AND m.author_type = 'agent'
			WHERE m.channel_id = $1
			AND m.created_at < (
				SELECT created_at FROM messages WHERE id = $2
			)
			ORDER BY m.created_at DESC
			LIMIT $3
		`
		args = []interface{}{channelID, before, queryLimit}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.AuthorID,
			&msg.AuthorType,
			&msg.Text,
			&msg.CreatedAt,
			&msg.AuthorEmail,
		)
		if err != nil {
			return nil, false, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("iterate messages: %w", err)
	}

	// Check if there are more messages
	hasMore := len(messages) > limit
	if hasMore {
		// Remove the extra message
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

// GetByID retrieves a message by ID
func (r *PostgresMessageRepository) GetByID(ctx context.Context, id string) (*models.Message, error) {
	query := `
		SELECT 
			m.id,
			m.channel_id,
			m.author_id,
			m.author_type,
			m.text,
			m.created_at,
			CASE 
				WHEN m.author_type = 'agent' THEN COALESCE(a.display_name, a.name, '')
				ELSE COALESCE(u.email, '')
			END as author_email
		FROM messages m
		LEFT JOIN users u ON m.author_id = u.id AND m.author_type = 'user'
		LEFT JOIN agents a ON m.author_id = a.id AND m.author_type = 'agent'
		WHERE m.id = $1
	`

	var msg models.Message
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&msg.ID,
		&msg.ChannelID,
		&msg.AuthorID,
		&msg.AuthorType,
		&msg.Text,
		&msg.CreatedAt,
		&msg.AuthorEmail,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrMessageNotFound
		}
		return nil, fmt.Errorf("get message by id: %w", err)
	}

	return &msg, nil
}
