package storage

import (
	"errors"
	"time"

	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

// TaskStorer defines the interface for task data operations.
type TaskStorer interface {
	GetAllTasks() ([]models.Task, error)
	GetTaskByID(taskID string) (models.Task, error)
	CreateTask(task *models.Task) error
	UpdateTask(task *models.Task) error
	DeleteTask(taskID string) error
}

// GormTaskStore implements TaskStorer using GORM.
type GormTaskStore struct {
	db *gorm.DB
}

// NewGormTaskStore creates a new GormTaskStore.
func NewGormTaskStore(db *gorm.DB) TaskStorer {
	return &GormTaskStore{db: db}
}

func (s *GormTaskStore) GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task
	result := s.db.Find(&tasks)
	if result.Error != nil {
		// Use custom ErrDatabase for consistency, assuming result.Error is a generic DB error
		return nil, ErrDatabase
	}
	return tasks, nil
}

func (s *GormTaskStore) CreateTask(task *models.Task) error {
	// Basic validation (can be expanded based on model requirements)
	if task.Description == "" || task.GardenID == "" { // Assuming Description and GardenID are mandatory
		return ErrValidation
	}
	// Potentially check if GardenID (and BedID if not nil) exist
	var garden models.Garden
	if err := s.db.First(&garden, "id = ?", task.GardenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrValidation // Referencing a non-existent garden
		}
		return ParseDatabaseError(err)
	}
	if task.BedID != nil && *task.BedID != "" {
		var bed models.Bed
		if err := s.db.First(&bed, "id = ? AND garden_id = ?", *task.BedID, task.GardenID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrValidation // Referencing a non-existent bed or bed not in the specified garden
			}
			return ParseDatabaseError(err)
		}
	}

	result := s.db.Create(task)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

func (s *GormTaskStore) GetTaskByID(taskID string) (models.Task, error) {
	var task models.Task
	result := s.db.Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Task{}, ErrRecordNotFound
		}
		return models.Task{}, ErrDatabase // Use custom error
	}
	return task, nil
}

func (s *GormTaskStore) UpdateTask(task *models.Task) error {
	if task.ID == "" { // ID must be present for an update
		return ErrValidation
	}
	if task.Description == "" || task.GardenID == "" { // Assuming Description and GardenID are mandatory
		return ErrValidation
	}

	// Check if the task to be updated actually exists
	var existingTask models.Task
	if err := s.db.First(&existingTask, "id = ?", task.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return ParseDatabaseError(err)
	}

	// Potentially check if referenced GardenID (and BedID if not nil) exist if they are being changed
	if task.GardenID != existingTask.GardenID { // If GardenID is part of the update
		var garden models.Garden
		if err := s.db.First(&garden, "id = ?", task.GardenID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrValidation // Referencing a non-existent garden
			}
			return ParseDatabaseError(err)
		}
	}
	if task.BedID != nil && (existingTask.BedID == nil || *task.BedID != *existingTask.BedID) {
		var bed models.Bed
		if err := s.db.First(&bed, "id = ? AND garden_id = ?", *task.BedID, task.GardenID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrValidation // Referencing a non-existent bed or bed not in the specified garden
			}
			return ParseDatabaseError(err)
		}
	} else if task.BedID == nil && existingTask.BedID != nil { // BedID is being set to null
		// This is allowed, but if there was a check to ensure BedID always belongs to GardenID, it would be here.
	}

	// Use a map for updates for explicit field changes and correct handling of zero values.
	updateFields := map[string]interface{}{
		"description": task.Description,
		"due_date":    task.DueDate,
		"status":      task.Status,
		"priority":    task.Priority,
		"garden_id":   task.GardenID,
		"bed_id":      task.BedID, // Can be nil to clear the association
		"updated_at":  time.Now(),
	}

	result := s.db.Model(&models.Task{}).Where("id = ?", task.ID).Updates(updateFields)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound // Should have been caught by First, but safeguard
	}
	return nil
}

func (s *GormTaskStore) DeleteTask(taskID string) error {
	result := s.db.Where("id = ?", taskID).Delete(&models.Task{})
	if result.Error != nil {
		return ParseDatabaseError(result.Error) // Use custom error parser
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
