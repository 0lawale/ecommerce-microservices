package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/product-service/service"
	"ecommerce/shared/models"
)

type ProductHandler struct {
	service *service.ProductService
	logger  *zap.Logger
}

func NewProductHandler(service *service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// CreateProduct creates a new product
// POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var product models.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	h.logger.Info("Creating product", zap.String("name", product.Name))

	created, err := h.service.CreateProduct(c.Request.Context(), &product)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Product created successfully",
		Data:    created,
	})
}

// GetProductByID retrieves a product by ID
// GET /api/v1/products/:id
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    product,
	})
}

// ListProducts lists products with pagination and optional category filter
// GET /api/v1/products?page=1&page_size=20&category=Electronics
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")

	products, err := h.service.ListProducts(c.Request.Context(), page, pageSize, category)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    products,
	})
}

// SearchProducts searches products by name
// GET /api/v1/products/search?q=laptop&page=1&page_size=20
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, err := h.service.SearchProducts(c.Request.Context(), query, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    products,
	})
}

// GetProductsByCategory retrieves products in a category
// GET /api/v1/products/category/:category
func (h *ProductHandler) GetProductsByCategory(c *gin.Context) {
	category := c.Param("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	products, err := h.service.GetProductsByCategory(c.Request.Context(), category, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get products by category", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    products,
	})
}

// UpdateProduct updates product information
// PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var updates models.Product
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	updated, err := h.service.UpdateProduct(c.Request.Context(), id, &updates)
	if err != nil {
		h.logger.Error("Failed to update product", zap.Error(err))
		statusCode := http.StatusInternalServerError
		if err == service.ErrProductNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Product updated successfully",
		Data:    updated,
	})
}

// UpdateStock updates product stock
// PUT /api/v1/products/:id/stock
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.service.UpdateStock(c.Request.Context(), id, req.Quantity); err != nil {
		h.logger.Error("Failed to update stock", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Stock updated successfully",
	})
}

// DeleteProduct removes a product
// DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
}

// HealthCheck returns service health
// GET /health
func (h *ProductHandler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Service:   "product-service",
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
func (h *ProductHandler) ReadinessCheck(c *gin.Context) {
	h.HealthCheck(c)
}
