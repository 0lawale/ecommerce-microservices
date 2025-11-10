package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"ecommerce/shared/models"
)

type OrderRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewOrderRepository(db *sql.DB, redisClient *redis.Client) *OrderRepository {
	return &OrderRepository{
		db:    db,
		redis: redisClient,
	}
}

// Create inserts a new order with items (uses transaction)
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate IDs and timestamps
	order.ID = uuid.New().String()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = "pending"

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, user_id, total_price, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, orderQuery,
		order.ID, order.UserID, order.TotalPrice, order.Status,
		order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range order.Items {
		item.ID = uuid.New().String()
		item.OrderID = order.ID

		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves an order with its items
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("order:%s", id)
	cached, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var order models.Order
		if err := json.Unmarshal([]byte(cached), &order); err == nil {
			return &order, nil
		}
	}

	// Get order
	orderQuery := `
		SELECT id, user_id, total_price, status, created_at, updated_at
		FROM orders WHERE id = $1
	`
	var order models.Order
	err = r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID, &order.UserID, &order.TotalPrice, &order.Status,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("order not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	order.Items = items

	// Cache for 10 minutes
	if data, err := json.Marshal(order); err == nil {
		r.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return &order, nil
}

// ListByUserID retrieves all orders for a user
func (r *OrderRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Order, error) {
	query := `
		SELECT id, user_id, total_price, status, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalPrice, &order.Status,
			&order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Get items for this order
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateStatus updates order status
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("order:%s", orderID)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// getOrderItems retrieves items for an order (helper method)
func (r *OrderRepository) getOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items WHERE order_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// HealthCheck verifies database connectivity
func (r *OrderRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
