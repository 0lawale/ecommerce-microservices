package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/0lawale/shared/models"
	"github.com/0lawale/user-service/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
)

// UserService handles business logic for users
type UserService struct {
	repo      *repository.UserRepository
	jwtSecret []byte
}

// NewUserService creates a new user service
func NewUserService(repo *repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, email, password, fullName string) (*models.User, error) {
	// Validate input
	if email == "" || password == "" || fullName == "" {
		return nil, errors.New("all fields are required")
	}

	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check if email already exists
	exists, err := s.repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password (never store plain text passwords!)
	passwordHash, err := s.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Role:         "customer", // Default role
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, email, password string) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials // Don't reveal if email exists
	}

	// Verify password
	if !s.comparePassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateProfile updates user information
func (s *UserService) UpdateProfile(ctx context.Context, userID string, email, fullName string) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update fields
	if email != "" {
		user.Email = email
	}
	if fullName != "" {
		user.FullName = fullName
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ListUsers returns all users (admin only)
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}

// DeleteUser removes a user (admin only)
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ValidateToken verifies a JWT token and returns the user
func (s *UserService) ValidateToken(tokenString string) (*models.User, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Get user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user_id in token")
	}

	// Retrieve user
	return s.repo.GetByID(context.Background(), userID)
}

// HealthCheck verifies service health
func (s *UserService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

// --- Private helper methods ---

// hashPassword creates a bcrypt hash of the password
func (s *UserService) hashPassword(password string) (string, error) {
	// Cost 10 is a good balance between security and performance
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// comparePassword verifies a password against its hash
func (s *UserService) comparePassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken creates a JWT token for a user
func (s *UserService) generateToken(user *models.User) (string, int64, error) {
	// Token expires in 24 hours
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Create claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     expiresAt,
		"iat":     time.Now().Unix(), // Issued at
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt, nil
}
