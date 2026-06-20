package domain

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Status  int
	Message string
	Cause   error
}

func NewError(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithCause(err error) *AppError {
	clone := *e
	clone.Cause = err
	return &clone
}

var (
	ErrTokenRevoked            = NewError(http.StatusBadRequest, "Invalid or revoked session.")
	ErrUnknownRedis            = NewError(http.StatusInternalServerError, "Unknown error when interacting with redis.")
	ErrRefreshTokenNotProvided = NewError(http.StatusBadRequest, "Refresh token not found in request.")
	ErrSessionNotFound   = NewError(http.StatusUnauthorized, "Session not found or already expired.")
	ErrSessionIPMismatch = NewError(http.StatusUnauthorized, "Invalid or revoked session.")
	ErrSessionDeleteConflict = NewError(http.StatusConflict, "Session was already used or deleted.")
	ErrTokenSigningFailed = NewError(http.StatusInternalServerError, "Failed to sign access token.")
	ErrOauthExchangeFailed = NewError(http.StatusInternalServerError, "Failed to exchange code for token.")
)