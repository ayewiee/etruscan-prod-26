package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var BatchOperationError = errors.New("batch operation error: not all rows were inserted")

func IsUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
