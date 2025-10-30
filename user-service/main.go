package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/shared/config"
	"ecommerce/shared/logger"
	"ecommerce/user-service/handlers"
	"ecommerce/user-service/repository"
	"ecommerce/user-service/service"
)

func main() {
	// 1. Load configuration from environment variables
	cfg := config.LoadConfig("user-service")

	// 2. Initialize logger
	log, err := logger.NewLogger(cfg.ServiceName, cfg.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting User Service",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	// 3. Initialize database connection
	db, err := repository.NewPostgresDB(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connection established")

	// 4. Run database migrations (create tables if they don't exist)
	if err := repository.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// 5. Initialize Redis for caching
	redisClient := repository.NewRedisClient(cfg.GetRedisURL(), cfg.RedisPassword)
	defer redisClient.Close()

	log.Info("Redis connection established")

	// 6. Initialize layers: Repository -> Service -> Handler
	userRepo := repository.NewUserRepository(db, redisClient)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(userService, log)

	// 7. Set up HTTP router (Gin framework)
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// 8. Register routes
	setupRoutes(router, userHandler)

	// 9. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Run server in a goroutine
	go func() {
		log.Info("Server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 10. Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

// setupRoutes configures all HTTP endpoints
func setupRoutes(router *gin.Engine, handler *handlers.UserHandler) {
	// Health check endpoint (Kubernetes uses this for liveness/readiness probes)
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handler.Register)
			auth.POST("/login", handler.Login)
		}

		// Protected routes (require JWT token)
		users := v1.Group("/users")
		users.Use(handlers.AuthMiddleware(handler))
		{
			users.GET("/me", handler.GetCurrentUser)
			users.PUT("/me", handler.UpdateProfile)
			users.GET("/:id", handler.GetUserByID)
		}

		// Admin-only routes
		admin := v1.Group("/admin")
		admin.Use(handlers.AuthMiddleware(handler), handlers.AdminMiddleware())
		{
			admin.GET("/users", handler.ListUsers)
			admin.DELETE("/users/:id", handler.DeleteUser)
		}
	}
}
