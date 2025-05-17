package models

import (
	"time"

	"github.com/google/uuid"
)

// Bed represents a garden bed
type Bed struct {
	ID        string    `json:"id"`
	GardenID  string    `json:"garden_id"` // Foreign key to Garden
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      string    `json:"size"`
	SoilType  string    `json:"soil_type"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBed creates a new Bed with default values
func NewBed(gardenID, name, bedType, size, soilType, notes string) Bed {
	now := time.Now()
	return Bed{
		ID:        uuid.New().String(),
		GardenID:  gardenID,
		Name:      name,
		Type:      bedType,
		Size:      size,
		SoilType:  soilType,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
