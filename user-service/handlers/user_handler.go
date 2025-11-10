package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/shared/models"
	"ecommerce/user-service/service"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	service *service.UserService
	logger  *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// Register creates a new user account
// POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		FullName string `json:"full_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	h.logger.Info("User registration attempt", zap.String("email", req.Email))

	user, err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err == service.ErrEmailExists {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.logger.Info("User registered successfully", zap.String("user_id", user.ID))

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// Login authenticates a user
// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	h.logger.Info("Login attempt", zap.String("email", req.Email))

	response, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))

		statusCode := http.StatusUnauthorized
		if err == service.ErrInvalidCredentials {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.logger.Info("Login successful", zap.String("user_id", response.User.ID))

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// GetCurrentUser returns the authenticated user's info
// GET /api/v1/users/me
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// User is set by AuthMiddleware
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// GetUserByID retrieves a user by ID
// GET /api/v1/users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    user,
	})
}

// UpdateProfile updates user information
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(*models.User)

	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	updatedUser, err := h.service.UpdateProfile(
		c.Request.Context(),
		currentUser.ID,
		req.Email,
		req.FullName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    updatedUser,
	})
}

// ListUsers returns all users (admin only)
// GET /api/v1/admin/users?page=1&page_size=10
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, err := h.service.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    users,
	})
}

// DeleteUser removes a user (admin only)
// DELETE /api/v1/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// HealthCheck returns service health status
// GET /health
func (h *UserHandler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Service:   "user-service",
		Timestamp: time.Now(),
		Checks:    make(map[string]string),
	}

	// Check database connectivity
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		response.Status = "unhealthy"
		response.Checks["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	response.Checks["database"] = "connected"
	c.JSON(http.StatusOK, response)
}

// ReadinessCheck checks if service is ready to handle traffic
// GET /ready
func (h *UserHandler) ReadinessCheck(c *gin.Context) {
	// Same as health check for now, but could include additional checks
	// like cache connectivity, dependent services, etc.
	h.HealthCheck(c)
}
