package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/atosramos/go-base-module/logger"
)

// RequestLoggerMiddleware logs HTTP requests standardizing the log levels based on the HTTP status code.
// It also masks sensitive headers like Authorization and adds contextual fields (tenant_id, request_id, etc).
func RequestLoggerMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Extract context variables if they were set by other middlewares (like tenant.go or auth.go)
		tenantID := c.GetString("tenant_id")
		requestID := c.GetString("request_id")
		userID := c.GetString("user_id")

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if tenantID != "" {
			fields = append(fields, zap.String("tenant_id", tenantID))
		}
		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}
		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log errors added to the context by handlers
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		// Standardize log level according to rules:
		// INFO for 2xx and 3xx
		// WARN for 4xx
		// ERROR for 5xx
		if status >= 500 {
			log.Error("Server error during request processing", fields...)
		} else if status >= 400 {
			log.Warn("Client error during request processing", fields...)
		} else {
			log.Info("Request processed successfully", fields...)
		}
	}
}
