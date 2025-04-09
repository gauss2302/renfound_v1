package models

import "errors"

var (
	// Generic errors
	ErrInternalServer = errors.New("internal server error")
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource already exists")
	ErrBadRequest     = errors.New("bad request")
	ErrValidation     = errors.New("validation error")

	// Authentication errors
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrInvalidInitData    = errors.New("invalid telegram init data")

	// User errors
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")

	// Session errors
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session")
)

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

func NewErrorResponse(err error, description string) ErrorResponse {
	return ErrorResponse{
		Error:       err.Error(),
		Description: description,
	}
}
