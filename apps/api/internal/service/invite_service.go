package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// InviteService handles invite operations
type InviteService struct {
	userRepo         repository.UserRepository
	inviteRepo       repository.InviteRepository
	subscriptionRepo repository.SubscriptionRepository
	jwtSecret        string
	baseURL          string
}

// NewInviteService creates a new invite service
func NewInviteService(
	userRepo repository.UserRepository,
	inviteRepo repository.InviteRepository,
	subscriptionRepo repository.SubscriptionRepository,
	jwtSecret string,
	baseURL string,
) *InviteService {
	return &InviteService{
		userRepo:         userRepo,
		inviteRepo:       inviteRepo,
		subscriptionRepo: subscriptionRepo,
		jwtSecret:        jwtSecret,
		baseURL:          baseURL,
	}
}

// ValidateInvite validates an invite token and returns invite + user info
func (s *InviteService) ValidateInvite(ctx context.Context, token string) (*models.Invite, *models.User, error) {
	tokenHash := s.hashToken(token)

	// Validate token
	invite, err := s.inviteRepo.ValidateToken(ctx, tokenHash)
	if err != nil {
		return nil, nil, err
	}

	// Get user info
	user, err := s.userRepo.GetByID(ctx, invite.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("get user: %w", err)
	}

	return invite, user, nil
}

// AcceptInvite consumes an invite token, sets user password, and optionally persists display name.
func (s *InviteService) AcceptInvite(ctx context.Context, token, password, displayName string) (string, error) {
	// Validate password strength first
	if len(password) < 12 {
		return "", models.ErrWeakPassword
	}

	tokenHash := s.hashToken(token)

	// Validate token
	invite, err := s.inviteRepo.ValidateToken(ctx, tokenHash)
	if err != nil {
		return "", err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, invite.UserID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	// Update user with password (+ optional display name)
	user.PasswordHash = string(passwordHash)
	if displayName != "" {
		user.DisplayName = displayName
	}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", fmt.Errorf("update user password/profile: %w", err)
	}

	// Consume invite token
	if err := s.inviteRepo.ConsumeToken(ctx, tokenHash); err != nil {
		return "", fmt.Errorf("consume invite token: %w", err)
	}

	// Generate JWT token
	jwtToken, err := s.generateJWTToken(user.ID, user.Email)
	if err != nil {
		return "", fmt.Errorf("generate JWT token: %w", err)
	}

	return jwtToken, nil
}

// CreateInvite creates a new invite for a human user and returns plaintext token + URL.
func (s *InviteService) CreateInvite(ctx context.Context, workspaceID, email, displayName string) (string, string, error) {
	if err := s.checkEntityLimit(ctx, workspaceID, 1); err != nil {
		return "", "", err
	}
	var userID string

	if existing, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		userID = existing.ID
	} else {
		userID = uuid.New().String()
		human := &models.User{
			ID:           userID,
			Email:        email,
			DisplayName:  displayName,
			UserType:     "human",
			PasswordHash: "",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := s.userRepo.Create(ctx, human); err != nil {
			return "", "", fmt.Errorf("create invite user: %w", err)
		}
	}

	token, tokenHash, err := s.generateInviteToken()
	if err != nil {
		return "", "", err
	}

	invite := &models.Invite{
		ID:        uuid.New().String(),
		UserID:    userID,
		Status:    models.InviteStatusPending,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	if err := s.inviteRepo.Create(ctx, invite, tokenHash); err != nil {
		return "", "", fmt.Errorf("create invite: %w", err)
	}

	return token, s.createInviteURL(token), nil
}

// hashToken creates a SHA-256 hash of the token
func (s *InviteService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

func (s *InviteService) generateInviteToken() (string, string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	token := "inv_" + base64.URLEncoding.EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token))
	return token, fmt.Sprintf("%x", hash), nil
}

func (s *InviteService) createInviteURL(token string) string {
	u, _ := url.Parse(s.baseURL)
	u.Path = "/invite"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String()
}

// generateJWTToken creates a JWT token for the user
func (s *InviteService) checkEntityLimit(ctx context.Context, workspaceID string, toAdd int64) error {
	sub, err := s.subscriptionRepo.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil // default permissive when subscription row missing
	}
	limit := int64(planEntityLimit(sub.Plan))
	if limit < 0 {
		return nil
	}
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return err
	}
	if count+toAdd > limit {
		return models.ErrEntityLimitReached
	}
	return nil
}

func planEntityLimit(plan string) int {
	switch plan {
	case "pro":
		return 15
	case "team":
		return 50
	case "enterprise":
		return -1
	default:
		return 3
	}
}

func (s *InviteService) generateJWTToken(userID, email string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
