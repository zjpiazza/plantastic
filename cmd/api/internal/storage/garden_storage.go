package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

// GardenStorer defines the interface for garden data operations.
type GardenStorer interface {
	GetAllGardens() ([]models.Garden, error)
	CreateGarden(garden *models.Garden) error
	GetGardenByID(gardenID string) (models.Garden, error)
	UpdateGarden(garden *models.Garden) error
	DeleteGarden(gardenID string) error
	CreateGardenWithTransaction(garden *models.Garden, beds []models.Bed) error
	GetAllGardensWithTimeout(timeout time.Duration) ([]models.Garden, error)
	GetGardensByQuery(params map[string]string) ([]models.Garden, error)
}

// GormGardenStore implements GardenStorer using GORM.
type GormGardenStore struct {
	db *gorm.DB
}

// NewGormGardenStore creates a new GormGardenStore.
// It's good practice for the constructor to return the interface type.
func NewGormGardenStore(db *gorm.DB) GardenStorer {
	return &GormGardenStore{db: db}
}

func (s *GormGardenStore) GetAllGardens() ([]models.Garden, error) {
	var gardens []models.Garden
	result := s.db.Find(&gardens)
	if result.Error != nil {
		return nil, ErrDatabase
	}
	return gardens, nil
}

func (s *GormGardenStore) CreateGarden(garden *models.Garden) error {
	if garden.Name == "" {
		return ErrValidation
	}
	result := s.db.Create(garden)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

func (s *GormGardenStore) GetGardenByID(gardenID string) (models.Garden, error) {
	var garden models.Garden
	result := s.db.Where("id = ?", gardenID).First(&garden)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Garden{}, ErrRecordNotFound
		}
		return models.Garden{}, ErrDatabase
	}
	return garden, nil
}

func (s *GormGardenStore) UpdateGarden(garden *models.Garden) error {
	// First, check if the record exists to return ErrRecordNotFound if it doesn't.
	// GORM's Updates method might not return an error for non-existent records if using a map or struct,
	// depending on the configuration and whether primary keys are set.
	var existingGarden models.Garden
	if err := s.db.First(&existingGarden, "id = ?", garden.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return ParseDatabaseError(err) // Could be other DB errors
	}

	// Validate before update (e.g., name not empty)
	if garden.Name == "" {
		return ErrValidation
	}

	// Use a map to specify which fields to update, preventing zero values from overwriting fields unintentionally.
	// Only update fields that are meant to be updatable.
	updateFields := map[string]interface{}{
		"name":        garden.Name,
		"description": garden.Description,
		"location":    garden.Location,
		"updated_at":  time.Now(), // Explicitly set updated_at
	}

	result := s.db.Model(&models.Garden{}).Where("id = ?", garden.ID).Updates(updateFields)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		// This case might indicate the record was deleted between the First() call and Updates(),
		// or perhaps the update data was identical to existing data (though Updates usually still reports 1 row affected if matched).
		// Considering we already checked existence, this might be redundant or indicate a concurrent modification.
		// For now, we can rely on the earlier First() check for ErrRecordNotFound.
	}
	return nil
}

func (s *GormGardenStore) DeleteGarden(gardenID string) error {
	result := s.db.Where("id = ?", gardenID).Delete(&models.Garden{})
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// CreateGardenWithTransaction creates a garden with related structures in a transaction
// TODO: This method needs to be adapted. If CreateBed becomes part of a BedStorer interface,
// this GormGardenStore might need a BedStorer instance, or the transaction logic
// needs to be handled at a higher service layer that coordinates both stores.
// For now, assuming CreateBed is a package-level func in storage that can take a *gorm.DB (tx).
func (s *GormGardenStore) CreateGardenWithTransaction(garden *models.Garden, beds []models.Bed) error {
	var innerError error
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Temporarily create a new GormGardenStore with the transaction tx for the CreateGarden call
		// This is not ideal if CreateGarden itself has complex logic relying on the 's' state, but for now:
		tempStoreForTx := &GormGardenStore{db: tx}
		if err := tempStoreForTx.CreateGarden(garden); err != nil { // Call the method on the store
			return err
		}

		// For CreateBed, we're assuming it's a package-level function for now.
		// If CreateBed is also refactored to be a method on a BedStore, this will need adjustment.
		// One option is that NewGormGardenStore could also accept a BedStore factory or instance.
		for i := range beds {
			beds[i].GardenID = garden.ID
			// Instantiate a BedStorer with the transaction
			tempBedStoreForTx := NewGormBedStore(tx)
			if err := tempBedStoreForTx.CreateBed(&beds[i]); err != nil {
				innerError = err
				return err
			}
		}
		return nil
	})

	if innerError != nil && (strings.Contains(innerError.Error(), "deadlock") ||
		strings.Contains(innerError.Error(), "Deadlock")) {
		return ErrTransactionFailed
	}
	if err != nil {
		if strings.Contains(err.Error(), "deadlock") ||
			strings.Contains(err.Error(), "Deadlock") {
			return ErrTransactionFailed
		}
		return ParseDatabaseError(err)
	}
	return nil
}

func (s *GormGardenStore) GetAllGardensWithTimeout(timeout time.Duration) ([]models.Garden, error) {
	var gardens []models.Garden
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := s.db.WithContext(ctx).Find(&gardens)
	if result.Error != nil {
		if errors.Is(result.Error, context.DeadlineExceeded) {
			return nil, ErrTimeout
		}
		return nil, ErrDatabase
	}
	return gardens, nil
}

func (s *GormGardenStore) GetGardensByQuery(params map[string]string) ([]models.Garden, error) {
	var gardens []models.Garden
	allowedParams := map[string]bool{
		"name": true, "createdStart": true, "createdEnd": true, "size": true,
	}
	for key := range params {
		if !allowedParams[key] {
			return nil, ErrInvalidQuery
		}
	}

	query := s.db
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

// Note: The CreateBed function is assumed to exist elsewhere in the storage package for CreateGardenWithTransaction.
// If CreateBed is also part of a BedStorer interface, the transaction logic will need further refactoring.
// For example, the component initiating the transaction might manage both GardenStorer and BedStorer.
