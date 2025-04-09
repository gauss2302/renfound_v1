package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"renfound_v1/internal/domain/models"
	"renfound_v1/internal/domain/repository"
)

type UserRepositoryImpl struct {
	db     *Database
	logger *zap.Logger
}

func NewUserRepository(db *Database, logger *zap.Logger) repository.UserRepository {
	return &UserRepositoryImpl{
		db:     db,
		logger: logger.With(zap.String("component", "user_repository")),
	}
}

func (r UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, telegram_id, username, first_name, last_name, photo_url, auth_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PhotoURL,
		user.AuthDate,
		user.CreatedAt,
		user.UpdatedAt)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err), zap.Int64("telegram_id", user.TelegramID))
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil

}

func (r UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, photo_url, auth_date, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &models.User{}

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
		&user.AuthDate,
		&user.CreatedAt,
		&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		r.logger.Error("Failed to get user by ID", zap.Error(err), zap.String("user_id", id.String()))
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r UserRepositoryImpl) GetByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, photo_url, auth_date, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PhotoURL,
		&user.AuthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		r.logger.Error("Failed to get user by Telegram ID", zap.Error(err), zap.Int64("telegram_id", telegramID))
		return nil, fmt.Errorf("failed to get user by Telegram ID: %w", err)
	}

	return user, nil
}

func (r UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3, photo_url = $4, auth_date = $5, updated_at = NOW()
		WHERE id = $6
	`

	result, err := r.db.Pool.Exec(ctx, query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.PhotoURL,
		user.AuthDate,
		user.ID,
	)
	if err != nil {
		r.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID.String()))
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", zap.Error(err), zap.String("user_id", id.String()))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

func (r UserRepositoryImpl) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
		session.CreatedAt,
		session.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create session", zap.Error(err), zap.String("user_id", session.UserID.String()))
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r UserRepositoryImpl) GetSessionByToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at
		FROM sessions
		WHERE refresh_token = $1
	`

	session := &models.Session{}
	err := r.db.Pool.QueryRow(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrSessionNotFound
		}
		r.logger.Error("Failed to get session by token", zap.Error(err))
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return session, nil
}

func (r UserRepositoryImpl) DeleteSession(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", id.String()))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrSessionNotFound
	}

	return nil
}

func (r UserRepositoryImpl) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user sessions", zap.Error(err), zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}
