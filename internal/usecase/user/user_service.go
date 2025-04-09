package user

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"renfound_v1/config"
	"renfound_v1/infrastructure/auth"
	"renfound_v1/internal/domain/models"
	"renfound_v1/internal/domain/repository"
	"renfound_v1/internal/utils/async"
)

type ServiceImpl struct {
	cfg          *config.AppConfig
	userRepo     repository.UserRepository
	telegramAuth *auth.TelegramAuth
	workerPool   *async.WorkerPool
	logger       *zap.Logger
}

func NewService(
	cfg *config.AppConfig,
	userRepo repository.UserRepository,
	telegramAuth *auth.TelegramAuth,
	workerPool *async.WorkerPool) Service {
	return &ServiceImpl{
		cfg:          cfg,
		userRepo:     userRepo,
		telegramAuth: telegramAuth,
		workerPool:   workerPool,
		logger:       cfg.Logger.With(zap.String("component", "user_service")),
	}
}

func (s ServiceImpl) AuthWithTelegram(ctx context.Context, initData, userAgent, ipAddress string) (*models.Tokens, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) RefreshTokens(ctx context.Context, refreshToken, userAgent, ipAddress string) (*models.Tokens, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) UpdateUser(ctx context.Context, user *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (s ServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}
