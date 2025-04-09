package user

import (
	"context"

	"github.com/google/uuid"
	"renfound_v1/internal/domain/models"
)

// Service defines the interface for user-related operations
type Service interface {
	// Authentication methods
	AuthWithTelegram(ctx context.Context, initData, userAgent, ipAddress string) (*models.Tokens, error)
	RefreshTokens(ctx context.Context, refreshToken, userAgent, ipAddress string) (*models.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// User methods
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
