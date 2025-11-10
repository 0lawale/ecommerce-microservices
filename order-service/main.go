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

	"ecommerce/order-service/handlers"
	"ecommerce/order-service/messaging"
	"ecommerce/order-service/repository"
	"ecommerce/order-service/service"
	"ecommerce/shared/config"
	"ecommerce/shared/logger"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig("order-service")

	// 2. Initialize logger
	log, err := logger.NewLogger(cfg.ServiceName, cfg.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Order Service",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	// 3. Initialize database
	db, err := repository.NewPostgresDB(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connection established")

	// 4. Run migrations
	if err := repository.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// 5. Initialize Redis
	redisClient := repository.NewRedisClient(cfg.GetRedisURL(), cfg.RedisPassword)
	defer redisClient.Close()

	log.Info("Redis connection established")

	// 6. Initialize RabbitMQ publisher
	publisher, err := messaging.NewRabbitMQPublisher(cfg.RabbitMQURL, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer publisher.Close()

	log.Info("RabbitMQ connection established")

	// 7. Initialize HTTP clients for inter-service communication
	userServiceClient := service.NewHTTPClient(cfg.UserServiceURL, 10*time.Second)
	productServiceClient := service.NewHTTPClient(cfg.ProductServiceURL, 10*time.Second)

	// 8. Initialize layers
	orderRepo := repository.NewOrderRepository(db, redisClient)
	orderService := service.NewOrderService(
		orderRepo,
		userServiceClient,
		productServiceClient,
		publisher,
		log.Logger,
	)
	orderHandler := handlers.NewOrderHandler(orderService, log.Logger)

	// 9. Set up router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// 10. Register routes
	setupRoutes(router, orderHandler)

	// 11. Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info("Server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 12. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

func setupRoutes(router *gin.Engine, handler *handlers.OrderHandler) {
	// Health checks
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		orders := v1.Group("/orders")
		{
			// All order endpoints require authentication
			// In production, add AuthMiddleware here
			orders.POST("", handler.CreateOrder)              // Create new order
			orders.GET("", handler.ListUserOrders)            // Get user's orders
			orders.GET("/:id", handler.GetOrderByID)          // Get single order
			orders.PUT("/:id/cancel", handler.CancelOrder)    // Cancel order
			orders.GET("/:id/status", handler.GetOrderStatus) // Get order status
		}
	}
}
