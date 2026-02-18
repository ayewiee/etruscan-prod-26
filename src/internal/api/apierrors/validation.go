package apierrors

import (
	"errors"
	"etruscan/internal/domain/models"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// pretty much everything here was taken from my 2nd stage project

func fieldErrorJSONPath(fe validator.FieldError, root any) string {
	ns := fe.StructNamespace()

	parts := strings.Split(ns, ".")
	if len(parts) == 0 {
		return fe.Field()
	}

	// remove root struct name
	parts = parts[1:]

	var path []string
	t := reflect.TypeOf(root)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	for _, p := range parts {
		field, ok := t.FieldByName(p)
		if !ok {
			// fallback to lowercased name
			path = append(path, strings.ToLower(p))
			continue
		}

		// get field name outta json tag
		jsonTag := field.Tag.Get("json")
		name := strings.Split(jsonTag, ",")[0]
		if name == "" || name == "-" {
			name = strings.ToLower(p)
		}

		path = append(path, name)

		// descend into nested struct
		t = field.Type
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
	}

	return strings.Join(path, ".")
}

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
		field := fieldErrorJSONPath(verr, req)
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
		"Some fields are invalid",
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
		"Some fields are invalid",
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
