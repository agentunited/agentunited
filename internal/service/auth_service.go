package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

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

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
