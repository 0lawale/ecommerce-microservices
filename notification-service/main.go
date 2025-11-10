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

	"ecommerce/notification-service/handlers"
	"ecommerce/notification-service/messaging"
	"ecommerce/notification-service/repository"
	"ecommerce/notification-service/service"
	"ecommerce/shared/config"
	"ecommerce/shared/logger"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig("notification-service")

	// 2. Initialize logger
	log, err := logger.NewLogger(cfg.ServiceName, cfg.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Notification Service",
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

	// 5. Initialize repository and service
	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo, log.Logger)

	// 6. Initialize RabbitMQ consumer
	consumer, err := messaging.NewRabbitMQConsumer(cfg.RabbitMQURL, notificationService, log.Logger)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer consumer.Close()

	log.Info("RabbitMQ consumer initialized")

	// 7. Start consuming messages in background
	go func() {
		log.Info("Starting to consume order events...")
		if err := consumer.StartConsuming(); err != nil {
			log.Error("Consumer error", zap.Error(err))
		}
	}()

	// 8. Set up HTTP server for health checks
	notificationHandler := handlers.NewNotificationHandler(notificationService, log.Logger)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	setupRoutes(router, notificationHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info("Health check server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 9. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down service...")

	// Stop consumer
	consumer.Close()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Service exited")
}

func setupRoutes(router *gin.Engine, handler *handlers.NotificationHandler) {
	// Health checks only - this service primarily consumes from RabbitMQ
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	// Optional API endpoints for viewing notifications
	v1 := router.Group("/api/v1")
	{
		notifications := v1.Group("/notifications")
		{
			notifications.GET("/user/:user_id", handler.GetUserNotifications)
			notifications.PUT("/:id/read", handler.MarkAsRead)
		}
	}
}
