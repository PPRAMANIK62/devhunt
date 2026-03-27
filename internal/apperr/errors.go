package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorType string

const (
	TypeNotFound     ErrorType = "NOT_FOUND"
	TypeUnauthorized ErrorType = "UNAUTHORIZED"
	TypeForbidden    ErrorType = "FORBIDDEN"
	TypeValidation   ErrorType = "VALIDATION_ERROR"
	TypeConflict     ErrorType = "CONFLICT"
	TypeInternal     ErrorType = "INTERNAL_ERROR"
	TypeExternal     ErrorType = "EXTERNAL_SERVICE_ERROR"
)

type AppError struct {
	Type    ErrorType
	Message string // shown to the client
	Cause   error  // logged internally, never sent to client
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

// Constructor for each error category
func NotFound(resource string) *AppError {
	return &AppError{
		Type:    TypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}
func Unauthorized(msg string) *AppError {
	return &AppError{
		Type:    TypeUnauthorized,
		Message: msg,
	}
}
func Forbidden(msg string) *AppError {
	return &AppError{
		Type:    TypeForbidden,
		Message: msg,
	}
}
func Validation(msg string) *AppError {
	return &AppError{
		Type:    TypeValidation,
		Message: msg,
	}
}
func Conflict(msg string) *AppError {
	return &AppError{
		Type:    TypeConflict,
		Message: msg,
	}
}
func Internal(msg string, cause error) *AppError {
	return &AppError{
		Type:    TypeInternal,
		Message: msg,
		Cause:   cause,
	}
}
func External(service string, cause error) *AppError {
	return &AppError{
		Type:    TypeExternal,
		Message: fmt.Sprintf("%s is temporarily unavailable", service),
	}
}

// HTTPStatus is the single place that maps error type -> HTTP status code
func HTTPStatus(err error) int {
	var e *AppError
	if !errors.As(err, &e) {
		return http.StatusInternalServerError
	}

	switch e.Type {
	case TypeNotFound:
		return http.StatusNotFound
	case TypeUnauthorized:
		return http.StatusUnauthorized
	case TypeForbidden:
		return http.StatusForbidden
	case TypeValidation:
		return http.StatusBadRequest
	case TypeConflict:
		return http.StatusConflict
	case TypeExternal:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
