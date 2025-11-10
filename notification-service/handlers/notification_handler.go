package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/notification-service/service"
	"ecommerce/shared/models"
)

type NotificationHandler struct {
	service *service.NotificationService
	logger  *zap.Logger
}

func NewNotificationHandler(svc *service.NotificationService, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: svc,
		logger:  log,
	}
}

func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	userID := c.Param("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	offset := (page - 1) * pageSize
	notifications, err := h.service.GetUserNotifications(c.Request.Context(), userID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    notifications,
	})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.MarkAsRead(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Notification marked as read",
	})
}

func (h *NotificationHandler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "healthy",
		Service:   "notification-service",
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

func (h *NotificationHandler) ReadinessCheck(c *gin.Context) {
	h.HealthCheck(c)
}
