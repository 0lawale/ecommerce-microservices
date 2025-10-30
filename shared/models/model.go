package models

import "time"

// Product represents an item in the catalog
type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Stock       int       `json:"stock" db:"stock"`
	Category    string    `json:"category" db:"category"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a system user
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // "-" means never serialize to JSON
	FullName     string    `json:"full_name" db:"full_name"`
	Role         string    `json:"role" db:"role"` // "admin" or "customer"
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Order represents a customer order
type Order struct {
	ID         string      `json:"id" db:"id"`
	UserID     string      `json:"user_id" db:"user_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price" db:"total_price"`
	Status     string      `json:"status" db:"status"` // "pending", "completed", "cancelled"
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// OrderItem represents a product in an order
type OrderItem struct {
	ID        string  `json:"id" db:"id"`
	OrderID   string  `json:"order_id" db:"order_id"`
	ProductID string  `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"` // Price at time of order
}

// Notification represents a notification to be sent
type Notification struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Type      string    `json:"type" db:"type"` // "email", "sms"
	Subject   string    `json:"subject" db:"subject"`
	Message   string    `json:"message" db:"message"`
	Status    string    `json:"status" db:"status"` // "pending", "sent", "failed"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// APIResponse is the standard response structure for all APIs
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthCheckResponse for Kubernetes liveness/readiness probes
type HealthCheckResponse struct {
	Status    string            `json:"status"` // "healthy" or "unhealthy"
	Service   string            `json:"service"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"` // e.g., {"database": "connected", "redis": "connected"}
}

// LoginRequest for user authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse contains JWT token
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      User   `json:"user"`
}

// CreateOrderRequest for placing orders
type CreateOrderRequest struct {
	Items []struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
}
