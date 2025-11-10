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

	"ecommerce/product-service/handlers"
	"ecommerce/product-service/repository"
	"ecommerce/product-service/service"
	"ecommerce/shared/config"
	"ecommerce/shared/logger"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig("product-service")

	// 2. Initialize logger
	log, err := logger.NewLogger(cfg.ServiceName, cfg.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting Product Service",
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

	// 6. Initialize layers
	productRepo := repository.NewProductRepository(db, redisClient)
	productService := service.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService, log.Logger)

	// 7. Set up router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// 8. Register routes
	setupRoutes(router, productHandler)

	// 9. Start server
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

	// 10. Graceful shutdown
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

func setupRoutes(router *gin.Engine, handler *handlers.ProductHandler) {
	// Health checks
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			// Public routes (anyone can view products)
			products.GET("", handler.ListProducts)       // List with filters
			products.GET("/:id", handler.GetProductByID) // Get single product
			products.GET("/category/:category", handler.GetProductsByCategory)
			products.GET("/search", handler.SearchProducts) // Search by name

			// Protected routes (require authentication - will add middleware in handler)
			// Admin only routes would need AdminMiddleware
			products.POST("", handler.CreateProduct)        // Create new product
			products.PUT("/:id", handler.UpdateProduct)     // Update product
			products.DELETE("/:id", handler.DeleteProduct)  // Delete product
			products.PUT("/:id/stock", handler.UpdateStock) // Update stock
		}
	}
}
