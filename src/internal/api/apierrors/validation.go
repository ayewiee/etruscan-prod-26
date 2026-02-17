package apierrors

import (
	"errors"
	"etruscan/internal/domain/models"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func ValidationError(err error, req interface{}) *models.ApiError {
	var verrs validator.ValidationErrors

	if ok := errors.As(err, &verrs); !ok {
		return models.NewApiError(
			models.ErrCodeInternal,
			"Internal validation error",
			nil,
			nil,
			err,
		)
	}

	var fieldErrors []models.FieldError

	for _, verr := range verrs {
		field := verr.Field()
		issue := validationErrorMessage(verr)
		rejectedValue := verr.Value()

		fieldErrors = append(fieldErrors, models.FieldError{
			Field:         field,
			Issue:         issue,
			RejectedValue: rejectedValue,
		})
	}

	return models.NewApiError(
		models.ErrCodeValidationFailed,
		"Некоторые поля не прошли валидацию",
		nil,
		fieldErrors,
		verrs,
	)
}

func DumbValidationError(field string, value interface{}, msg string, cause error) (err *models.ApiError) {
	err = MultipleDumbValidationErrors(models.FieldError{
		Field:         field,
		Issue:         msg,
		RejectedValue: value,
	})
	err.Cause = cause
	return
}

func MultipleDumbValidationErrors(fieldErrs ...models.FieldError) *models.ApiError {
	return models.NewApiError(
		models.ErrCodeValidationFailed,
		"Некоторые поля не прошли валидацию",
		nil,
		fieldErrs,
		nil,
	)
}

// validationErrorMessage returns a human-readable message based on the tag.
func validationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min", "gte":
		return fmt.Sprintf("must be >= %s", fe.Param())
	case "max", "lte":
		return fmt.Sprintf("must be <= %s", fe.Param())
	default:
		return fmt.Sprintf("failed validation (%s)", fe.Tag())
	}
}
