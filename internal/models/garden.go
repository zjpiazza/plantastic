package models

import (
	"time"

	"github.com/google/uuid"
)

// Garden represents a garden in the system
type Garden struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewGarden creates a new Garden with default values
func NewGarden(name, location, description string) Garden {
	now := time.Now()
	return Garden{
		ID:          uuid.New().String(),
		Name:        name,
		Location:    location,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
