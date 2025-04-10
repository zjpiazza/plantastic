package storage

import "errors"

// Custom error types for the storage layer
var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrDatabase            = errors.New("database error")
	ErrValidation          = errors.New("validation error")
	ErrConflict            = errors.New("resource conflict")
	ErrForeignKeyViolation = errors.New("foreign key violation")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrInvalidQuery        = errors.New("invalid query parameters")
	ErrOperationFailed     = errors.New("operation failed")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrTimeout             = errors.New("operation timed out")
)
