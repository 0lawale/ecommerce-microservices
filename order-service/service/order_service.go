// order-service/service/order_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"ecommerce/order-service/messaging"
	"ecommerce/order-service/repository"
	"ecommerce/shared/models"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidOrder      = errors.New("invalid order data")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type OrderService struct {
	repo                 *repository.OrderRepository
	userServiceClient    *http.Client
	productServiceClient *http.Client
	publisher            *messaging.RabbitMQPublisher
	logger               *zap.Logger
}

func NewOrderService(
	repo *repository.OrderRepository,
	userClient *http.Client,
	productClient *http.Client,
	publisher *messaging.RabbitMQPublisher,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		repo:                 repo,
		userServiceClient:    userClient,
		productServiceClient: productClient,
		publisher:            publisher,
		logger:               logger,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, userID string, req *models.CreateOrderRequest) (*models.Order, error) {
	s.logger.Info("Creating order", zap.String("user_id", userID))

	// Step 1: Validate user exists
	if err := s.validateUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Step 2: Get product details - FIX HERE
	productIDs := make([]string, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := s.getProductDetails(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Step 3: Calculate total and create order items
	totalPrice := 0.0
	orderItems := make([]models.OrderItem, 0, len(req.Items))

	for _, item := range req.Items {
		product, exists := products[item.ProductID]
		if !exists {
			return nil, fmt.Errorf("product %s not found", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s: available=%d, requested=%d",
				product.Name, product.Stock, item.Quantity)
		}

		orderItem := models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
		orderItems = append(orderItems, orderItem)
		totalPrice += product.Price * float64(item.Quantity)
	}

	// Step 4: Create order
	order := &models.Order{
		UserID:     userID,
		Items:      orderItems,
		TotalPrice: totalPrice,
		Status:     "pending",
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("Order created", zap.String("order_id", order.ID))

	// Step 5: Reserve stock
	if err := s.reserveStock(ctx, order.Items); err != nil {
		s.repo.UpdateStatus(ctx, order.ID, "cancelled")
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	// Step 6: Update status
	if err := s.repo.UpdateStatus(ctx, order.ID, "confirmed"); err != nil {
		s.logger.Error("Failed to update order status", zap.Error(err))
	}

	// Step 7: Publish event
	go func() {
		event := messaging.OrderEvent{
			OrderID:    order.ID,
			UserID:     userID,
			TotalPrice: totalPrice,
			Status:     "confirmed",
			CreatedAt:  time.Now(),
		}
		if err := s.publisher.PublishOrderEvent(event); err != nil {
			s.logger.Error("Failed to publish order event", zap.Error(err))
		}
	}()

	return order, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(ctx context.Context, orderID, userID string) (*models.Order, error) {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.UserID != userID {
		return nil, errors.New("unauthorized access to order")
	}

	return order, nil
}

// ListUserOrders retrieves all orders for a user
func (s *OrderService) ListUserOrders(ctx context.Context, userID string, page, pageSize int) ([]*models.Order, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.repo.ListByUserID(ctx, userID, pageSize, offset)
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	if order.UserID != userID {
		return errors.New("unauthorized")
	}

	if order.Status == "cancelled" {
		return errors.New("order already cancelled")
	}
	if order.Status == "completed" {
		return errors.New("cannot cancel completed order")
	}

	if err := s.releaseStock(ctx, order.Items); err != nil {
		s.logger.Error("Failed to release stock", zap.Error(err))
	}

	if err := s.repo.UpdateStatus(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	go func() {
		event := messaging.OrderEvent{
			OrderID: orderID,
			UserID:  userID,
			Status:  "cancelled",
		}
		s.publisher.PublishOrderEvent(event)
	}()

	return nil
}

// GetOrderStatus retrieves order status
func (s *OrderService) GetOrderStatus(ctx context.Context, orderID, userID string) (string, error) {
	order, err := s.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return "", err
	}
	return order.Status, nil
}

// HealthCheck verifies service health
func (s *OrderService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}

// --- Helper methods ---

func (s *OrderService) validateUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	return nil
}

// FIX: Updated to accept []string instead of the items struct
func (s *OrderService) getProductDetails(ctx context.Context, productIDs []string) (map[string]*models.Product, error) {
	products := make(map[string]*models.Product)

	// In production, make actual HTTP call to Product Service
	// For now, mock data
	for _, id := range productIDs {
		products[id] = &models.Product{
			ID:    id,
			Name:  "Product " + id,
			Price: 99.99,
			Stock: 100,
		}
	}

	return products, nil
}

func (s *OrderService) reserveStock(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		s.logger.Info("Reserving stock",
			zap.String("product_id", item.ProductID),
			zap.Int("quantity", item.Quantity),
		)
	}
	return nil
}

func (s *OrderService) releaseStock(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		s.logger.Info("Releasing stock",
			zap.String("product_id", item.ProductID),
			zap.Int("quantity", item.Quantity),
		)
	}
	return nil
}

// NewHTTPClient creates HTTP client with timeout
func NewHTTPClient(baseURL string, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}
