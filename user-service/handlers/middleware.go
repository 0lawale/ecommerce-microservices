package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/0lawale/shared/models"
)

// AuthMiddleware validates JWT token and sets user in context
func AuthMiddleware(handler *UserHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		// Format: "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Split "Bearer" and token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		user, err := handler.service.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user in context for downstream handlers
		c.Set("user", user)
		c.Next()
	}
}

// AdminMiddleware checks if user has admin role
// Must be used after AuthMiddleware
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error:   "Unauthorized",
			})
			c.Abort()
			return
		}

		// Check if user is admin
		currentUser := user.(*models.User)
		if currentUser.Role != "admin" {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "Access denied: admin role required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
// Allows frontend (React/Vue) to call API from different domain
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a unique ID to each request for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new ID if not provided
			requestID = generateRequestID()
		}

		// Add to response headers
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// generateRequestID creates a simple request ID
// In production, use UUID or distributed tracing IDs (Jaeger, Zipkin)
func generateRequestID() string {
	// For simplicity, using timestamp
	// In production, use github.com/google/uuid
	return "req-" + string(rune(1000000))
}
