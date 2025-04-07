package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

func GetAllGardens(db *gorm.DB) ([]models.Garden, error) {
	var gardens []models.Garden
	result := db.Find(&gardens)
	if result.Error != nil {
		return nil, ErrDatabase
	}
	return gardens, nil
}

func CreateGarden(db *gorm.DB, garden *models.Garden) error {
	if garden.Name == "" {
		return ErrValidation
	}

	result := db.Create(garden)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

func GetGardenByID(db *gorm.DB, gardenID string) (models.Garden, error) {
	var garden models.Garden
	result := db.Where("id = ?", gardenID).First(&garden)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Garden{}, ErrRecordNotFound
		}
		return models.Garden{}, ErrDatabase
	}
	return garden, nil
}

func UpdateGarden(db *gorm.DB, garden *models.Garden) error {
	if garden.Name == "" {
		return ErrValidation
	}

	var existingGarden models.Garden
	if err := db.First(&existingGarden, "id = ?", garden.ID).Error; err != nil {
		return ParseDatabaseError(err)
	}

	result := db.Save(garden)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

func DeleteGarden(db *gorm.DB, gardenID string) error {
	result := db.Where("id = ?", gardenID).Delete(&models.Garden{})
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// CreateGardenWithTransaction creates a garden with related structures in a transaction
func CreateGardenWithTransaction(db *gorm.DB, garden *models.Garden, beds []models.Bed) error {
	// Store the original errors that might be lost in transaction wrapping
	var innerError error

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := CreateGarden(tx, garden); err != nil {
			return err
		}

		for i := range beds {
			beds[i].GardenID = garden.ID
			if err := CreateBed(tx, &beds[i]); err != nil {
				// Save the inner error before it gets wrapped
				innerError = err
				return err
			}
		}

		return nil
	})

	// First check the inner error which has more detailed information
	if innerError != nil && (strings.Contains(innerError.Error(), "deadlock") ||
		strings.Contains(innerError.Error(), "Deadlock")) {
		return ErrTransactionFailed
	}

	// Then check the transaction error
	if err != nil {
		if strings.Contains(err.Error(), "deadlock") ||
			strings.Contains(err.Error(), "Deadlock") {
			return ErrTransactionFailed
		}
		return ParseDatabaseError(err)
	}

	return nil
}

// GetAllGardensWithTimeout gets all gardens with timeout detection
func GetAllGardensWithTimeout(db *gorm.DB, timeout time.Duration) ([]models.Garden, error) {
	var gardens []models.Garden

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use context with the query
	result := db.WithContext(ctx).Find(&gardens)

	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, ErrDatabase
	}

	return gardens, nil
}

// GetGardensByQuery retrieves gardens filtered by query parameters
func GetGardensByQuery(db *gorm.DB, params map[string]string) ([]models.Garden, error) {
	var gardens []models.Garden

	// Validate query parameters
	allowedParams := map[string]bool{"name": true, "createdStart": true, "createdEnd": true, "size": true}
	for key := range params {
		if !allowedParams[key] {
			return nil, ErrInvalidQuery
		}
	}

	query := db

	// Apply filters based on valid parameters
	if name, ok := params["name"]; ok {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if createdStart, ok := params["createdStart"]; ok {
		if _, err := time.Parse(time.RFC3339, createdStart); err != nil {
			return nil, ErrValidation
		}
		query = query.Where("created_at >= ?", createdStart)
	}

	if createdEnd, ok := params["createdEnd"]; ok {
		if _, err := time.Parse(time.RFC3339, createdEnd); err != nil {
			return nil, ErrValidation
		}
		query = query.Where("created_at <= ?", createdEnd)
	}

	result := query.Find(&gardens)
	if result.Error != nil {
		return nil, ErrDatabase
	}

	return gardens, nil
}
