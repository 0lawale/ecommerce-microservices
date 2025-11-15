package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ecommerce/api-gateway/handlers"
	"ecommerce/shared/config"
	"ecommerce/shared/logger"
)

func main() {
	cfg := config.LoadConfig("api-gateway")

	log, err := logger.NewLogger(cfg.ServiceName, cfg.IsDevelopment())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting API Gateway",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	proxyHandler := handlers.NewProxyHandler(
		cfg.UserServiceURL,
		cfg.ProductServiceURL,
		cfg.OrderServiceURL,
		log.Logger,
	)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// ADD CORS MIDDLEWARE HERE
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	setupRoutes(router, proxyHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info("API Gateway listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("API Gateway exited")
}

func setupRoutes(router *gin.Engine, handler *handlers.ProxyHandler) {
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadinessCheck)

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handler.ProxyToUserService)
			auth.POST("/login", handler.ProxyToUserService)
		}

		users := api.Group("/users")
		{
			users.GET("/me", handler.ProxyToUserService)
			users.PUT("/me", handler.ProxyToUserService)
			users.GET("/:id", handler.ProxyToUserService)
		}

		products := api.Group("/products")
		{
			products.GET("", handler.ProxyToProductService)
			products.GET("/:id", handler.ProxyToProductService)
			products.GET("/category/:category", handler.ProxyToProductService)
			products.GET("/search", handler.ProxyToProductService)
			products.POST("", handler.ProxyToProductService)
			products.PUT("/:id", handler.ProxyToProductService)
			products.DELETE("/:id", handler.ProxyToProductService)
			products.PUT("/:id/stock", handler.ProxyToProductService)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", handler.ProxyToOrderService)
			orders.GET("", handler.ProxyToOrderService)
			orders.GET("/:id", handler.ProxyToOrderService)
			orders.PUT("/:id/cancel", handler.ProxyToOrderService)
			orders.GET("/:id/status", handler.ProxyToOrderService)
		}
	}
}
