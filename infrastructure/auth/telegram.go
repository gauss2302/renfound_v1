package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"go.uber.org/zap"

	"renfound_v1/config"
	"renfound_v1/internal/domain/models"
)

// TelegramAuth handles authentication with Telegram
type TelegramAuth struct {
	cfg    *config.AppConfig
	logger *zap.Logger
}

// NewTelegramAuth creates a new TelegramAuth
func NewTelegramAuth(cfg *config.AppConfig) *TelegramAuth {
	logger := cfg.Logger.With(zap.String("component", "telegram_auth"))

	// Validate JWT configuration
	if cfg.Config.JWT.AccessSecret == "" || cfg.Config.JWT.RefreshSecret == "" {
		logger.Warn("JWT secrets not configured properly")
	}

	if cfg.Config.JWT.AccessTTL == 0 {
		logger.Info("Using default JWT access token TTL of 15 minutes")
		cfg.Config.JWT.AccessTTL = 15 * time.Minute
	}

	if cfg.Config.JWT.RefreshTTL == 0 {
		logger.Info("Using default JWT refresh token TTL of 7 days")
		cfg.Config.JWT.RefreshTTL = 7 * 24 * time.Hour
	}

	return &TelegramAuth{
		cfg:    cfg,
		logger: logger,
	}
}

// TelegramUser represents user data from Telegram init data
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
}

// ValidateInitData validates Telegram init data and returns user information
func (a *TelegramAuth) ValidateInitData(ctx context.Context, initData string) (*TelegramUser, error) {
	// Use bot token from config
	botToken := a.cfg.Config.JWT.AccessSecret // Using JWT access secret as bot token for simplicity

	// Validate the init data using the package
	// Allow a 24 hour expiration time
	err := initdata.Validate(initData, botToken, 24*time.Hour)
	if err != nil {
		a.logger.Error("Failed to validate init data", zap.Error(err))
		return nil, models.ErrInvalidInitData
	}

	// Parse the init data after validation
	data, err := initdata.Parse(initData)
	if err != nil {
		a.logger.Error("Failed to parse init data", zap.Error(err))
		return nil, models.ErrInvalidInitData
	}

	// Extract user information - checking for nil first
	if &data.User == nil {
		a.logger.Warn("No user data in init data")
		return nil, models.ErrInvalidInitData
	}

	// Now we know the User struct fields, we can safely access them
	telegramUser := &TelegramUser{
		ID:        data.User.ID,
		FirstName: data.User.FirstName,
		LastName:  data.User.LastName,
		Username:  data.User.Username,
		PhotoURL:  data.User.PhotoURL,
		AuthDate:  data.AuthDate().Unix(),
	}

	a.logger.Info("Successfully validated Telegram init data",
		zap.Int64("telegram_id", telegramUser.ID),
		zap.String("first_name", telegramUser.FirstName),
		zap.Int64("auth_date", telegramUser.AuthDate))

	return telegramUser, nil
}

// GenerateTokens generates JWT tokens for a user
func (a *TelegramAuth) GenerateTokens(userID uuid.UUID, telegramID int64) (*models.Tokens, error) {
	// Generate access token
	accessToken, err := a.generateAccessToken(userID, telegramID)
	if err != nil {
		a.logger.Error("Failed to generate access token",
			zap.Error(err),
			zap.String("user_id", userID.String()),
			zap.Int64("telegram_id", telegramID))
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := a.generateRefreshToken(userID)
	if err != nil {
		a.logger.Error("Failed to generate refresh token",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	a.logger.Debug("Generated tokens successfully",
		zap.String("user_id", userID.String()),
		zap.Int64("telegram_id", telegramID))

	return &models.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (a *TelegramAuth) ValidateAccessToken(tokenString string) (*models.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.Config.JWT.AccessSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, models.ErrExpiredToken
		}
		a.logger.Error("Failed to parse access token", zap.Error(err))
		return nil, models.ErrInvalidToken
	}

	if !token.Valid {
		return nil, models.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	telegramIDFloat, ok := claims["telegram_id"].(float64)
	if !ok {
		return nil, models.ErrInvalidToken
	}

	telegramID := int64(telegramIDFloat)

	return &models.Claims{
		UserID:     userID,
		TelegramID: telegramID,
	}, nil
}

// generateAccessToken generates a JWT access token
func (a *TelegramAuth) generateAccessToken(userID uuid.UUID, telegramID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID.String(),
		"telegram_id": telegramID,
		"exp":         time.Now().Add(a.cfg.Config.JWT.AccessTTL).Unix(),
		"iat":         time.Now().Unix(),
		"type":        "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.cfg.Config.JWT.AccessSecret))
	if err != nil {
		a.logger.Error("Failed to sign access token", zap.Error(err))
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// generateRefreshToken generates a JWT refresh token
func (a *TelegramAuth) generateRefreshToken(userID uuid.UUID) (string, error) {
	jti := uuid.New().String() // Add a unique ID to the token for revocation

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(a.cfg.Config.JWT.RefreshTTL).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     jti,
		"type":    "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.cfg.Config.JWT.RefreshSecret))
	if err != nil {
		a.logger.Error("Failed to sign refresh token", zap.Error(err))
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (a *TelegramAuth) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.cfg.Config.JWT.RefreshSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", models.ErrExpiredToken
		}
		a.logger.Error("Failed to parse refresh token", zap.Error(err))
		return "", models.ErrInvalidToken
	}

	if !token.Valid {
		return "", models.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", models.ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", models.ErrInvalidToken
	}

	return userID, nil
}
