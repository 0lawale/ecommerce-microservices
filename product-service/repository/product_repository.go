// product-service/repository/product_repository.go
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"ecommerce/shared/models"
)

type ProductRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewProductRepository(db *sql.DB, redisClient *redis.Client) *ProductRepository {
	return &ProductRepository{
		db:    db,
		redis: redisClient,
	}
}

// Create inserts a new product
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	product.ID = uuid.New().String()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	query := `
		INSERT INTO products (id, name, description, price, stock, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Description, product.Price,
		product.Stock, product.Category, product.CreatedAt, product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID with caching
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%s", id)
	cached, err := r.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cached), &product); err == nil {
			return &product, nil
		}
	}

	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products WHERE id = $1
	`

	var product models.Product
	err = r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if data, err := json.Marshal(product); err == nil {
		r.redis.Set(ctx, cacheKey, data, 30*time.Minute)
	}

	return &product, nil
}

// List retrieves products with pagination and filters
func (r *ProductRepository) List(ctx context.Context, limit, offset int, category string) ([]*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
	`
	args := []interface{}{}
	argPosition := 1

	if category != "" {
		query += fmt.Sprintf(" WHERE category = $%d", argPosition)
		args = append(args, category)
		argPosition++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

// SearchByName searches products by name
func (r *ProductRepository) SearchByName(ctx context.Context, searchTerm string, limit, offset int) ([]*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
		WHERE LOWER(name) LIKE LOWER($1) OR LOWER(description) LIKE LOWER($1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + searchTerm + "%"

	rows, err := r.db.QueryContext(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

// GetByCategory retrieves products by category
func (r *ProductRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*models.Product, error) {
	return r.List(ctx, limit, offset, category)
}

// Update modifies product information
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock = $4, category = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		product.Name, product.Description, product.Price, product.Stock,
		product.Category, product.UpdatedAt, product.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	cacheKey := fmt.Sprintf("product:%s", product.ID)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// UpdateStock updates product stock with transaction
func (r *ProductRepository) UpdateStock(ctx context.Context, productID string, quantity int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var currentStock int
	query := `SELECT stock FROM products WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, productID).Scan(&currentStock)
	if err == sql.ErrNoRows {
		return fmt.Errorf("product not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get stock: %w", err)
	}

	newStock := currentStock + quantity
	if newStock < 0 {
		return fmt.Errorf("insufficient stock: current=%d, requested=%d", currentStock, -quantity)
	}

	updateQuery := `UPDATE products SET stock = $1, updated_at = $2 WHERE id = $3`
	_, err = tx.ExecContext(ctx, updateQuery, newStock, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	cacheKey := fmt.Sprintf("product:%s", productID)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// Delete removes a product
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	cacheKey := fmt.Sprintf("product:%s", id)
	r.redis.Del(ctx, cacheKey)

	return nil
}

// GetMultipleByIDs retrieves multiple products
func (r *ProductRepository) GetMultipleByIDs(ctx context.Context, ids []string) ([]*models.Product, error) {
	if len(ids) == 0 {
		return []*models.Product{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, price, stock, category, created_at, updated_at
		FROM products
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.Stock, &product.Category, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

// HealthCheck verifies database connectivity
func (r *ProductRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
