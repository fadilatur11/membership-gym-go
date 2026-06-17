package errors

import "net/http"

type AppError struct {
	Code    int
	Message string
	Errs    any
}

func (e *AppError) Error() string { return e.Message }

func New(code int, message string, errs any) *AppError {
	return &AppError{Code: code, Message: message, Errs: errs}
}

func BadRequest(message string) *AppError   { return New(http.StatusBadRequest, message, nil) }
func Unauthorized(message string) *AppError { return New(http.StatusUnauthorized, message, nil) }
func Forbidden(message string) *AppError    { return New(http.StatusForbidden, message, nil) }
func NotFound(message string) *AppError     { return New(http.StatusNotFound, message, nil) }
func Conflict(message string) *AppError     { return New(http.StatusConflict, message, nil) }
func Internal(message string) *AppError     { return New(http.StatusInternalServerError, message, nil) }
