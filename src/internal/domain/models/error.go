package models

import (
	"fmt"
)

type ErrorCode string

const (
	ErrCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrCodeValidationFailed   ErrorCode = "VALIDATION_FAILED"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"
	ErrCodeDSLParseError      ErrorCode = "DSL_PARSE_ERROR"
	ErrCodeInternal           ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeLocked             ErrorCode = "LOCKED"
)

var (
	ErrInvalidJSON = NewApiError(
		ErrCodeBadRequest,
		"Invalid JSON",
		map[string]interface{}{
			"hint": "Check commas/quotation marks",
		},
		nil, nil,
	)
	ErrUnauthorized = NewApiError(
		ErrCodeUnauthorized,
		"Authorization token is invalid",
		nil, nil, nil,
	)
	ErrInvalidCredentials = NewApiError(
		ErrCodeUnauthorized,
		"Invalid email or password",
		nil, nil, nil,
	)
	ErrForbidden = NewApiError(
		ErrCodeForbidden,
		"Not enough permission",
		nil, nil, nil,
	)
	ErrInternal = NewApiError(
		ErrCodeInternal,
		"Internal server error",
		nil, nil, nil,
	)
)

type FieldError struct {
	Field         string
	Issue         string
	RejectedValue interface{}
}

func (f FieldError) Error() string {
	return fmt.Sprintf("FieldError: [%s] %s", f.Field, f.Issue)
}

type ApiError struct {
	Code        ErrorCode
	Message     string
	Details     map[string]interface{}
	FieldErrors []FieldError
	Cause       error
}

func (e ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e ApiError) Unwrap() error {
	return e.Cause
}

func NewApiError(
	code ErrorCode,
	message string,
	details map[string]interface{},
	fieldErrors []FieldError,
	cause error,
) *ApiError {
	return &ApiError{
		Code:        code,
		Message:     message,
		Details:     details,
		FieldErrors: fieldErrors,
		Cause:       cause,
	}
}

func NewErrNotFound(message string, details map[string]interface{}, cause error) *ApiError {
	return NewApiError(
		ErrCodeNotFound,
		message,
		details, nil, cause,
	)
}

func NewErrForbidden(message string) *ApiError {
	return NewApiError(
		ErrCodeForbidden,
		message,
		nil, nil, nil,
	)
}

func NewErrLocked(message string) *ApiError {
	return NewApiError(
		ErrCodeLocked,
		message,
		nil, nil, nil,
	)
}
