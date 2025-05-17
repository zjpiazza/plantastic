package models

import (
	"time"

	"github.com/google/uuid"
)

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

// Task status constants
const (
	TaskStatusPending    = "Pending"
	TaskStatusInProgress = "In Progress"
	TaskStatusCompleted  = "Completed"
	TaskStatusOverdue    = "Overdue"
	TaskStatusCancelled  = "Cancelled"
)

// Priority levels
const (
	PriorityLow    = "Low"
	PriorityMedium = "Medium"
	PriorityHigh   = "High"
)

// NewTask creates a new Task with default values
func NewTask(gardenID string, bedID *string, description string, dueDate time.Time, status string, priority string) Task {
	now := time.Now()

	// Set default status if empty
	if status == "" {
		status = TaskStatusPending
	}

	// Set default priority if empty
	if priority == "" {
		priority = PriorityMedium
	}

	return Task{
		ID:          uuid.New().String(),
		GardenID:    gardenID,
		BedID:       bedID,
		Description: description,
		DueDate:     dueDate,
		Status:      status,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
