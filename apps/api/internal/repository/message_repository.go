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
	GetByID(ctx context.Context, messageID string) (*models.Message, error)
	Update(ctx context.Context, messageID, text string) (*models.Message, error)
	Delete(ctx context.Context, messageID string) error
	GetByChannel(ctx context.Context, channelID string, limit int, before string) ([]*models.Message, bool, error)
	Search(ctx context.Context, query string, channelID string, limit int) ([]*models.Message, error)
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
		INSERT INTO messages (channel_id, author_id, author_type, text, attachment_url, attachment_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	// Handle NULL values for attachment fields
	var attachmentURL, attachmentName interface{}
	if message.AttachmentURL != "" {
		attachmentURL = message.AttachmentURL
	}
	if message.AttachmentName != "" {
		attachmentName = message.AttachmentName
	}

	err := r.db.Pool.QueryRow(
		ctx,
		query,
		message.ChannelID,
		message.AuthorID,
		message.AuthorType,
		message.Text,
		attachmentURL,
		attachmentName,
		message.CreatedAt,
		message.CreatedAt, // Set updated_at to created_at initially
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

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
				m.attachment_url,
				m.attachment_name,
				m.created_at,
				m.updated_at,
				CASE 
					WHEN m.author_type = 'agent' THEN COALESCE(NULLIF(a.display_name, ''), NULLIF(a.name, ''), '')
					ELSE COALESCE(NULLIF(u.display_name, ''), split_part(u.email, '@', 1), '')
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
				m.attachment_url,
				m.attachment_name,
				m.created_at,
				m.updated_at,
				CASE 
					WHEN m.author_type = 'agent' THEN COALESCE(NULLIF(a.display_name, ''), NULLIF(a.name, ''), '')
					ELSE COALESCE(NULLIF(u.display_name, ''), split_part(u.email, '@', 1), '')
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
		var attachmentURL, attachmentName *string
		err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.AuthorID,
			&msg.AuthorType,
			&msg.Text,
			&attachmentURL,
			&attachmentName,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.AuthorEmail,
		)

		// Handle NULL attachment fields
		if attachmentURL != nil {
			msg.AttachmentURL = *attachmentURL
		}
		if attachmentName != nil {
			msg.AttachmentName = *attachmentName
		}
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
			m.attachment_url,
			m.attachment_name,
			m.created_at,
			m.updated_at,
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
	var attachmentURL, attachmentName *string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&msg.ID,
		&msg.ChannelID,
		&msg.AuthorID,
		&msg.AuthorType,
		&msg.Text,
		&attachmentURL,
		&attachmentName,
		&msg.CreatedAt,
		&msg.UpdatedAt,
		&msg.AuthorEmail,
	)

	// Handle NULL attachment fields
	if attachmentURL != nil {
		msg.AttachmentURL = *attachmentURL
	}
	if attachmentName != nil {
		msg.AttachmentName = *attachmentName
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrMessageNotFound
		}
		return nil, fmt.Errorf("get message by id: %w", err)
	}

	return &msg, nil
}

// Update modifies a message's text content
func (r *PostgresMessageRepository) Update(ctx context.Context, messageID, text string) (*models.Message, error) {
	query := `
		UPDATE messages
		SET text = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, channel_id, author_id, author_type, text, created_at, updated_at
	`

	var msg models.Message
	err := r.db.Pool.QueryRow(ctx, query, messageID, text).Scan(
		&msg.ID,
		&msg.ChannelID,
		&msg.AuthorID,
		&msg.AuthorType,
		&msg.Text,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrMessageNotFound
		}
		return nil, fmt.Errorf("update message: %w", err)
	}

	return &msg, nil
}

// Delete removes a message by ID
func (r *PostgresMessageRepository) Delete(ctx context.Context, messageID string) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrMessageNotFound
	}

	return nil
}

// Search performs full-text search across messages
func (r *PostgresMessageRepository) Search(ctx context.Context, query string, channelID string, limit int) ([]*models.Message, error) {
	// Build the SQL query
	sqlQuery := `
		SELECT 
			m.id,
			m.channel_id,
			m.author_id,
			m.author_type,
			m.text,
			m.attachment_url,
			m.attachment_name,
			m.created_at,
			m.updated_at,
			CASE 
				WHEN m.author_type = 'agent' THEN COALESCE(a.display_name, a.name, '')
				ELSE COALESCE(u.email, '')
			END as author_email,
			ts_rank(m.search_vector, plainto_tsquery('english', $1)) as rank
		FROM messages m
		LEFT JOIN users u ON m.author_id = u.id AND m.author_type = 'user'
		LEFT JOIN agents a ON m.author_id = a.id AND m.author_type = 'agent'
		WHERE m.search_vector @@ plainto_tsquery('english', $1)
	`

	args := []interface{}{query}
	argIndex := 2

	// Add channel filter if specified
	if channelID != "" {
		sqlQuery += fmt.Sprintf(" AND m.channel_id = $%d", argIndex)
		args = append(args, channelID)
		argIndex++
	}

	// Order by relevance (rank) and recency
	sqlQuery += ` ORDER BY rank DESC, m.created_at DESC`

	// Add limit
	if limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.db.Pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		var rank float32 // We don't use this in the model, but need to scan it

		var attachmentURL, attachmentName *string
		err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.AuthorID,
			&msg.AuthorType,
			&msg.Text,
			&attachmentURL,
			&attachmentName,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.AuthorEmail,
			&rank,
		)

		// Handle NULL attachment fields
		if attachmentURL != nil {
			msg.AttachmentURL = *attachmentURL
		}
		if attachmentName != nil {
			msg.AttachmentName = *attachmentName
		}
		if err != nil {
			return nil, fmt.Errorf("scan search result: %w", err)
		}

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return messages, nil
}
