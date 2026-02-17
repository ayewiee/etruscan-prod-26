package apierrors

import (
	"etruscan/internal/domain/models"
	"net/http"
)

func HTTPStatusFromErrorCode(code models.ErrorCode) int {
	switch code {
	case models.ErrCodeBadRequest:
		return http.StatusBadRequest

	case models.ErrCodeValidationFailed, models.ErrCodeDSLParseError:
		return http.StatusUnprocessableEntity

	case models.ErrCodeUnauthorized:
		return http.StatusUnauthorized

	case models.ErrCodeForbidden:
		return http.StatusForbidden

	case models.ErrCodeNotFound:
		return http.StatusNotFound

	case models.ErrCodeEmailAlreadyExists:
		return http.StatusConflict

	case models.ErrCodeInternal:
		return http.StatusInternalServerError

	default:
		return http.StatusInternalServerError // fallback
	}
}

// ErrorCodeFromHTTPStatus for fallback (for echo.HTTPError)
func ErrorCodeFromHTTPStatus(status int) models.ErrorCode {
	switch status {
	case http.StatusBadRequest:
		return models.ErrCodeBadRequest
	case http.StatusUnauthorized:
		return models.ErrCodeUnauthorized
	case http.StatusForbidden:
		return models.ErrCodeForbidden
	case http.StatusNotFound:
		return models.ErrCodeNotFound
	case http.StatusConflict:
		return models.ErrCodeBadRequest
	default:
		return models.ErrCodeInternal
	}
}

func FieldErrorsListToDTO(ferrs []models.FieldError) []FieldErrorDTO {
	var dtoerrs []FieldErrorDTO
	for _, ferr := range ferrs {
		dtoerrs = append(dtoerrs, FieldErrorDTO{
			Field:         ferr.Field,
			Issue:         ferr.Issue,
			RejectedValue: ferr.RejectedValue,
		})
	}
	return dtoerrs
}

func APIErrorToErrorResponse(err models.ApiError, traceId string, timestamp string, path string) ErrorResponseDTO {
	return ErrorResponseDTO{
		Code:        string(err.Code),
		Message:     err.Message,
		TraceID:     traceId,
		Timestamp:   timestamp,
		Path:        path,
		Details:     err.Details,
		FieldErrors: FieldErrorsListToDTO(err.FieldErrors),
	}
}
