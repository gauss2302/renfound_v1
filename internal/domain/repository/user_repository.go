package repository

import (
	"context"

	"github.com/google/uuid"
	"renfound_v1/internal/domain/models"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	// User operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Session operations
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUserSessions(ctx context.Context, userID uuid.UUID) error
}
