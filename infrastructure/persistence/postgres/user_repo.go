package postgres

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"renfound_v1/internal/domain/models"
	"renfound_v1/internal/domain/repository"
)

type UserRepositoryImpl struct {
	db     *Database
	logger *zap.Logger
}

func (u UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) CreateSession(ctx context.Context, session *models.Session) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) DeleteSession(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (u UserRepositoryImpl) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func NewUserRepository(db *Database, logger *zap.Logger) repository.UserRepository {
	return &UserRepositoryImpl{
		db:     db,
		logger: logger.With(zap.String("component", "user_repository")),
	}
}
