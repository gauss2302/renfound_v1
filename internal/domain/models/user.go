package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user entity
type User struct {
	ID         uuid.UUID `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username,omitempty"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	PhotoURL   string    `json:"photo_url,omitempty"`
	AuthDate   int64     `json:"auth_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewUser(telegramID int64, username, firstName, lastName, photoURL string, authDate int64) *User {
	return &User{
		ID:         uuid.New(),
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		PhotoURL:   photoURL,
		AuthDate:   authDate,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewSession(userID uuid.UUID, refreshToken, userAgent, ipAddress string, ttl time.Duration) *Session {
	return &Session{
		ID:           uuid.New(),
		UserID:       userID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(ttl),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	UserID     string `json:"user_id"`
	TelegramID int64  `json:"telegram_id"`
}
