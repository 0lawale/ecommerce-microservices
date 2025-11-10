package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/order-service/service"
	"ecommerce/shared/models"
)

type OrderHandler struct {
	service *service.OrderService
	logger  *zap.Logger
}

func NewOrderHandler(service *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder creates a new order
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// In production, get userID from JWT token (AuthMiddleware)
	// For now, get from header or body
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "test-user-123" // Default for testing
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	h.logger.Info("Creating order",
		zap.String("user_id", userID),
		zap.Int("items_count", len(req.Items)),
	)

	order, err := h.service.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err == service.ErrInsufficientStock {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Order created successfully",
		Data:    order,
	})
}

// GetOrderByID retrieves an order by ID
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "test-user-123"
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), orderID, userID)
	if err != nil {
		statusCode := http.StatusNotFound
		if err.Error() == "unauthorized access to order" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    order,
	})
}

// ListUserOrders lists all orders for a user
// GET /api/v1/orders?page=1&page_size=20
func (h *OrderHandler) ListUserOrders(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "test-user-123"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, err := h.service.ListUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    orders,
	})
}

// CancelOrder cancels an order
// PUT /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "test-user-123"
	}

	if err := h.service.CancelOrder(c.Request.Context(), orderID, userID); err != nil {
		h.logger.Error("Failed to cancel order", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err.Error() == "unauthorized" {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Order cancelled successfully",
	})
}

// GetOrderStatus retrieves order status
// GET /api/v1/orders/:id/status
func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "test-user-123"
	}

	status, err := h.service.GetOrderStatus(c.Request.Context(), orderID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]string{
			"order_id": orderID,
			"status":   status,
		},
	})
}

// HealthCheck returns service health
// GET /health
func (h *OrderHandler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Service:   "order-service",
		Timestamp: time.Now(),
		Checks:    make(map[string]string),
	}

	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		response.Status = "unhealthy"
		response.Checks["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Checks["database"] = "connected"
	c.JSON(http.StatusOK, response)
}

// ReadinessCheck checks if service is ready
// GET /ready
func (h *OrderHandler) ReadinessCheck(c *gin.Context) {
	h.HealthCheck(c)
}
