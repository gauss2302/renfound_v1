package user

import (
	"context"
	"errors"
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

func (s *ServiceImpl) AuthWithTelegram(ctx context.Context, initData, userAgent, ipAddress string) (*models.Tokens, error) {
	// Validate Telegram init data
	telegramUser, err := s.telegramAuth.ValidateInitData(ctx, initData)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	user, err := s.userRepo.GetByTelegramID(ctx, telegramUser.ID)
	if err != nil {
		if !errors.Is(err, models.ErrUserNotFound) {
			s.logger.Error("Failed to get user by Telegram ID", zap.Error(err), zap.Int64("telegram_id", telegramUser.ID))
			return nil, models.ErrInternalServer
		}

		// Create user if not found
		user = models.NewUser(
			telegramUser.ID,
			telegramUser.Username,
			telegramUser.FirstName,
			telegramUser.LastName,
			telegramUser.PhotoURL,
			telegramUser.AuthDate,
		)

		if err := s.userRepo.Create(ctx, user); err != nil {
			s.logger.Error("Failed to create user", zap.Error(err), zap.Int64("telegram_id", telegramUser.ID))
			return nil, models.ErrInternalServer
		}
	} else {
		// Update existing user with new data
		user.Username = telegramUser.Username
		user.FirstName = telegramUser.FirstName
		user.LastName = telegramUser.LastName
		user.PhotoURL = telegramUser.PhotoURL
		user.AuthDate = telegramUser.AuthDate

		if err := s.userRepo.Update(ctx, user); err != nil {
			s.logger.Error("Failed to update user", zap.Error(err), zap.Int64("telegram_id", telegramUser.ID))
			return nil, models.ErrInternalServer
		}
	}

	// Generate tokens
	tokens, err := s.telegramAuth.GenerateTokens(user.ID, user.TelegramID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil, models.ErrInternalServer
	}

	// Create session asynchronously
	session := models.NewSession(
		user.ID,
		tokens.RefreshToken,
		userAgent,
		ipAddress,
		s.cfg.Config.JWT.RefreshTTL,
	)

	// Submit task to worker pool
	s.workerPool.Submit(func() {
		// Use background context as the original ctx might be cancelled
		bgCtx := context.Background()
		if err := s.userRepo.CreateSession(bgCtx, session); err != nil {
			s.logger.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID.String()))
		}
	})

	return tokens, nil
}

func (s *ServiceImpl) RefreshTokens(ctx context.Context, refreshToken, userAgent, ipAddress string) (*models.Tokens, error) {
	// Validate refresh token
	userIDStr, err := s.telegramAuth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		s.logger.Error("Invalid user ID in token", zap.Error(err), zap.String("user_id", userIDStr))
		return nil, models.ErrInvalidToken
	}

	// Check if session exists
	session, err := s.userRepo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			return nil, models.ErrInvalidToken
		}
		s.logger.Error("Failed to get session", zap.Error(err), zap.String("refresh_token", refreshToken))
		return nil, models.ErrInternalServer
	}

	// Verify session belongs to the user
	if session.UserID != userID {
		s.logger.Warn("Session user ID mismatch",
			zap.String("token_user_id", userID.String()),
			zap.String("session_user_id", session.UserID.String()))
		return nil, models.ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, models.ErrInvalidToken
		}
		s.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, models.ErrInternalServer
	}

	// Delete old session
	if err := s.userRepo.DeleteSession(ctx, session.ID); err != nil {
		if !errors.Is(err, models.ErrSessionNotFound) {
			s.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", session.ID.String()))
			return nil, models.ErrInternalServer
		}
	}

	// Generate new tokens
	tokens, err := s.telegramAuth.GenerateTokens(user.ID, user.TelegramID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil, models.ErrInternalServer
	}

	// Create new session asynchronously
	newSession := models.NewSession(
		user.ID,
		tokens.RefreshToken,
		userAgent,
		ipAddress,
		s.cfg.Config.JWT.RefreshTTL,
	)

	// Submit task to worker pool
	s.workerPool.Submit(func() {
		// Use background context as the original ctx might be cancelled
		bgCtx := context.Background()
		if err := s.userRepo.CreateSession(bgCtx, newSession); err != nil {
			s.logger.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID.String()))
		}
	})

	return tokens, nil
}

func (s *ServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	// Check if session exists
	session, err := s.userRepo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			// Already logged out
			return nil
		}
		s.logger.Error("Failed to get session", zap.Error(err), zap.String("refresh_token", refreshToken))
		return models.ErrInternalServer
	}

	// Delete session
	if err := s.userRepo.DeleteSession(ctx, session.ID); err != nil {
		if !errors.Is(err, models.ErrSessionNotFound) {
			s.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", session.ID.String()))
			return models.ErrInternalServer
		}
	}

	return nil
}

func (s *ServiceImpl) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	// Delete all sessions for the user
	if err := s.userRepo.DeleteUserSessions(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user sessions", zap.Error(err), zap.String("user_id", userID.String()))
		return models.ErrInternalServer
	}

	return nil
}

func (s *ServiceImpl) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, models.ErrUserNotFound
		}
		s.logger.Error("Failed to get user", zap.Error(err), zap.String("user_id", id.String()))
		return nil, models.ErrInternalServer
	}

	return user, nil
}

func (s *ServiceImpl) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return nil, models.ErrUserNotFound
		}
		s.logger.Error("Failed to get user by Telegram ID", zap.Error(err), zap.Int64("telegram_id", telegramID))
		return nil, models.ErrInternalServer
	}

	return user, nil
}

func (s *ServiceImpl) UpdateUser(ctx context.Context, user *models.User) error {
	if err := s.userRepo.Update(ctx, user); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.ErrUserNotFound
		}
		s.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID.String()))
		return models.ErrInternalServer
	}

	return nil
}

func (s *ServiceImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	//remove all sessions of a user
	if err := s.userRepo.DeleteUserSessions(ctx, id); err != nil {
		s.logger.Error("Failed to delete user sessions", zap.Error(err), zap.String("user_id", id.String()))
		return models.ErrInternalServer
	}

	// Delete the user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.ErrUserNotFound
		}
		s.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id.String()))
		return models.ErrInternalServer
	}
	return nil
}
