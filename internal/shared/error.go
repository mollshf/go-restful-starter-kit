package shared

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("APIError: %d %s", e.Status, e.Message)
}

func NewNotFoundError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusNotFound,
		Code:    code,
		Message: message,
	}
}

func NewInternalError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    code,
		Message: message,
	}
}

func NewBadRequestError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    code,
		Message: message,
	}
}

func NewUnauthorizedError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    code,
		Message: message,
	}
}

func NewForbiddenError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    code,
		Message: message,
	}
}

func NewTooManyRequestsError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusTooManyRequests,
		Code:    code,
		Message: message,
	}
}

func NewValidationError(message string, code string, details any) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    code,
		Message: message,
		Details: details,
	}
}

func NewInternalServerError(message string, code string) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    code,
		Message: message,
	}
}
