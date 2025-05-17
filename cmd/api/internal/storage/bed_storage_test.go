package storage_test

import (
	// "database/sql/driver" // Uncomment if AnyTime helper is used

	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

// newMockDBForStorageTest is already defined in garden_storage_test.go
// If these files are in the same package storage_test, it can be reused.
// For clarity if running tests separately or to avoid potential issues,
// it can be redefined or put in a shared test helper file.
// For now, assuming it can be seen or we'll manage if not.

func TestGormBedStore_GetAllBeds_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	expectedBeds := []models.Bed{
		{ID: "b1", GardenID: "g1", Name: "Rose Bed", Type: "Flower", Size: "2x4", SoilType: "Loamy", Notes: "Sunny spot", CreatedAt: now, UpdatedAt: now},
		{ID: "b2", GardenID: "g1", Name: "Vegetable Patch", Type: "Vegetable", Size: "4x8", SoilType: "Compost-rich", Notes: "Needs watering", CreatedAt: now, UpdatedAt: now},
	}

	rows := sqlmock.NewRows([]string{"id", "garden_id", "name", "type", "size", "soil_type", "notes", "created_at", "updated_at"}).
		AddRow(expectedBeds[0].ID, expectedBeds[0].GardenID, expectedBeds[0].Name, expectedBeds[0].Type, expectedBeds[0].Size, expectedBeds[0].SoilType, expectedBeds[0].Notes, expectedBeds[0].CreatedAt, expectedBeds[0].UpdatedAt).
		AddRow(expectedBeds[1].ID, expectedBeds[1].GardenID, expectedBeds[1].Name, expectedBeds[1].Type, expectedBeds[1].Size, expectedBeds[1].SoilType, expectedBeds[1].Notes, expectedBeds[1].CreatedAt, expectedBeds[1].UpdatedAt)

	sql := `SELECT * FROM "beds"`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WillReturnRows(rows)

	actualBeds, err := store.GetAllBeds()
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedBeds, actualBeds)
}

func TestGormBedStore_GetBedByID_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	bedID := "b1"
	expectedBed := models.Bed{ID: bedID, GardenID: "g1", Name: "Rose Bed", Type: "Flower", Size: "2x4", SoilType: "Loamy", Notes: "Sunny spot", CreatedAt: now, UpdatedAt: now}

	rows := sqlmock.NewRows([]string{"id", "garden_id", "name", "type", "size", "soil_type", "notes", "created_at", "updated_at"}).
		AddRow(expectedBed.ID, expectedBed.GardenID, expectedBed.Name, expectedBed.Type, expectedBed.Size, expectedBed.SoilType, expectedBed.Notes, expectedBed.CreatedAt, expectedBed.UpdatedAt)

	sql := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(bedID, 1).WillReturnRows(rows)

	actualBed, err := store.GetBedByID(bedID)
	assert.NoError(t, err)
	assert.Equal(t, expectedBed, actualBed)
}

func TestGormBedStore_CreateBed_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToCreate := &models.Bed{ID: "b_new", GardenID: "g1", Name: "New Herb Bed", Type: "Herb", Size: "1x3", SoilType: "Sandy", Notes: "Good drainage"}

	// 1. Mock the Garden lookup
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow("g1")
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(bedToCreate.GardenID, 1).WillReturnRows(gardenRows)

	// 2. Mock the Bed INSERT
	mock.ExpectBegin()
	sqlBedInsert := `INSERT INTO "beds" ("id","garden_id","name","type","size","soil_type","notes","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	mock.ExpectExec(regexp.QuoteMeta(sqlBedInsert)).
		WithArgs(bedToCreate.ID, bedToCreate.GardenID, bedToCreate.Name, bedToCreate.Type, bedToCreate.Size, bedToCreate.SoilType, bedToCreate.Notes, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := store.CreateBed(bedToCreate)
	assert.NoError(t, err)
}

func TestGormBedStore_CreateBed_GardenNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToCreate := &models.Bed{GardenID: "nonexistent_g1", Name: "Orphan Bed"}

	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(bedToCreate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.CreateBed(bedToCreate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormBedStore_GetBedsByGardenID_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	gardenID := "g1"
	expectedBeds := []models.Bed{
		{ID: "b1", GardenID: gardenID, Name: "Rose Bed", CreatedAt: now, UpdatedAt: now},
		{ID: "b2", GardenID: gardenID, Name: "Herb Patch", CreatedAt: now, UpdatedAt: now},
	}

	rows := sqlmock.NewRows([]string{"id", "garden_id", "name", "type", "size", "soil_type", "notes", "created_at", "updated_at"}).
		AddRow(expectedBeds[0].ID, expectedBeds[0].GardenID, expectedBeds[0].Name, expectedBeds[0].Type, expectedBeds[0].Size, expectedBeds[0].SoilType, expectedBeds[0].Notes, expectedBeds[0].CreatedAt, expectedBeds[0].UpdatedAt).
		AddRow(expectedBeds[1].ID, expectedBeds[1].GardenID, expectedBeds[1].Name, expectedBeds[1].Type, expectedBeds[1].Size, expectedBeds[1].SoilType, expectedBeds[1].Notes, expectedBeds[1].CreatedAt, expectedBeds[1].UpdatedAt)

	sql := `SELECT * FROM "beds" WHERE garden_id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID).WillReturnRows(rows)

	actualBeds, err := store.GetBedsByGardenID(gardenID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedBeds, actualBeds)
}

func TestGormBedStore_GetBedsByGardenID_NoneFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenID := "g2_empty"

	rows := sqlmock.NewRows([]string{"id", "garden_id", "name", "type", "size", "soil_type", "notes", "created_at", "updated_at"})
	sql := `SELECT * FROM "beds" WHERE garden_id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID).WillReturnRows(rows)

	actualBeds, err := store.GetBedsByGardenID(gardenID)
	assert.NoError(t, err)
	assert.Empty(t, actualBeds)
}

func TestGormBedStore_GetBedsByGardenID_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, err := db.DB()
	require.NoError(t, err)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	gardenID := "g_error"
	dbErr := errors.New("fetch beds by garden failed")

	sql := `SELECT * FROM "beds" WHERE garden_id = $1`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WithArgs(gardenID).WillReturnError(dbErr)

	_, err = store.GetBedsByGardenID(gardenID)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormBedStore_UpdateBed_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	bedToUpdate := &models.Bed{ID: "b1", GardenID: "g1", Name: "Updated Rose Bed", Type: "Flower Updated", Size: "3x5", SoilType: "Clay", Notes: "More sun", UpdatedAt: now}

	// 1. Mock the First() call to check if record exists
	existingRow := sqlmock.NewRows([]string{"id", "garden_id", "name"}).AddRow(bedToUpdate.ID, "g_original_for_select", "Old Rose Bed")
	sqlSelectOne := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectOne)).WithArgs(bedToUpdate.ID, 1).WillReturnRows(existingRow)

	// 2. Mock Garden validation lookup (for the new/updated GardenID)
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow(bedToUpdate.GardenID)
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(bedToUpdate.GardenID, 1).WillReturnRows(gardenRows)

	// 3. Mock the UPDATE statement
	mock.ExpectBegin()
	sqlUpdate := `UPDATE "beds" SET "garden_id"=$1,"name"=$2,"notes"=$3,"size"=$4,"soil_type"=$5,"type"=$6,"updated_at"=$7 WHERE id = $8`
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).
		WithArgs(bedToUpdate.GardenID, bedToUpdate.Name, bedToUpdate.Notes, bedToUpdate.Size, bedToUpdate.SoilType, bedToUpdate.Type, sqlmock.AnyArg(), bedToUpdate.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.UpdateBed(bedToUpdate)
	assert.NoError(t, err)
}

func TestGormBedStore_UpdateBed_NotFound(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // mock is not used
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	// defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }() // Mock will not be fully met if validation fails first

	// Scenario: GardenID is empty, causing early validation failure
	bedToUpdate := &models.Bed{ID: "b_nonexistent", Name: "Ghost Bed", GardenID: ""} // GardenID is empty

	// No DB calls expected due to early validation failure (bed.GardenID == "")

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation) // Expect ErrValidation due to empty GardenID
}

// New test specifically for when the Bed record is not found, but other validations pass
func TestGormBedStore_UpdateBed_ActualRecordNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToUpdate := &models.Bed{ID: "b_truly_nonexistent", Name: "Valid Name", GardenID: "g1"} // Valid name and GardenID

	// Mock Query 1 (Bed Lookup) to return gorm.ErrRecordNotFound
	sqlSelectBed := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectBed)).WithArgs(bedToUpdate.ID, 1).WillReturnError(gorm.ErrRecordNotFound)

	// No further mocks needed as it should return ErrRecordNotFound from bed lookup

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormBedStore_UpdateBed_ValidationError_NameEmpty(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // mock is not used if validation fails before DB
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()

	bedToUpdate := &models.Bed{ID: "b1", GardenID: "g1", Name: ""} // Invalid: Name is empty

	// No DB interaction expected, as validation (bed.Name == "") fails first.

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation)
	// No mock.ExpectationsWereMet() needed as no mocks are set
}

func TestGormBedStore_UpdateBed_GardenNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToUpdate := &models.Bed{ID: "b1", GardenID: "nonexistent_g", Name: "Bed With Bad Garden"}

	// 1. Mock Bed Select (succeeds)
	existingBedRow := sqlmock.NewRows([]string{"id", "garden_id"}).AddRow(bedToUpdate.ID, "g_original")
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(bedToUpdate.ID, 1).WillReturnRows(existingBedRow)

	// 2. Mock Garden Select for validation (fails)
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(bedToUpdate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormBedStore_UpdateBed_DBErrorOnSelect(t *testing.T) { // This test title implies error on *Bed* select
	db, _ := newMockDBForStorageTest(t) // mock is not used
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	// defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }() // Mock will not be fully met if validation fails first

	// Scenario: GardenID is empty, causing early validation failure
	bedToUpdate := &models.Bed{ID: "b1", Name: "Bed Update Fail Select", GardenID: ""} // GardenID is empty
	// dbErr := errors.New("db error on select bed") // This error won't be reached, and dbErr is unused

	// No DB calls expected due to early validation failure (bed.GardenID == "")

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation) // Expect ErrValidation due to empty GardenID
}

// New test specifically for DB error on Bed record select, other validations pass
func TestGormBedStore_UpdateBed_DBErrorOnActualBedSelect(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToUpdate := &models.Bed{ID: "b1", Name: "Valid Name", GardenID: "g1"} // Valid name and GardenID
	dbSelectErr := errors.New("db error on bed select")

	// Mock Query 1 (Bed Lookup) to return a generic DB error
	sqlSelectBed := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectBed)).WithArgs(bedToUpdate.ID, 1).WillReturnError(dbSelectErr)

	// No further mocks needed

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrDatabase) // Expect custom ErrDatabase after parsing
}

func TestGormBedStore_UpdateBed_DBErrorOnUpdate(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedToUpdate := &models.Bed{ID: "b1", GardenID: "g1", Name: "Bed Update Fail Update", Type: "T", Size: "S", SoilType: "ST", Notes: "N"}
	dbUpdateErr := errors.New("db error on update bed")

	// 1. Mock Bed Select (succeeds)
	existingBedRow := sqlmock.NewRows([]string{"id", "garden_id"}).AddRow(bedToUpdate.ID, "g_original")
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 ORDER BY "beds"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(bedToUpdate.ID, 1).WillReturnRows(existingBedRow)

	// 2. Mock Garden validation lookup (succeeds)
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow(bedToUpdate.GardenID)
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(bedToUpdate.GardenID, 1).WillReturnRows(gardenRows)

	// 3. Mock Update (fails)
	mock.ExpectBegin()
	sqlUpdate := `UPDATE "beds" SET "garden_id"=$1,"name"=$2,"notes"=$3,"size"=$4,"soil_type"=$5,"type"=$6,"updated_at"=$7 WHERE id = $8`
	mock.ExpectExec(regexp.QuoteMeta(sqlUpdate)).
		WithArgs(bedToUpdate.GardenID, bedToUpdate.Name, bedToUpdate.Notes, bedToUpdate.Size, bedToUpdate.SoilType, bedToUpdate.Type, sqlmock.AnyArg(), bedToUpdate.ID).
		WillReturnError(dbUpdateErr)
	mock.ExpectRollback()

	err := store.UpdateBed(bedToUpdate)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormBedStore_DeleteBed_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedIDToDelete := "b1_delete"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "beds" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(bedIDToDelete).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.DeleteBed(bedIDToDelete)
	assert.NoError(t, err)
}

func TestGormBedStore_DeleteBed_NotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedIDToDelete := "b_nonexistent_delete"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "beds" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(bedIDToDelete).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := store.DeleteBed(bedIDToDelete)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormBedStore_DeleteBed_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormBedStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedIDToDelete := "b1_err_delete"
	dbErr := errors.New("db error on delete bed")

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "beds" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(bedIDToDelete).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := store.DeleteBed(bedIDToDelete)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

// TODO: Add tests for UpdateBed (Success, NotFound, ValidationError, DBError on selects/update)
// TODO: Add tests for DeleteBed (Success, NotFound, DBError)
