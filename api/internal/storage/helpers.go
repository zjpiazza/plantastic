package storage

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ParseDatabaseError examines a database error and returns a more specific error type
func ParseDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())

	// Check for specific error types based on error messages (case insensitive)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrRecordNotFound

	case strings.Contains(errMsg, "deadlock") || strings.Contains(errMsg, "deadlock found"):
		return ErrTransactionFailed

	case strings.Contains(errMsg, "connection"):
		return ErrDatabase

	case strings.Contains(errMsg, "timeout"):
		return ErrTimeout

	case strings.Contains(errMsg, "unique constraint failed"):
		return ErrConflict

	case strings.Contains(errMsg, "foreign key constraint failed"):
		return ErrForeignKeyViolation

	default:
		return ErrDatabase
	}
}

// IsTemporaryError checks if the error is temporary and the operation could be retried
func IsTemporaryError(err error) bool {
	return errors.Is(err, ErrTimeout) ||
		errors.Is(err, ErrTransactionFailed) ||
		(errors.Is(err, ErrDatabase) &&
			strings.Contains(err.Error(), "connection"))
}
