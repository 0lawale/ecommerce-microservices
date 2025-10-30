package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	ServiceName string
	Port        string
	Environment string // "development", "staging", "production"

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis configuration
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// JWT configuration
	JWTSecret string

	// Message Queue (RabbitMQ)
	RabbitMQURL string

	// Other services URLs (for inter-service communication)
	UserServiceURL    string
	ProductServiceURL string
	OrderServiceURL   string
}

// LoadConfig loads configuration from environment variables
// In Kubernetes, these come from ConfigMaps and Secrets
func LoadConfig(serviceName string) *Config {
	return &Config{
		ServiceName: serviceName,
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", serviceName),

		// Redis
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		// RabbitMQ
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),

		// Service URLs (used by API Gateway and inter-service calls)
		UserServiceURL:    getEnv("USER_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
	}
}

// GetDatabaseURL returns PostgreSQL connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

// GetRedisURL returns Redis connection string
func (c *Config) GetRedisURL() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv gets environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets environment variable as integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
