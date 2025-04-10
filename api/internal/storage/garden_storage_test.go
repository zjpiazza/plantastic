package storage_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/zjpiazza/plantastic/api/internal/storage"
	"github.com/zjpiazza/plantastic/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates and returns an in-memory test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate schemas
	db.AutoMigrate(&models.Garden{}, &models.Bed{})

	return db
}

// seedTestData adds sample data to the test database
func seedTestData(db *gorm.DB) {
	gardens := []models.Garden{
		{ID: "g1", Name: "Garden 1"},
		{ID: "g2", Name: "Garden 2"},
	}

	beds := []models.Bed{
		{ID: "b1", Name: "Bed 1", GardenID: "g1"},
		{ID: "b2", Name: "Bed 2", GardenID: "g1"},
	}

	db.Create(&gardens)
	db.Create(&beds)
}

func TestGetAllBeds(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&models.Bed{})

	// Seed the database with test data
	db.Create(&models.Bed{ID: "b1", Name: "Test Bed 1", GardenID: "g1"})
	db.Create(&models.Bed{ID: "b2", Name: "Test Bed 2", GardenID: "g1"})

	// Call the function
	beds, err := storage.GetAllBeds(db)
	if err != nil {
		t.Fatalf("Failed to get beds: %v", err)
	}

	// Assert the results
	if len(beds) != 2 {
		t.Errorf("Expected 2 beds, got %d", len(beds))
	}
	if beds[0].Name != "Test Bed 1" {
		t.Errorf("Expected first bed name to be 'Test Bed 1', got '%s'", beds[0].Name)
	}
}

// TestGetBedByID tests the GetBedByID function
func TestGetBedByID(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&models.Bed{})

	// Test case: Bed exists
	db.Create(&models.Bed{ID: "b1", Name: "Test Bed 1", GardenID: "g1"})

	bed, err := storage.GetBedByID(db, "b1")
	if err != nil {
		t.Fatalf("Failed to get bed: %v", err)
	}

	if bed.ID != "b1" || bed.Name != "Test Bed 1" {
		t.Errorf("Got incorrect bed data: %v", bed)
	}

	// Test case: Bed does not exist
	_, err = storage.GetBedByID(db, "nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent bed, got nil")
	}
}

// TestCreateBed tests the CreateBed function
// TestUpdateBed tests the UpdateBed function
// TestDeleteBed tests the DeleteBed function

func TestGetGardenByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := storage.GetGardenByID(db, "nonexistent")
	if !errors.Is(err, storage.ErrRecordNotFound) {
		t.Errorf("Expected ErrRecordNotFound, got %v", err)
	}
}

func TestCreateBed_ForeignKeyViolation(t *testing.T) {
	db := setupTestDB(t)

	// Try to create a bed with a non-existent garden ID
	bed := models.Bed{
		ID:       "b1",
		Name:     "Test Bed",
		GardenID: "nonexistent",
	}

	err := storage.CreateBed(db, &bed)
	if !errors.Is(err, storage.ErrValidation) {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestDatabaseErrors(t *testing.T) {
	db := setupTestDB(t)

	// Create a hook that simulates a database error
	db.Callback().Query().Before("gorm:query").Register("simulate_error", func(db *gorm.DB) {
		db.AddError(fmt.Errorf("simulated database error"))
	})

	_, err := storage.GetAllGardens(db)
	if !errors.Is(err, storage.ErrDatabase) {
		t.Errorf("Expected ErrDatabase, got %v", err)
	}
}

func TestCreateGarden_ValidationError(t *testing.T) {
	db := setupTestDB(t)

	// Add validation to your models or in your storage methods
	// For example, require garden name to not be empty
	garden := models.Garden{ID: "g1", Name: ""}

	err := storage.CreateGarden(db, &garden)
	if !errors.Is(err, storage.ErrValidation) {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}

func TestUpdateGarden_Errors(t *testing.T) {
	tests := []struct {
		name    string
		garden  models.Garden
		setupDB func(*gorm.DB)
		wantErr error
	}{
		{
			name:    "garden not found",
			garden:  models.Garden{ID: "nonexistent", Name: "Test"},
			setupDB: func(db *gorm.DB) {},
			wantErr: storage.ErrRecordNotFound,
		},
		{
			name:   "validation error",
			garden: models.Garden{ID: "g1", Name: ""},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Original"})
			},
			wantErr: storage.ErrValidation,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			err := storage.UpdateGarden(db, &tt.garden)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// Garden CRUD Test Tables
func TestGardenCRUD(t *testing.T) {
	// Define test cases for GetGardenByID
	getGardenTests := []struct {
		name     string
		gardenID string
		setupDB  func(*gorm.DB)
		wantErr  error
	}{
		{
			name:     "garden exists",
			gardenID: "g1",
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: nil,
		},
		{
			name:     "garden not found",
			gardenID: "nonexistent",
			setupDB:  func(db *gorm.DB) {},
			wantErr:  storage.ErrRecordNotFound,
		},
		{
			name:     "database error",
			gardenID: "g1",
			setupDB: func(db *gorm.DB) {
				db.Callback().Query().Before("gorm:query").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run GetGardenByID tests
	for _, tt := range getGardenTests {
		t.Run("GetGardenByID_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			_, err := storage.GetGardenByID(db, tt.gardenID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	// Define test cases for CreateGarden
	createGardenTests := []struct {
		name    string
		garden  models.Garden
		setupDB func(*gorm.DB)
		wantErr error
	}{
		{
			name:    "valid garden",
			garden:  models.Garden{ID: "g1", Name: "Garden 1"},
			setupDB: func(db *gorm.DB) {},
			wantErr: nil,
		},
		{
			name:    "validation error - empty name",
			garden:  models.Garden{ID: "g1", Name: ""},
			setupDB: func(db *gorm.DB) {},
			wantErr: storage.ErrValidation,
		},
		{
			name:   "conflict error - duplicate ID",
			garden: models.Garden{ID: "g1", Name: "Garden 1"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Existing Garden"})
			},
			wantErr: storage.ErrConflict,
		},
		{
			name:   "database error",
			garden: models.Garden{ID: "g1", Name: "Garden 1"},
			setupDB: func(db *gorm.DB) {
				db.Callback().Create().Before("gorm:create").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run CreateGarden tests
	for _, tt := range createGardenTests {
		t.Run("CreateGarden_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			err := storage.CreateGarden(db, &tt.garden)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	// Define test cases for UpdateGarden
	updateGardenTests := []struct {
		name    string
		garden  models.Garden
		setupDB func(*gorm.DB)
		wantErr error
	}{
		{
			name:   "valid update",
			garden: models.Garden{ID: "g1", Name: "Updated Garden"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: nil,
		},
		{
			name:    "garden not found",
			garden:  models.Garden{ID: "nonexistent", Name: "Updated Garden"},
			setupDB: func(db *gorm.DB) {},
			wantErr: storage.ErrRecordNotFound,
		},
		{
			name:   "validation error - empty name",
			garden: models.Garden{ID: "g1", Name: ""},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: storage.ErrValidation,
		},
		{
			name:   "database error",
			garden: models.Garden{ID: "g1", Name: "Updated Garden"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
				db.Callback().Update().Before("gorm:update").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run UpdateGarden tests
	for _, tt := range updateGardenTests {
		t.Run("UpdateGarden_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			err := storage.UpdateGarden(db, &tt.garden)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	// Define test cases for DeleteGarden
	deleteGardenTests := []struct {
		name     string
		gardenID string
		setupDB  func(*gorm.DB)
		wantErr  error
	}{
		{
			name:     "valid delete",
			gardenID: "g1",
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: nil,
		},
		{
			name:     "garden not found",
			gardenID: "nonexistent",
			setupDB:  func(db *gorm.DB) {},
			wantErr:  storage.ErrRecordNotFound,
		},
		{
			name:     "database error",
			gardenID: "g1",
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
				db.Callback().Delete().Before("gorm:delete").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run DeleteGarden tests
	for _, tt := range deleteGardenTests {
		t.Run("DeleteGarden_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			err := storage.DeleteGarden(db, tt.gardenID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// Bed CRUD Test Tables
func TestBedCRUD(t *testing.T) {
	// Define test cases for CreateBed
	createBedTests := []struct {
		name    string
		bed     models.Bed
		setupDB func(*gorm.DB)
		wantErr error
	}{
		{
			name: "valid bed",
			bed:  models.Bed{ID: "b1", Name: "Bed 1", GardenID: "g1"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: nil,
		},
		{
			name: "validation error - empty name",
			bed:  models.Bed{ID: "b1", Name: "", GardenID: "g1"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
			},
			wantErr: storage.ErrValidation,
		},
		{
			name:    "foreign key validation",
			bed:     models.Bed{ID: "b1", Name: "Bed 1", GardenID: "nonexistent"},
			setupDB: func(db *gorm.DB) {},
			wantErr: storage.ErrValidation,
		},
		{
			name: "confict_error",
			bed:  models.Bed{ID: "b1", Name: "Bed 1", GardenID: "g1"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
				db.Create(&models.Bed{ID: "b1", Name: "Existing Bed", GardenID: "g1"})
			},
			wantErr: storage.ErrConflict,
		},
		{
			name: "database error",
			bed:  models.Bed{ID: "b1", Name: "Bed 1", GardenID: "g1"},
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
				db.Callback().Create().Before("gorm:create").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run CreateBed tests
	for _, tt := range createBedTests {
		t.Run("CreateBed_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			err := storage.CreateBed(db, &tt.bed)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	// Define test cases for GetBedByID
	getBedTests := []struct {
		name    string
		bedID   string
		setupDB func(*gorm.DB)
		wantErr error
	}{
		{
			name:  "bed exists",
			bedID: "b1",
			setupDB: func(db *gorm.DB) {
				db.Create(&models.Garden{ID: "g1", Name: "Garden 1"})
				db.Create(&models.Bed{ID: "b1", Name: "Bed 1", GardenID: "g1"})
			},
			wantErr: nil,
		},
		{
			name:    "bed not found",
			bedID:   "nonexistent",
			setupDB: func(db *gorm.DB) {},
			wantErr: storage.ErrRecordNotFound,
		},
		{
			name:  "database error",
			bedID: "b1",
			setupDB: func(db *gorm.DB) {
				db.Callback().Query().Before("gorm:query").Register("error", func(db *gorm.DB) {
					db.AddError(fmt.Errorf("database error"))
				})
			},
			wantErr: storage.ErrDatabase,
		},
	}

	// Run GetBedByID tests
	for _, tt := range getBedTests {
		t.Run("GetBedByID_"+tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupDB(db)

			_, err := storage.GetBedByID(db, tt.bedID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected %v, got %v", tt.wantErr, err)
			}
		})
	}

	// Add similar test tables for UpdateBed and DeleteBed
}

// Test for timeout errors
func TestTimeoutErrors(t *testing.T) {
	db := setupTestDB(t)

	// Register a callback that simulates a timeout
	db.Callback().Query().Before("gorm:query").Register("timeout", func(db *gorm.DB) {
		db.AddError(context.DeadlineExceeded)
	})

	_, err := storage.GetAllGardensWithTimeout(db, 100*time.Millisecond)
	if !errors.Is(err, storage.ErrTimeout) {
		t.Errorf("Expected ErrTimeout, got %v", err)
	}
}

// Test for transaction errors
func TestTransactionErrors(t *testing.T) {
	db := setupTestDB(t)

	// Test garden and beds
	garden := models.Garden{ID: "g1", Name: "Garden 1"}
	beds := []models.Bed{
		{ID: "b1", Name: "Bed 1"},
		{ID: "b2", Name: "Bed 2"},
	}

	// Use a more direct approach to simulate deadlocks
	db.Callback().Create().After("gorm:create").Register("deadlock", func(db *gorm.DB) {
		if _, ok := db.Statement.Dest.(*models.Bed); ok {
			db.AddError(fmt.Errorf("Error 1213: Deadlock found when trying to get lock"))
		}
	})

	err := storage.CreateGardenWithTransaction(db, &garden, beds)
	// Add debug output to help diagnose the issue
	if err != nil {
		t.Logf("Got error type: %T, error message: %q", err, err.Error())
	}

	if !errors.Is(err, storage.ErrTransactionFailed) {
		t.Errorf("Expected ErrTransactionFailed, got %v", err)
	}
}

// Test for invalid query parameters
func TestInvalidQueryParameters(t *testing.T) {
	db := setupTestDB(t)

	// Test with invalid parameter
	params := map[string]string{
		"invalid_param": "value",
	}

	_, err := storage.GetGardensByQuery(db, params)
	if !errors.Is(err, storage.ErrInvalidQuery) {
		t.Errorf("Expected ErrInvalidQuery, got %v", err)
	}

	// Test with invalid date format
	params = map[string]string{
		"createdStart": "not-a-date",
	}

	_, err = storage.GetGardensByQuery(db, params)
	if !errors.Is(err, storage.ErrValidation) {
		t.Errorf("Expected ErrValidation, got %v", err)
	}
}
