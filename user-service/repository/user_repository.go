package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"ecommerce/shared/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db    *sql.DB
	redis *redis.Client
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB, redisClient *redis.Client) *UserRepository {
	return &UserRepository{
		db:    db,
		redis: redisClient,
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, password_hash, full_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FullName, user.Role, user.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID with Redis caching
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	// Try cache first (reduces database load)
	cacheKey := fmt.Sprintf("user:%s", id)
	cached, err := r.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// Cache hit! Deserialize and return
		var user models.User
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	// Cache miss - query database
	query := `
		SELECT id, email, password_hash, full_name, role, created_at
		FROM users WHERE id = $1
	`

	var user models.User
	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.FullName, &user.Role, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Store in cache for 10 minutes
	if data, err := json.Marshal(user); err == nil {
		r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email (for login)
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at
		FROM users WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.FullName, &user.Role, &user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update modifies user information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $1, full_name = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.FullName, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", user.ID)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// List retrieves all users (with pagination)
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Email, &user.PasswordHash,
			&user.FullName, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user:%s", id)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	return exists, nil
}

// HealthCheck verifies database connectivity
func (r *UserRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
