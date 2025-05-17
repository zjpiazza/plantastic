package storage

import (
	"errors"
	"time"

	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

// BedStorer defines the interface for bed data operations.
type BedStorer interface {
	GetAllBeds() ([]models.Bed, error)
	GetBedByID(bedID string) (models.Bed, error)
	CreateBed(bed *models.Bed) error
	UpdateBed(bed *models.Bed) error
	DeleteBed(bedID string) error
	GetBedsByGardenID(gardenID string) ([]models.Bed, error)
}

// GormBedStore implements BedStorer using GORM.
type GormBedStore struct {
	db *gorm.DB
}

// NewGormBedStore creates a new GormBedStore.
func NewGormBedStore(db *gorm.DB) BedStorer {
	return &GormBedStore{db: db}
}

func (s *GormBedStore) GetAllBeds() ([]models.Bed, error) {
	var beds []models.Bed
	result := s.db.Find(&beds)
	if result.Error != nil {
		return nil, ErrDatabase
	}
	return beds, nil
}

func (s *GormBedStore) GetBedByID(bedID string) (models.Bed, error) {
	var bed models.Bed
	result := s.db.Where("id = ?", bedID).First(&bed)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Bed{}, ErrRecordNotFound
		}
		return models.Bed{}, ErrDatabase
	}
	return bed, nil
}

func (s *GormBedStore) CreateBed(bed *models.Bed) error {
	if bed.Name == "" || bed.GardenID == "" {
		return ErrValidation
	}

	// Check if referenced garden exists
	var garden models.Garden
	if err := s.db.First(&garden, "id = ?", bed.GardenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrValidation // Garden not found is a validation issue for creating a bed
		}
		return ParseDatabaseError(err) // Other DB error
	}

	result := s.db.Create(bed)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	return nil
}

func (s *GormBedStore) UpdateBed(bed *models.Bed) error {
	// Validate essential fields first
	if bed.ID == "" { // ID must be present for an update
		return ErrValidation
	}
	if bed.Name == "" || bed.GardenID == "" { // Name and GardenID are mandatory
		return ErrValidation
	}

	// Check if the bed to be updated actually exists
	var existingBed models.Bed
	if err := s.db.First(&existingBed, "id = ?", bed.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return ParseDatabaseError(err) // Other DB error during existence check
	}

	// Check if the referenced garden exists (if GardenID is being changed or just to be sure)
	var garden models.Garden
	if err := s.db.First(&garden, "id = ?", bed.GardenID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrValidation // Referencing a non-existent garden is a validation issue
		}
		return ParseDatabaseError(err) // Other DB error during garden check
	}

	// Use a map for updates to only change specified fields and handle zero values correctly.
	// Ensure 'updated_at' is set.
	updateFields := map[string]interface{}{
		"name":       bed.Name,
		"garden_id":  bed.GardenID,
		"type":       bed.Type,
		"size":       bed.Size,
		"soil_type":  bed.SoilType,
		"notes":      bed.Notes,
		"updated_at": time.Now(),
	}

	result := s.db.Model(&models.Bed{}).Where("id = ?", bed.ID).Updates(updateFields)
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	// GORM's Updates method with a map will not return ErrRecordNotFound if the record doesn't exist.
	// The check for result.RowsAffected == 0 is crucial here.
	if result.RowsAffected == 0 {
		return ErrRecordNotFound // Should have been caught by the .First call, but as a safeguard.
	}
	return nil
}

func (s *GormBedStore) DeleteBed(bedID string) error {
	result := s.db.Where("id = ?", bedID).Delete(&models.Bed{})
	if result.Error != nil {
		return ParseDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (s *GormBedStore) GetBedsByGardenID(gardenID string) ([]models.Bed, error) {
	var beds []models.Bed
	result := s.db.Where("garden_id = ?", gardenID).Find(&beds)
	if result.Error != nil {
		// Do not treat gorm.ErrRecordNotFound as an error here,
		// an empty slice is a valid result if no beds are found for that garden.
		// GORM .Find() does not return gorm.ErrRecordNotFound if no records are found, it just returns an empty slice.
		// So, a specific check for gorm.ErrRecordNotFound is usually not needed for .Find().
		// Any error here would be an actual database issue.
		return nil, ErrDatabase // For other database errors
	}
	return beds, nil
}
