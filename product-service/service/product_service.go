package service

import (
	"context"
	"errors"
	"fmt"

	"ecommerce/product-service/repository"
	"ecommerce/shared/models"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidPrice      = errors.New("price must be positive")
	ErrInvalidStock      = errors.New("stock cannot be negative")
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	// Validate input
	if product.Name == "" {
		return nil, errors.New("product name is required")
	}
	if product.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if product.Stock < 0 {
		return nil, ErrInvalidStock
	}

	// Set default category if empty
	if product.Category == "" {
		product.Category = "Uncategorized"
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// ListProducts retrieves products with pagination
func (s *ProductService) ListProducts(ctx context.Context, page, pageSize int, category string) ([]*models.Product, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset, category)
}

// SearchProducts searches products by name
func (s *ProductService) SearchProducts(ctx context.Context, query string, page, pageSize int) ([]*models.Product, error) {
	if query == "" {
		return s.ListProducts(ctx, page, pageSize, "")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.repo.SearchByName(ctx, query, pageSize, offset)
}

// GetProductsByCategory retrieves products in a category
func (s *ProductService) GetProductsByCategory(ctx context.Context, category string, page, pageSize int) ([]*models.Product, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.repo.GetByCategory(ctx, category, pageSize, offset)
}

// UpdateProduct updates product information
func (s *ProductService) UpdateProduct(ctx context.Context, id string, updates *models.Product) (*models.Product, error) {
	// Get existing product
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// Update fields
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Price > 0 {
		existing.Price = updates.Price
	}
	if updates.Stock >= 0 {
		existing.Stock = updates.Stock
	}
	if updates.Category != "" {
		existing.Category = updates.Category
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return existing, nil
}

// UpdateStock updates product stock (called by Order Service)
func (s *ProductService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	return s.repo.UpdateStock(ctx, productID, quantity)
}

// ReserveStock reserves stock for an order (decreases stock)
func (s *ProductService) ReserveStock(ctx context.Context, productID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Decrease stock (negative quantity)
	return s.repo.UpdateStock(ctx, productID, -quantity)
}

// ReleaseStock releases reserved stock (increases stock) - for cancelled orders
func (s *ProductService) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	// Increase stock (positive quantity)
	return s.repo.UpdateStock(ctx, productID, quantity)
}

// DeleteProduct removes a product
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetMultipleProducts retrieves multiple products by IDs (for order validation)
func (s *ProductService) GetMultipleProducts(ctx context.Context, ids []string) ([]*models.Product, error) {
	return s.repo.GetMultipleByIDs(ctx, ids)
}

// CheckStockAvailability checks if products have sufficient stock
func (s *ProductService) CheckStockAvailability(ctx context.Context, items map[string]int) error {
	// Get product IDs
	productIDs := make([]string, 0, len(items))
	for id := range items {
		productIDs = append(productIDs, id)
	}

	// Fetch products
	products, err := s.repo.GetMultipleByIDs(ctx, productIDs)
	if err != nil {
		return fmt.Errorf("failed to get products: %w", err)
	}

	// Build map for quick lookup
	productMap := make(map[string]*models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	// Check each item
	for productID, quantity := range items {
		product, exists := productMap[productID]
		if !exists {
			return fmt.Errorf("product %s not found", productID)
		}
		if product.Stock < quantity {
			return fmt.Errorf("insufficient stock for %s: available=%d, requested=%d",
				product.Name, product.Stock, quantity)
		}
	}

	return nil
}

// HealthCheck verifies service health
func (s *ProductService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}
