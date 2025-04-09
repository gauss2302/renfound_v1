package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"renfound_v1/infrastructure/auth"
	"renfound_v1/internal/domain/models"
	"strings"
)

type AuthMiddleware struct {
	telegramAuth *auth.TelegramAuth
	logger       *zap.Logger
}

func NewAuthMiddleware(telegramAuth *auth.TelegramAuth, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		telegramAuth: telegramAuth,
		logger:       logger.With(zap.String("component", "auth_middleware")),
	}
}

func (m *AuthMiddleware) Authenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		//get token from the header
		authheader := c.Get("Authorization")
		if authheader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
				models.ErrUnauthorized,
				"Missing auth header",
			))
		}

		// check correct format for the token
		parts := strings.Split(authheader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
				models.ErrUnauthorized,
				"Invalid auth header format"))
		}

		token := parts[1]

		//validate
		claims, err := m.telegramAuth.ValidateAccessToken(token)
		if err != nil {
			status := fiber.StatusUnauthorized
			errMsg := "Invalid token"

			if err == models.ErrExpiredToken {
				errMsg = "Token has expired"
			}

			return c.Status(status).JSON(models.NewErrorResponse(
				models.ErrUnauthorized, errMsg))
		}

		//parse user id from claims
		userID, err := uuid.Parse(claims.UserID)

		if err != nil {
			m.logger.Error("Invalid user ID in token", zap.Error(err), zap.String("user_id", claims.UserID))
			return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
				models.ErrUnauthorized,
				"Invalid token",
			))
		}

		// set user id and tg id in context for later
		c.Locals("userID", userID)
		c.Locals("telegramID", claims.TelegramID)

		return c.Next()
	}
}

func (m *AuthMiddleware) OptionalAuthenticate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Check if the token is in the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		token := parts[1]

		// Validate token
		claims, err := m.telegramAuth.ValidateAccessToken(token)
		if err != nil {
			// Just continue without authentication
			return c.Next()
		}

		// Parse user ID from claims
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			// Just continue without authentication
			return c.Next()
		}

		// Set user ID and telegram ID in context for later use
		c.Locals("userID", userID)
		c.Locals("telegramID", claims.TelegramID)

		return c.Next()
	}
}
