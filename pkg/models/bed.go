package models

import "time"

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
