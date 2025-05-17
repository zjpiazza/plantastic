package storage

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound    = errors.New("record not found")
	ErrValidation        = errors.New("validation failed")
	ErrDatabase          = errors.New("database error")
	ErrConflict          = errors.New("conflict error")
	ErrTimeout           = errors.New("timeout error")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrInvalidQuery      = errors.New("invalid query parameter")
)

// ParseDatabaseError translates GORM and database driver errors into custom storage errors.
func ParseDatabaseError(err error) error {
	if err == nil {
		return nil
	}

	// Check for GORM specific errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	// Check for common error messages (e.g., unique constraint for SQLite)
	// This might need to be adjusted based on the specific database driver and its error messages.
	// For SQLite, a unique constraint violation often includes "UNIQUE constraint failed".
	// For PostgreSQL, it might be "duplicate key value violates unique constraint".
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "unique constraint failed") || strings.Contains(errStr, "duplicate key") {
		return ErrConflict
	}

	// Default to a generic database error
	return ErrDatabase
}
