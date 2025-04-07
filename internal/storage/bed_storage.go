package storage

import (
	"errors"

	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

// GetAllBeds returns all beds from the database
func GetAllBeds(db *gorm.DB) ([]models.Bed, error) {
	var beds []models.Bed
	result := db.Find(&beds)
	if result.Error != nil {
		return nil, ErrDatabase
	}
	return beds, nil
}

// GetBedByID retrieves a bed by its ID
func GetBedByID(db *gorm.DB, bedID string) (models.Bed, error) {
	var bed models.Bed
	result := db.Where("id = ?", bedID).First(&bed)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Bed{}, ErrRecordNotFound
		}
		return models.Bed{}, ErrDatabase
	}
	return bed, nil
}

// CreateBed creates a new bed
func CreateBed(db *gorm.DB, bed *models.Bed) error {
	// Validate bed
	if bed.Name == "" || bed.GardenID == "" {
		return ErrValidation
	}

	// Check if garden exists
	var garden models.Garden
	if err := db.First(&garden, "id = ?", bed.GardenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This is more of a validation error - the referenced garden doesn't exist
			return ErrValidation
		}
		return ErrDatabase
	}

	result := db.Create(bed)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

// UpdateBed updates an existing bed
func UpdateBed(db *gorm.DB, bed *models.Bed) error {
	// Validate bed
	if bed.Name == "" || bed.GardenID == "" {
		return ErrValidation
	}

	// Check if bed exists
	var existingBed models.Bed
	if err := db.First(&existingBed, "id = ?", bed.ID).Error; err != nil {
		return ParseDatabaseError(err)
	}

	// Check if garden exists
	var garden models.Garden
	if err := db.First(&garden, "id = ?", bed.GardenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// This is more of a validation error since we're checking before the actual update
			return ErrValidation
		}
		return ParseDatabaseError(err)
	}

	result := db.Save(bed)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

// DeleteBed deletes a bed by ID
func DeleteBed(db *gorm.DB, bedID string) error {
	result := db.Where("id = ?", bedID).Delete(&models.Bed{})
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
