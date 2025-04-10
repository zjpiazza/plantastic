package models

import "time"

// Task represents a task in the system
type Task struct {
	ID          string    `json:"id"`
	GardenID    string    `json:"garden_id"`     // Foreign key to Garden
	BedID       *string   `json:"garden_bed_id"` // Foreign key to Bed (nullable)
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
