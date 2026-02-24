package apierrors

import (
	"errors"
	"etruscan/internal/domain/models"

	"github.com/jackc/pgx/v5/pgconn"
)

// dbErrorToApiError attempts to convert low-level PostgreSQL constraint errors
// into structured API errors that the global handler can serialize properly.
func dbErrorToApiError(err error) *models.ApiError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	case "23505": // unique_violation
		switch pgErr.ConstraintName {
		// users.email UNIQUE
		case "users_email_key":
			return models.NewApiError(
				models.ErrCodeEmailAlreadyExists,
				"User with this email already exists",
				map[string]interface{}{
					"constraint": pgErr.ConstraintName,
				},
				[]models.FieldError{
					{
						Field: "email",
						Issue: "must be unique",
					},
				},
				pgErr,
			)

		// everything else that is a unique violation
		default:
			return models.NewApiError(
				models.ErrCodeValidationFailed,
				"Unique constraint violated",
				map[string]interface{}{
					"constraint": pgErr.ConstraintName,
				},
				nil,
				pgErr,
			)
		}

	// 23503: foreign_key_violation
	// 23514: check_violation
	// 23502: not_null_violation
	case "23503", "23514", "23502":
		return models.NewApiError(
			models.ErrCodeValidationFailed,
			"Database constraint violated",
			map[string]interface{}{
				"constraint": pgErr.ConstraintName,
				"code":       pgErr.Code,
			},
			nil,
			pgErr,
		)
	default:
		return nil
	}
}
