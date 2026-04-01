package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/agentunited/backend/internal/email"
	"github.com/agentunited/backend/internal/models"
	"github.com/agentunited/backend/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
)

// emailRegex validates email format
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// AuthService handles authentication operations
type AuthService interface {
	Register(ctx context.Context, email, password string) (*models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetCurrentUser(ctx context.Context, userID string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID, displayName, avatarURL string) (*models.User, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ForgotPassword(ctx context.Context, emailAddr string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// authService implements AuthService
type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account
func (s *authService) Register(ctx context.Context, email, password string) (*models.User, error) {
	// Validate email format
	if !isValidEmail(email) {
		return nil, models.ErrInvalidEmail
	}

	// Validate password strength
	if !isStrongPassword(password) {
		return nil, models.ErrWeakPassword
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return nil, fmt.Errorf("check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, models.ErrEmailTaken
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Clear password hash before returning (security)
	user.PasswordHash = ""

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			// Don't reveal whether user exists - return generic error
			return "", models.ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user by email: %w", err)
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Password doesn't match - return generic error
		return "", models.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("generate JWT: %w", err)
	}

	return token, nil
}

// generateJWT creates a JWT token for the given user
func (s *authService) generateJWT(user *models.User) (string, error) {
	// Set expiration to 24 hours from now
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := &JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}

// isValidEmail checks if email format is valid
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

// isStrongPassword checks if password meets strength requirements
// Requirements: min 8 characters, at least 1 letter and 1 number
func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
		if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	return hasLetter && hasNumber
}

// GetCurrentUser returns the current user profile
func (s *authService) GetCurrentUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// UpdateProfile updates display name and avatar URL for a user
func (s *authService) UpdateProfile(ctx context.Context, userID, displayName, avatarURL string) (*models.User, error) {
	if err := s.userRepo.UpdateProfile(ctx, userID, displayName, avatarURL); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}

// ChangePassword updates a user's password after validating current password
func (s *authService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if !isStrongPassword(newPassword) {
		return models.ErrWeakPassword
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.ErrInvalidCredentials
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return models.ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ForgotPassword generates a reset token, stores it, and emails the user.
// Always returns nil (200) — never leaks whether an account exists.
func (s *authService) ForgotPassword(ctx context.Context, emailAddr string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	if emailAddr == "" {
		return nil
	}

	user, err := s.userRepo.GetByEmail(ctx, emailAddr)
	if err != nil || user == nil {
		// No-op — don't reveal account existence.
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		return fmt.Errorf("generate reset token: %w", err)
	}

	if err := s.userRepo.CreatePasswordResetToken(ctx, token, user.ID); err != nil {
		return fmt.Errorf("store reset token: %w", err)
	}

	resetLink := "https://app.agentunited.ai/reset-password?token=" + token
	subject := "Reset your Agent United password"
	body := fmt.Sprintf(
		"Click to reset your password: %s\n\nThis link expires in 1 hour. If you didn't request a reset, ignore this email.",
		resetLink,
	)
	email.Send(ctx, user.Email, subject, body)
	return nil
}

// ResetPassword validates the token and updates the user's password.
func (s *authService) ResetPassword(ctx context.Context, token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return models.ErrInvalidOrExpiredToken
	}
	if len(newPassword) < 8 {
		return models.ErrWeakPassword
	}

	userID, err := s.userRepo.GetUserIDByResetToken(ctx, token)
	if err != nil || userID == "" {
		return models.ErrInvalidOrExpiredToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Single-use: invalidate token immediately.
	if err := s.userRepo.DeletePasswordResetToken(ctx, token); err != nil {
		// Non-fatal — token will expire naturally.
		fmt.Printf("warn: delete reset token: %v\n", err)
	}
	return nil
}

// generateResetToken generates a 32-byte cryptographically random hex token.
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
