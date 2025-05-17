package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

const (
	// UserCodeDuration defines how long a user code is valid for activation.
	UserCodeDuration = 15 * time.Minute
	// UserCodeFormat is the format string for generating user codes.
	UserCodeFormat = "PLANT-%s"
)

// Manager handles the device linking flow using a database.
type Manager struct {
	db *gorm.DB
}

// NewManager creates a new device manager with a database connection.
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// NewCode generates a new user code for a given deviceID, stores it in the database,
// and returns the user code.
func (m *Manager) NewCode(deviceID string) (string, error) {
	if deviceID == "" {
		return "", fmt.Errorf("deviceID cannot be empty")
	}

	userCode := fmt.Sprintf(UserCodeFormat, strings.ToUpper(uuid.New().String()[:6]))
	expiresAt := time.Now().Add(UserCodeDuration)

	deviceAuth := models.Device{
		// ID will be set by BeforeCreate hook
		DeviceID:  deviceID,
		UserCode:  userCode,
		ExpiresAt: expiresAt,
	}

	if err := m.db.Create(&deviceAuth).Error; err != nil {
		// TODO: Handle potential UserCode collision more gracefully if relying on DB uniqueness,
		// though UUID[:6] makes it rare. Could retry with a new code.
		return "", fmt.Errorf("failed to create device auth record: %w", err)
	}
	return userCode, nil
}

// ValidateUserCode checks if a user code is valid (exists and not expired).
// It returns the device record if valid.
func (m *Manager) ValidateUserCode(userCode string) (*models.Device, error) {
	var deviceAuth models.Device
	err := m.db.Where("user_code = ?", userCode).First(&deviceAuth).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid user code")
		}
		return nil, fmt.Errorf("database error validating user code: %w", err)
	}

	if time.Now().After(deviceAuth.ExpiresAt) {
		// Optionally, clean up expired codes here or via a background job.
		return nil, fmt.Errorf("user code expired")
	}

	if deviceAuth.Token != "" {
		// Code has already been used to successfully link and get a token.
		// Depending on desired behavior, this could be an error or still considered valid.
		// For now, let's treat it as an error for subsequent validation attempts once a token is set.
		return nil, fmt.Errorf("user code already activated")
	}

	return &deviceAuth, nil
}

// SetTokenForUserCode associates a session token with a user code if the code is valid.
func (m *Manager) SetTokenForUserCode(userCode string, token string) error {
	if userCode == "" || token == "" {
		return fmt.Errorf("userCode and token cannot be empty")
	}

	// Validate the code first (exists, not expired, not already activated with a token)
	// ValidateUserCode now returns an error if already activated, so we don't need to check Token != "" here
	deviceAuth, err := m.ValidateUserCode(userCode)
	if err != nil {
		return fmt.Errorf("cannot set token: %w", err)
	}
	// If ValidateUserCode allowed already activated codes, we would check deviceAuth.Token == "" here.

	result := m.db.Model(&deviceAuth).Where("user_code = ?", userCode).Update("token", token)
	if result.Error != nil {
		return fmt.Errorf("failed to update token for user code %s: %w", userCode, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no device record found for user code %s to update token (should not happen after ValidateUserCode)", userCode)
	}
	return nil
}

// GetTokenForUserCode retrieves the session token for a user code if it has been set and the code is valid.
func (m *Manager) GetTokenForUserCode(userCode string) (string, error) {
	var deviceAuth models.Device
	err := m.db.Where("user_code = ?", userCode).First(&deviceAuth).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("invalid user code")
		}
		return "", fmt.Errorf("database error retrieving token: %w", err)
	}

	// It's okay if it's "expired" for polling, as long as a token was set before expiry.
	// However, if a token was never set, and it's now past ExpiresAt, the user can't activate it anymore.
	if deviceAuth.Token == "" && time.Now().After(deviceAuth.ExpiresAt) {
		return "", fmt.Errorf("user code expired and was not activated")
	}

	if deviceAuth.Token == "" {
		return "", fmt.Errorf("token not yet available for this user code") // Or a specific error/nil to indicate pending
	}

	return deviceAuth.Token, nil
}
