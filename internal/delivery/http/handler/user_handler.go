package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"renfound_v1/internal/domain/models"
	"renfound_v1/internal/usecase/user"
	"renfound_v1/internal/utils/validator"
)

type UserHandler struct {
	userService user.Service
	validator   *validator.Validator
	logger      *zap.Logger
}

func NewUserHandler(
	userService user.Service,
	validator *validator.Validator,
	logger *zap.Logger,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
		logger:      logger.With(zap.String("component", "user_handler")),
	}
}

type TelegramAuthRequest struct {
	InitData string `json:"initData" validate:"required"`
}

func (h *UserHandler) AuthWithTelegram(c *fiber.Ctx) error {
	// parse req
	var req TelegramAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			models.ErrBadRequest, "Invalid req body"))
	}

	// validate req
	if validationErrors, err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			models.ErrInternalServer, "Validation err"))
	} else if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       models.ErrValidation.Error(),
			"description": "Validation failed",
			"errors":      validationErrors,
		})
	}

	//get IP addr
	userAgent := c.Get("User-Agent")
	idAddress := c.IP()

	//Authenticate
	tokens, err := h.userService.AuthWithTelegram(c.Context(), req.InitData, userAgent, idAddress)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, models.ErrInvalidInitData) || errors.Is(err, models.ErrInvalidSignature) {
			status = fiber.StatusBadRequest
		} else if errors.Is(err, models.ErrExpiredToken) {
			status = fiber.StatusUnauthorized
		}

		return c.Status(status).JSON(models.NewErrorResponse(err, ""))
	}
	return c.Status(fiber.StatusOK).JSON(tokens)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// both token refreshing
func (h *UserHandler) RefreshTokens(c *fiber.Ctx) error {
	//parse req
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			models.ErrBadRequest,
			"Invalid req body",
		))
	}

	//validate
	if validationErrors, err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			models.ErrInternalServer,
			"Validation error",
		))
	} else if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       models.ErrValidation.Error(),
			"description": "Validation failed",
			"errors":      validationErrors,
		})

	}
	userAgent := c.Get("User-Agent")
	ipAddress := c.IP()

	tokens, err := h.userService.RefreshTokens(c.Context(), req.RefreshToken, userAgent, ipAddress)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, models.ErrInvalidToken) || errors.Is(err, models.ErrSessionNotFound) {
			status = fiber.StatusUnauthorized
		} else if errors.Is(err, models.ErrExpiredToken) {
			status = fiber.StatusUnauthorized
		}

		return c.Status(status).JSON(models.NewErrorResponse(err, ""))
	}

	return c.Status(fiber.StatusOK).JSON(tokens)
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *UserHandler) Logout(c *fiber.Ctx) error {
	// Parse request
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewErrorResponse(
			models.ErrBadRequest,
			"Invalid request body",
		))
	}

	// Validate request
	if validationErrors, err := h.validator.Validate(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(
			models.ErrInternalServer,
			"Validation error",
		))
	} else if len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       models.ErrValidation.Error(),
			"description": "Validation failed",
			"errors":      validationErrors,
		})
	}

	// Logout user
	if err := h.userService.Logout(c.Context(), req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(err, ""))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// LogoutAll logs out all sessions for a user
func (h *UserHandler) LogoutAll(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
			models.ErrUnauthorized,
			"Missing user ID",
		))
	}

	// Logout all sessions
	if err := h.userService.LogoutAll(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewErrorResponse(err, ""))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "All sessions logged out successfully",
	})
}

// GetMe gets the authenticated user
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
			models.ErrUnauthorized,
			"Missing user ID",
		))
	}

	// Get user
	user, err := h.userService.GetUser(c.Context(), userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, models.ErrUserNotFound) {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(models.NewErrorResponse(err, ""))
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// DeleteMe deletes the authenticated user
func (h *UserHandler) DeleteMe(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewErrorResponse(
			models.ErrUnauthorized,
			"Missing user ID",
		))
	}

	// Delete user
	if err := h.userService.DeleteUser(c.Context(), userID); err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, models.ErrUserNotFound) {
			status = fiber.StatusNotFound
		}

		return c.Status(status).JSON(models.NewErrorResponse(err, ""))
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}
