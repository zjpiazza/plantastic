package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Device represents a device authentication record in the database.
type Device struct {
	ID         string     `gorm:"type:uuid;primary_key;" json:"id"`      // Unique identifier for the device record
	UserCode   string     `gorm:"uniqueIndex;not null" json:"user_code"` // The code shown to the user (e.g., PLANT-XXXXXX)
	DeviceID   string     `gorm:"uniqueIndex;not null" json:"device_id"` // A more permanent ID for the device being linked (could be user-generated or auto) - Placeholder for now
	Token      string     `json:"token"`                                 // The session token once authenticated
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ExpiresAt  time.Time  `json:"expires_at"`             // When the user_code or linking process expires
	LastUsedAt *time.Time `json:"last_used_at,omitempty"` // Optional: track when the token was last used
	// UserID    string    `json:"user_id"` // Optional: to link to a user if you have a users table
}

// BeforeCreate will set a UUID for the ID if it's not set.
func (device *Device) BeforeCreate(tx *gorm.DB) (err error) {
	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	return
}
