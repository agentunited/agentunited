package repository

import (
	"context"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/jackc/pgx/v5"
)

// InviteRepository handles invite data operations
type InviteRepository interface {
	Create(ctx context.Context, invite *models.Invite, tokenHash string) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error)
	ValidateToken(ctx context.Context, tokenHash string) (*models.Invite, error)
	ConsumeToken(ctx context.Context, tokenHash string) error
}

type inviteRepository struct {
	db *DB
}

// NewInviteRepository creates a new invite repository
func NewInviteRepository(db *DB) InviteRepository {
	return &inviteRepository{db: db}
}

func (r *inviteRepository) Create(ctx context.Context, invite *models.Invite, tokenHash string) error {
	query := `
		INSERT INTO invites (id, user_id, workspace_id, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.Pool.QueryRow(ctx, query,
		invite.ID, invite.UserID, nullableUUID(invite.WorkspaceID), tokenHash, invite.Status,
		invite.ExpiresAt, invite.CreatedAt).Scan(&invite.ID, &invite.CreatedAt)
	
	if err != nil {
		return err
	}
	return nil
}

func (r *inviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Invite, error) {
	query := `
		SELECT id, user_id, COALESCE(workspace_id::text, ''), status, expires_at, created_at, consumed_at
		FROM invites
		WHERE token_hash = $1`

	var invite models.Invite
	err := r.db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&invite.ID, &invite.UserID, &invite.WorkspaceID, &invite.Status,
		&invite.ExpiresAt, &invite.CreatedAt, &invite.ConsumedAt)
	
	if err == pgx.ErrNoRows {
		return nil, models.ErrInviteNotFound
	}
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

func (r *inviteRepository) ValidateToken(ctx context.Context, tokenHash string) (*models.Invite, error) {
	invite, err := r.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(invite.ExpiresAt) {
		return nil, models.ErrInviteExpired
	}

	// Check if already consumed
	if invite.Status == models.InviteStatusConsumed {
		return nil, models.ErrInviteAlreadyConsumed
	}

	return invite, nil
}

func (r *inviteRepository) ConsumeToken(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE invites 
		SET status = 'consumed', consumed_at = NOW()
		WHERE token_hash = $1 AND status = 'pending'`

	result, err := r.db.Pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return models.ErrInviteNotFound
	}

	return nil
}

func nullableUUID(id string) any {
	if id == "" {
		return nil
	}
	return id
}