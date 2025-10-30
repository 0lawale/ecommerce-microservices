package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for structured logging
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new logger instance
// Production mode: JSON formatted logs
// Development mode: Console formatted logs with colors
func NewLogger(serviceName string, isDevelopment bool) (*Logger, error) {
	var config zap.Config

	if isDevelopment {
		// Development config: human-readable console output
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Production config: JSON output for log aggregation (ELK stack)
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Add service name to all logs for tracking in centralized logging
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
		"host":    getHostname(),
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{logger}, nil
}

// getHostname returns the hostname for tracking which instance logged
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// WithFields adds structured fields to logs
// Example: logger.WithFields(zap.String("user_id", "123")).Info("User logged in")
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.With(fields...)}
}

// LogHTTPRequest logs incoming HTTP requests with timing
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	l.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
	)
}
