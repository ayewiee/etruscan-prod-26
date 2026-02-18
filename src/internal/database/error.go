package database

import "errors"

var BatchOperationError = errors.New("batch operation error: not all rows were inserted")
