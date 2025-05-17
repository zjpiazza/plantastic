package storage_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockDBForStorageTest creates a new GORM DB instance with sqlmock for storage tests.
func newMockDBForStorageTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err, "Failed to create sqlmock")

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err, "Failed to open gorm with mock connection")

	return gormDB, mock
}

func TestGormGardenStore_GetAllGardens_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)

	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	expectedGardens := []models.Garden{
		{ID: "g1", Name: "Garden 1", Location: "Loc1", Description: "Desc1", CreatedAt: now, UpdatedAt: now},
		{ID: "g2", Name: "Garden 2", Location: "Loc2", Description: "Desc2", CreatedAt: now, UpdatedAt: now},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "location", "description", "created_at", "updated_at"}).
		AddRow(expectedGardens[0].ID, expectedGardens[0].Name, expectedGardens[0].Location, expectedGardens[0].Description, expectedGardens[0].CreatedAt, expectedGardens[0].UpdatedAt).
		AddRow(expectedGardens[1].ID, expectedGardens[1].Name, expectedGardens[1].Location, expectedGardens[1].Description, expectedGardens[1].CreatedAt, expectedGardens[1].UpdatedAt)

	// Assuming no soft delete for this basic query for now.
	sql := `SELECT * FROM "gardens"`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WillReturnRows(rows)

	actualGardens, err := store.GetAllGardens()

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedGardens, actualGardens)
}

func TestGormGardenStore_GetAllGardens_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)

	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	dbErr := errors.New("database GetAllGardens error")
	sql := `SELECT * FROM "gardens"`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WillReturnError(dbErr)

	actualGardens, err := store.GetAllGardens()

	assert.Nil(t, actualGardens)
	assert.ErrorIs(t, err, storage.ErrDatabase) // Expecting our custom wrapped error
}

func TestGormGardenStore_GetGardenByID_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	gardenID := "g1"
	expectedGarden := models.Garden{ID: gardenID, Name: "Garden 1", Location: "Loc1", Description: "Desc1", CreatedAt: now, UpdatedAt: now}

	rows := sqlmock.NewRows([]string{"id", "name", "location", "description", "created_at", "updated_at"}).
		AddRow(expectedGarden.ID, expectedGarden.Name, expectedGarden.Location, expectedGarden.Description, expectedGarden.CreatedAt, expectedGarden.UpdatedAt)

	// GORM's First() method typically adds a LIMIT 1
	// The regexp.QuoteMeta escapes special characters. The (.*) matches any conditions GORM might add for soft deletes if active.
	sql := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID, 1).WillReturnRows(rows)

	actualGarden, err := store.GetGardenByID(gardenID)

	assert.NoError(t, err)
	assert.Equal(t, expectedGarden, actualGarden)
}

func TestGormGardenStore_GetGardenByID_NotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenID := "nonexistent"

	sql := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	_, err = store.GetGardenByID(gardenID)

	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormGardenStore_GetGardenByID_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenID := "g1"
	dbErr := errors.New("some db error")

	sql := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID, 1).WillReturnError(dbErr)

	_, err = store.GetGardenByID(gardenID)

	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormGardenStore_CreateGarden_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenToCreate := &models.Garden{ID: "g_create_success", Name: "New Garden", Location: "New Loc", Description: "New Desc"}

	mock.ExpectBegin()
	sqlInsert := `INSERT INTO "gardens" ("id","name","location","description","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6)`
	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).
		WithArgs(gardenToCreate.ID, gardenToCreate.Name, gardenToCreate.Location, gardenToCreate.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.CreateGarden(gardenToCreate)
	assert.NoError(t, err)
	// Optionally, assert that gardenToCreate.CreatedAt, UpdatedAt are populated if your CreateGarden method does that.
}

func TestGormGardenStore_CreateGarden_ValidationError(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // Mock is not used here as validation is pre-DB
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()

	// No DB interaction expected, so no mock expectations.
	gardenToCreate := &models.Garden{Name: ""} // Invalid: Name is empty
	err = store.CreateGarden(gardenToCreate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormGardenStore_CreateGarden_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenToCreate := &models.Garden{ID: "g_create_dberror", Name: "Error Garden", Location: "Err Loc", Description: "Err Desc"}
	dbErr := errors.New("create garden db error")

	mock.ExpectBegin()
	sqlInsert := `INSERT INTO "gardens" ("id","name","location","description","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6)`
	mock.ExpectExec(regexp.QuoteMeta(sqlInsert)).
		WithArgs(gardenToCreate.ID, gardenToCreate.Name, gardenToCreate.Location, gardenToCreate.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err = store.CreateGarden(gardenToCreate)

	parsedErr := storage.ParseDatabaseError(dbErr)
	assert.ErrorIs(t, err, parsedErr)
}

func TestGormGardenStore_UpdateGarden_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	gardenToUpdate := &models.Garden{ID: "g1", Name: "Updated Name", Location: "Updated Loc", Description: "Updated Desc", UpdatedAt: now}

	// 1. Mock the First() call to check if record exists
	existingRow := sqlmock.NewRows([]string{"id", "name", "location", "description", "created_at", "updated_at"}).
		AddRow(gardenToUpdate.ID, "Old Name", "Old Loc", "Old Desc", now.Add(-time.Hour), now.Add(-time.Hour))
	sqlSelectOne := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(gardenToUpdate.ID, 1).WillReturnRows(existingRow)

	mock.ExpectBegin()
	sqlUpdate := `UPDATE "gardens" SET "description"=$1,"location"=$2,"name"=$3,"updated_at"=$4 WHERE id = $5`
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).
		WithArgs(gardenToUpdate.Description, gardenToUpdate.Location, gardenToUpdate.Name, sqlmock.AnyArg(), gardenToUpdate.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = store.UpdateGarden(gardenToUpdate)
	assert.NoError(t, err)
}

func TestGormGardenStore_UpdateGarden_ValidationError(t *testing.T) {
	// This test focuses on validation errors.
	// Case 1: Validation error before any DB call (e.g. empty ID if checked by UpdateGarden first)
	// For this, we don't need a mock DB, just the store instance with a nil/dummy DB is fine if no DB methods are hit.
	// However, our current UpdateGarden checks ID then tries a First() call.
	// So, the primary validation test here is for an error *after* the First() call but before Update exec.

	// Test case: Empty Name after fetching existing record
	sqlDBFromMock, mockForSubTest, err := sqlmock.New() // Correctly get all 3 return values
	require.NoError(t, err, "Failed to create sqlmock for sub-test")
	defer sqlDBFromMock.Close()

	dbForTest, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDBFromMock}), &gorm.Config{})
	require.NoError(t, err, "Failed to open gorm for sub-test")
	storeWithMockedDb := storage.NewGormGardenStore(dbForTest) // Use this store

	gardenWithEmptyName := &models.Garden{ID: "g1", Name: ""} // Invalid: Name is empty
	existingRow := sqlmock.NewRows([]string{"id", "name", "location", "description", "created_at", "updated_at"}).
		AddRow("g1", "Old Name", "Old Loc", "Old Desc", time.Now(), time.Now()) // Add all columns expected by First

	sqlSelectOne := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mockForSubTest.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(gardenWithEmptyName.ID, 1).WillReturnRows(existingRow)
	// No EXEC expected as it should fail validation before the update call.

	err = storeWithMockedDb.UpdateGarden(gardenWithEmptyName)
	assert.ErrorIs(t, err, storage.ErrValidation, "Expected validation error for empty name")
	assert.NoError(t, mockForSubTest.ExpectationsWereMet(), "Sub-test mock expectations not met")
}

func TestGormGardenStore_UpdateGarden_NotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenToUpdate := &models.Garden{ID: "nonexistent", Name: "Updated Name"}

	// Mock the First() call to return ErrRecordNotFound
	sqlSelectOne := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(gardenToUpdate.ID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err = store.UpdateGarden(gardenToUpdate)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormGardenStore_UpdateGarden_DBErrorOnSelect(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenToUpdate := &models.Garden{ID: "g1", Name: "Updated Name"}
	dbErr := errors.New("db error on select")

	sqlSelectOne := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(gardenToUpdate.ID, 1).WillReturnError(dbErr)

	err = store.UpdateGarden(gardenToUpdate)
	parsedErr := storage.ParseDatabaseError(dbErr)
	assert.ErrorIs(t, err, parsedErr)
}

func TestGormGardenStore_UpdateGarden_DBErrorOnUpdate(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	gardenToUpdate := &models.Garden{ID: "g1", Name: "Updated Name", Location: "Updated Loc", Description: "Updated Desc", UpdatedAt: now}
	dbUpdateErr := errors.New("db error on update")

	// 1. Mock the First() call
	existingRow := sqlmock.NewRows([]string{"id"}).AddRow(gardenToUpdate.ID)
	sqlSelectOne := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(gardenToUpdate.ID, 1).WillReturnRows(existingRow)

	mock.ExpectBegin()
	sqlUpdate := `UPDATE "gardens" SET "description"=$1,"location"=$2,"name"=$3,"updated_at"=$4 WHERE id = $5`
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).
		WithArgs(gardenToUpdate.Description, gardenToUpdate.Location, gardenToUpdate.Name, sqlmock.AnyArg(), gardenToUpdate.ID).
		WillReturnError(dbUpdateErr)
	mock.ExpectRollback()

	err = store.UpdateGarden(gardenToUpdate)
	parsedErr := storage.ParseDatabaseError(dbUpdateErr)
	assert.ErrorIs(t, err, parsedErr)
}

func TestGormGardenStore_DeleteGarden_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenIDToDelete := "g1"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "gardens" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(gardenIDToDelete).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = store.DeleteGarden(gardenIDToDelete)
	assert.NoError(t, err)
}

func TestGormGardenStore_DeleteGarden_NotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenIDToDelete := "nonexistent"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "gardens" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(gardenIDToDelete).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = store.DeleteGarden(gardenIDToDelete)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormGardenStore_DeleteGarden_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenIDToDelete := "g1"
	dbErr := errors.New("delete failed")

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "gardens" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(gardenIDToDelete).WillReturnError(dbErr)
	mock.ExpectRollback()

	err = store.DeleteGarden(gardenIDToDelete)
	assert.ErrorIs(t, err, storage.ErrDatabase) // Assuming ParseDatabaseError maps it
}

func TestGormGardenStore_CreateGardenWithTransaction_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormGardenStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenToCreate := &models.Garden{ID: "g_tx_1", Name: "Transactional Garden", Location: "TX Loc", Description: "TX Desc"}
	bedsToCreate := []models.Bed{
		{ID: "b_tx_1", Name: "TX Bed 1", Type: "Flower"},
		{ID: "b_tx_2", Name: "TX Bed 2", Type: "Vegetable"},
	}

	// Expect transaction to begin
	mock.ExpectBegin()

	// Garden insert
	sqlGardenInsert := `INSERT INTO "gardens" ("id","name","location","description","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6)`
	mock.ExpectExec(regexp.QuoteMeta(sqlGardenInsert)).
		WithArgs(gardenToCreate.ID, gardenToCreate.Name, gardenToCreate.Location, gardenToCreate.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock Bed creations (loop)
	for _, bed := range bedsToCreate {
		// BedStorer.CreateBed (called with tx)
		// 1. Mock Garden lookup by BedStorer.CreateBed
		sqlGardenSelectForBed := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
		mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelectForBed)).
			WithArgs(gardenToCreate.ID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(gardenToCreate.ID))

		// 2. Mock Bed INSERT by BedStorer.CreateBed
		sqlBedInsert := `INSERT INTO "beds" ("id","garden_id","name","type","size","soil_type","notes","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
		mock.ExpectExec(regexp.QuoteMeta(sqlBedInsert)).
			WithArgs(bed.ID, gardenToCreate.ID, bed.Name, bed.Type, bed.Size, bed.SoilType, bed.Notes, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	// Expect transaction to commit
	mock.ExpectCommit()

	err = store.CreateGardenWithTransaction(gardenToCreate, bedsToCreate)
	assert.NoError(t, err)
}

// TODO: Add more tests for CreateGardenWithTransaction (e.g., garden creation fails, one of bed creations fail, deadlock simulation if possible)

// AnyTime struct and Match method can be copied here if needed for time.Time argument matching in other tests.
// type AnyTime struct{}
// func (a AnyTime) Match(v driver.Value) bool { _, ok := v.(time.Time); return ok }
