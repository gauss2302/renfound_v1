package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.With(zap.String("component", "http_middleware")),
	}
}

// Logger is a middleware that logs HTTP requests
func (m *LoggingMiddleware) Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Generate request ID
		reqID := uuid.New().String()
		c.Locals("requestID", reqID)

		// Set request ID header
		c.Set("X-Request-ID", reqID)

		// Process request
		err := c.Next()

		// Latency
		latency := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Get user ID if available
		var userID interface{}
		if c.Locals("userID") != nil {
			userID = c.Locals("userID")
		}

		// Determine log level based on status code
		logFunc := m.logger.Info
		if status >= 500 {
			logFunc = m.logger.Error
		} else if status >= 400 {
			logFunc = m.logger.Warn
		}

		// Log the request
		logFunc("HTTP Request",
			zap.String("request_id", reqID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("query", string(c.Request().URI().QueryString())),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
			zap.Any("user_id", userID),
			zap.Int64("body_size", int64(len(c.Request().Body()))),
		)

		return err
	}
}

// RecoverWithLogger is a middleware that recovers from panics and logs them
func (m *LoggingMiddleware) RecoverWithLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Get request ID if available
				reqID, _ := c.Locals("requestID").(string)
				if reqID == "" {
					reqID = uuid.New().String()
				}

				m.logger.Error("Recovered from panic",
					zap.String("request_id", reqID),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.Any("panic", r),
					zap.String("ip", c.IP()),
					zap.String("user_agent", c.Get("User-Agent")),
				)

				// Return internal server error
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal Server Error",
				})
			}
		}()

		return c.Next()
	}
}
