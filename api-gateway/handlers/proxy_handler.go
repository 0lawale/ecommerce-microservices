package handlers

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/shared/models"
)

type ProxyHandler struct {
	userServiceURL    string
	productServiceURL string
	orderServiceURL   string
	logger            *zap.Logger
	httpClient        *http.Client
}

func NewProxyHandler(userURL, productURL, orderURL string, logger *zap.Logger) *ProxyHandler {
	return &ProxyHandler{
		userServiceURL:    userURL,
		productServiceURL: productURL,
		orderServiceURL:   orderURL,
		logger:            logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToUserService forwards requests to User Service
func (h *ProxyHandler) ProxyToUserService(c *gin.Context) {
	h.proxyRequest(c, h.userServiceURL, "user-service")
}

// ProxyToProductService forwards requests to Product Service
func (h *ProxyHandler) ProxyToProductService(c *gin.Context) {
	h.proxyRequest(c, h.productServiceURL, "product-service")
}

// ProxyToOrderService forwards requests to Order Service
func (h *ProxyHandler) ProxyToOrderService(c *gin.Context) {
	h.proxyRequest(c, h.orderServiceURL, "order-service")
}

// proxyRequest is the core proxy logic
func (h *ProxyHandler) proxyRequest(c *gin.Context, targetBaseURL, serviceName string) {
	startTime := time.Now()

	// Build target URL
	targetURL := targetBaseURL + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	h.logger.Info("Proxying request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("target_service", serviceName),
		zap.String("target_url", targetURL),
	)

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create proxy request
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		h.logger.Error("Failed to create proxy request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to proxy request",
		})
		return
	}

	// Copy headers (important for authentication)
	for key, values := range c.Request.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Add request tracking header
	proxyReq.Header.Set("X-Gateway-Request-ID", c.GetString("request_id"))

	// Execute proxy request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.logger.Error("Proxy request failed",
			zap.Error(err),
			zap.String("service", serviceName),
		)
		c.JSON(http.StatusServiceUnavailable, models.APIResponse{
			Success: false,
			Error:   "Service unavailable: " + serviceName,
		})
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   "Failed to read service response",
		})
		return
	}

	// Log response time
	duration := time.Since(startTime)
	h.logger.Info("Request proxied successfully",
		zap.String("service", serviceName),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", duration),
	)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// HealthCheck checks gateway and all backend services
func (h *ProxyHandler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Service:   "api-gateway",
		Timestamp: time.Now(),
		Checks:    make(map[string]string),
	}

	// Check all backend services
	services := map[string]string{
		"user-service":    h.userServiceURL + "/health",
		"product-service": h.productServiceURL + "/health",
		"order-service":   h.orderServiceURL + "/health",
	}

	allHealthy := true
	for name, url := range services {
		resp, err := h.httpClient.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			response.Checks[name] = "unhealthy"
			allHealthy = false
		} else {
			response.Checks[name] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if !allHealthy {
		response.Status = "degraded"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReadinessCheck checks if gateway is ready
func (h *ProxyHandler) ReadinessCheck(c *gin.Context) {
	// For gateway, readiness is same as health
	h.HealthCheck(c)
}
