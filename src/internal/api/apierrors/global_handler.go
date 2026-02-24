package apierrors

import (
	"errors"
	"etruscan/internal/domain/models"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GlobalErrorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		traceID := c.Response().Header().Get("X-Request-Id")
		path := c.Request().URL.Path
		timestamp := time.Now().UTC().Format(time.RFC3339)

		var (
			httpCode    = http.StatusInternalServerError
			strCode     = models.ErrCodeInternal
			message     = "Internal Server Error"
			details     map[string]interface{}
			fieldErrors []FieldErrorDTO
		)

		var echoHttpErr *echo.HTTPError
		isHTTPError := errors.As(err, &echoHttpErr)
		if isHTTPError {
			httpCode = echoHttpErr.Code
			strCode = ErrorCodeFromHTTPStatus(httpCode) // fallback
			message = fmt.Sprint(echoHttpErr.Message)
		}

		baseErr := err
		if isHTTPError && echoHttpErr.Internal != nil {
			baseErr = echoHttpErr.Internal
		}

		var apiErr *models.ApiError
		// try to unwrap API errors first
		errors.As(baseErr, &apiErr)

		// if it's not already an API error, see if it's a DB constraint error
		if apiErr == nil {
			apiErr = dbErrorToApiError(baseErr)
		}

		if apiErr != nil {
			strCode = apiErr.Code
			message = apiErr.Message
			details = apiErr.Details
			httpCode = HTTPStatusFromErrorCode(strCode)
			fieldErrors = FieldErrorsListToDTO(apiErr.FieldErrors)
		}

		response := ErrorResponseDTO{
			Code:        string(strCode),
			Message:     message,
			TraceID:     traceID,
			Timestamp:   timestamp,
			Path:        path,
			Details:     details,
			FieldErrors: fieldErrors,
		}

		fields := []zap.Field{
			zap.Int("status", httpCode),
			zap.String("message", message),
			zap.Error(err),
			zap.String("uri", c.Request().RequestURI),
			zap.String("method", c.Request().Method),
			zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
		}

		if httpCode >= 500 {
			log.Error("http_error", fields...)
		} else {
			log.Warn("http_error", fields...)
		}

		_ = c.JSON(httpCode, response)
	}
}
