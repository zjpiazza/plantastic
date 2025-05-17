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
	"github.com/zjpiazza/plantastic/internal/models" // Assuming postgres, adjust if different
	"gorm.io/gorm"
)

// newMockDBForStorageTest is defined in another _test.go file in this package (e.g., garden_storage_test.go)

func TestGormTaskStore_GetAllTasks_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	bedID1 := "b1"
	expectedTasks := []models.Task{
		{ID: "t1", GardenID: "g1", BedID: &bedID1, Description: "Water roses", DueDate: now.Add(24 * time.Hour), Status: "Pending", Priority: "High", CreatedAt: now, UpdatedAt: now},
		{ID: "t2", GardenID: "g1", BedID: nil, Description: "Plan vegetable layout", DueDate: now.Add(72 * time.Hour), Status: "Todo", Priority: "Medium", CreatedAt: now, UpdatedAt: now},
	}

	rows := sqlmock.NewRows([]string{"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at"}).
		AddRow(expectedTasks[0].ID, expectedTasks[0].GardenID, expectedTasks[0].BedID, expectedTasks[0].Description, expectedTasks[0].DueDate, expectedTasks[0].Status, expectedTasks[0].Priority, expectedTasks[0].CreatedAt, expectedTasks[0].UpdatedAt).
		AddRow(expectedTasks[1].ID, expectedTasks[1].GardenID, expectedTasks[1].BedID, expectedTasks[1].Description, expectedTasks[1].DueDate, expectedTasks[1].Status, expectedTasks[1].Priority, expectedTasks[1].CreatedAt, expectedTasks[1].UpdatedAt)

	sql := `SELECT * FROM "tasks"`
	mock.ExpectQuery(regexp.QuoteMeta(sql)).WillReturnRows(rows)

	actualTasks, err := store.GetAllTasks()
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedTasks, actualTasks)
}

func TestGormTaskStore_GetTaskByID_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	taskID := "t1"
	bedID := "b1"
	expectedTask := models.Task{ID: taskID, GardenID: "g1", BedID: &bedID, Description: "Water roses", DueDate: now.Add(24 * time.Hour), Status: "Pending", Priority: "High", CreatedAt: now, UpdatedAt: now}

	rows := sqlmock.NewRows([]string{"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at"}).
		AddRow(expectedTask.ID, expectedTask.GardenID, expectedTask.BedID, expectedTask.Description, expectedTask.DueDate, expectedTask.Status, expectedTask.Priority, expectedTask.CreatedAt, expectedTask.UpdatedAt)

	sqlSelect := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelect)).WithArgs(taskID, 1).WillReturnRows(rows)

	actualTask, err := store.GetTaskByID(taskID)
	assert.NoError(t, err)
	// Using assert.True and comparing fields individually to avoid time.Location issues in assert.Equal
	assert.True(t, expectedTask.ID == actualTask.ID &&
		expectedTask.GardenID == actualTask.GardenID &&
		((expectedTask.BedID == nil && actualTask.BedID == nil) || (expectedTask.BedID != nil && actualTask.BedID != nil && *expectedTask.BedID == *actualTask.BedID)) &&
		expectedTask.Description == actualTask.Description &&
		expectedTask.DueDate.Unix() == actualTask.DueDate.Unix(), // Compare Unix timestamp for time
		"Expected task %v, got %v", expectedTask, actualTask)

}

func TestGormTaskStore_CreateTask_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedID := "b1"
	taskToCreate := &models.Task{ID: "task_new_1", GardenID: "g1", BedID: &bedID, Description: "Weed main bed", DueDate: time.Now().Add(48 * time.Hour), Status: "Todo", Priority: "Medium"}

	// 1. Mock Garden lookup
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow("g1")
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToCreate.GardenID, 1).WillReturnRows(gardenRows)

	// 2. Mock Bed lookup (since BedID is provided)
	bedRows := sqlmock.NewRows([]string{"id", "garden_id"}).AddRow(*taskToCreate.BedID, taskToCreate.GardenID)
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 AND garden_id = $2 ORDER BY "beds"."id" LIMIT $3`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(*taskToCreate.BedID, taskToCreate.GardenID, 1).WillReturnRows(bedRows)

	// 3. Mock Task INSERT
	mock.ExpectBegin()
	sqlTaskInsert := `INSERT INTO "tasks" ("id","garden_id","bed_id","description","due_date","status","priority","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskInsert)).
		WithArgs(taskToCreate.ID, taskToCreate.GardenID, taskToCreate.BedID, taskToCreate.Description, taskToCreate.DueDate, taskToCreate.Status, taskToCreate.Priority, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := store.CreateTask(taskToCreate)
	assert.NoError(t, err)
}

func TestGormTaskStore_CreateTask_Success_NoBedID(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskToCreate := &models.Task{ID: "task_new_2", GardenID: "g1", BedID: nil, Description: "General garden check", DueDate: time.Now().Add(48 * time.Hour), Status: "Todo", Priority: "Low"}

	// 1. Mock Garden lookup
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow("g1")
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToCreate.GardenID, 1).WillReturnRows(gardenRows)

	// No Bed lookup expected as BedID is nil

	// 2. Mock Task INSERT
	mock.ExpectBegin()
	sqlTaskInsert := `INSERT INTO "tasks" ("id","garden_id","bed_id","description","due_date","status","priority","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskInsert)).
		WithArgs(taskToCreate.ID, taskToCreate.GardenID, taskToCreate.BedID, taskToCreate.Description, taskToCreate.DueDate, taskToCreate.Status, taskToCreate.Priority, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := store.CreateTask(taskToCreate)
	assert.NoError(t, err)
}

func TestGormTaskStore_CreateTask_GardenNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskToCreate := &models.Task{GardenID: "nonexistent_g1", Description: "Task for non-existent garden"}

	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToCreate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.CreateTask(taskToCreate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormTaskStore_CreateTask_BedNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	bedID := "nonexistent_b1"
	taskToCreate := &models.Task{GardenID: "g1", BedID: &bedID, Description: "Task for non-existent bed"}

	// 1. Mock Garden lookup (success)
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow("g1")
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToCreate.GardenID, 1).WillReturnRows(gardenRows)

	// 2. Mock Bed lookup (fail)
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 AND garden_id = $2 ORDER BY "beds"."id" LIMIT $3`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(*taskToCreate.BedID, taskToCreate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.CreateTask(taskToCreate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormTaskStore_CreateTask_MissingFieldsValidationError(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // Mock not strictly needed if validation hits first
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB() // db.DB() might be nil if newMockDBForStorageTest returns a nil gorm.DB on error
	if rawSqlDB != nil {
		defer rawSqlDB.Close()
	}
	require.NoError(t, errDB)

	// Test case 1: Missing Description
	taskMissingDesc := &models.Task{GardenID: "g1"}
	err := store.CreateTask(taskMissingDesc)
	assert.ErrorIs(t, err, storage.ErrValidation, "Expected ErrValidation for missing description")

	// Test case 2: Missing GardenID
	taskMissingGardenID := &models.Task{Description: "Valid Description"}
	err = store.CreateTask(taskMissingGardenID)
	assert.ErrorIs(t, err, storage.ErrValidation, "Expected ErrValidation for missing GardenID")
}

func TestGormTaskStore_CreateTask_DBErrorOnInsert(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskToCreate := &models.Task{ID: "task_err_insert", GardenID: "g1", Description: "Test Desc", DueDate: time.Now()}
	dbErr := errors.New("DB insert error")

	// 1. Mock Garden lookup (success)
	gardenRows := sqlmock.NewRows([]string{"id"}).AddRow("g1")
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToCreate.GardenID, 1).WillReturnRows(gardenRows)

	// No Bed lookup if BedID is nil

	// 2. Mock Task INSERT (fail)
	mock.ExpectBegin()
	sqlTaskInsert := `INSERT INTO "tasks" ("id","garden_id","bed_id","description","due_date","status","priority","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskInsert)).
		WithArgs(taskToCreate.ID, taskToCreate.GardenID, taskToCreate.BedID, taskToCreate.Description, taskToCreate.DueDate, taskToCreate.Status, taskToCreate.Priority, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(dbErr)
	mock.ExpectRollback()

	err := store.CreateTask(taskToCreate)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormTaskStore_UpdateTask_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	now := time.Now()
	taskID := "t1"
	originalBedID := "b1"
	updatedBedID := "b2"
	taskToUpdate := &models.Task{
		ID:          taskID,
		GardenID:    "g1",
		BedID:       &updatedBedID,
		Description: "Updated Task Description",
		DueDate:     now.Add(48 * time.Hour),
		Status:      "In Progress",
		Priority:    "High",
	}

	// 1. Mock Task Lookup (s.db.First(&existingTask...))
	existingTaskRows := sqlmock.NewRows([]string{
		"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at",
	}).AddRow(
		taskID, taskToUpdate.GardenID, &originalBedID, "Original Description",
		now.Add(-24*time.Hour), "Todo", "Medium", now.Add(-48*time.Hour), now.Add(-48*time.Hour),
	)
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskID, 1).WillReturnRows(existingTaskRows)

	// 2a. Mock Bed Lookup (validateTaskDataAndRefs - s.db.First(&bed...))
	bedRows := sqlmock.NewRows([]string{"id", "garden_id"}).AddRow(*taskToUpdate.BedID, taskToUpdate.GardenID)
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 AND garden_id = $2 ORDER BY "beds"."id" LIMIT $3`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(*taskToUpdate.BedID, taskToUpdate.GardenID, 1).WillReturnRows(bedRows)

	// 3. Mock Task UPDATE
	mock.ExpectBegin()
	sqlTaskUpdate := `UPDATE "tasks" SET "bed_id"=$1,"description"=$2,"due_date"=$3,"garden_id"=$4,"priority"=$5,"status"=$6,"updated_at"=$7 WHERE id = $8`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskUpdate)).
		WithArgs(taskToUpdate.BedID, taskToUpdate.Description, taskToUpdate.DueDate, taskToUpdate.GardenID, taskToUpdate.Priority, taskToUpdate.Status, sqlmock.AnyArg(), taskToUpdate.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.UpdateTask(taskToUpdate)
	assert.NoError(t, err)
}

func TestGormTaskStore_UpdateTask_Success_BedIDToNil(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskID := "t1"
	originalBedID := "b1"
	taskToUpdate := &models.Task{
		ID:          taskID,
		GardenID:    "g1",
		BedID:       nil, // BedID is set to nil
		Description: "Description updated, bed removed",
		// Other fields (DueDate, Status, Priority) will be their zero values if not set
		// GORM will update them if they are in the updateFields map in the actual code
	}

	// 1. Mock Task Lookup
	existingTaskRows := sqlmock.NewRows([]string{
		"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at",
	}).AddRow(
		taskID, taskToUpdate.GardenID, &originalBedID, "Original Description",
		time.Now().Add(-24*time.Hour), "Todo", "Medium",
		time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour),
	)
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskID, 1).WillReturnRows(existingTaskRows)

	// 3. Mock Task UPDATE
	mock.ExpectBegin()
	// updateFields in GormTaskStore.UpdateTask for nil BedID will include "bed_id": nil
	// Alphabetical order of likely fields being updated (assuming others are zero/empty and included):
	// bed_id, description, due_date (zero), garden_id, priority (empty), status (empty), updated_at
	sqlTaskUpdate := `UPDATE "tasks" SET "bed_id"=$1,"description"=$2,"due_date"=$3,"garden_id"=$4,"priority"=$5,"status"=$6,"updated_at"=$7 WHERE id = $8`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskUpdate)).
		WithArgs(nil, taskToUpdate.Description, taskToUpdate.DueDate, taskToUpdate.GardenID, taskToUpdate.Priority, taskToUpdate.Status, sqlmock.AnyArg(), taskToUpdate.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.UpdateTask(taskToUpdate)
	assert.NoError(t, err)
}

func TestGormTaskStore_UpdateTask_NotFound(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // mock is not used
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	// No mock.ExpectationsWereMet() as no mocks are set

	// GardenID is empty by default, Description is not. ID is not empty.
	// validateTaskDataAndRefs will fail on empty GardenID before DB lookup.
	taskToUpdate := &models.Task{ID: "nonexistent_task", Description: "Ghost Task"}

	// No DB mocks needed, as initial validation (empty GardenID) should fail.

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation) // Expect ErrValidation due to empty GardenID
}

// New test for when the Task record itself is not found, but basic validations pass.
func TestGormTaskStore_UpdateTask_ActualRecordNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskToUpdate := &models.Task{ID: "truly_nonexistent_task", GardenID: "g1", Description: "Valid Desc"}

	// 1. Mock Task Lookup to return gorm.ErrRecordNotFound
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskToUpdate.ID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormTaskStore_UpdateTask_ValidationError_EmptyDescription(t *testing.T) {
	db, _ := newMockDBForStorageTest(t) // mock is not used
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	// No mock.ExpectationsWereMet() as no mocks are set

	taskID := "t1"
	// Description is empty, GardenID is not. ID is not empty.
	// validateTaskDataAndRefs will fail on empty Description before DB lookup.
	taskToUpdate := &models.Task{ID: taskID, GardenID: "g1", Description: ""}

	// No DB mocks needed, as initial validation (empty Description) should fail.

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormTaskStore_UpdateTask_ValidationError_GardenNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskID := "t1"
	originalGardenID := "original_g1"
	taskToUpdate := &models.Task{ID: taskID, GardenID: "nonexistent_g_update", Description: "Valid Desc"}

	// 1. Mock GetTaskByID (succeeds)
	existingTaskRows := sqlmock.NewRows([]string{
		"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at",
	}).AddRow(
		taskID, originalGardenID, nil, "Original Description",
		time.Now().Add(-24*time.Hour), "Todo", "Medium",
		time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour),
	)
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskID, 1).WillReturnRows(existingTaskRows)

	// 2. Mock Garden lookup (fails) for the new GardenID
	sqlGardenSelect := `SELECT * FROM "gardens" WHERE id = $1 ORDER BY "gardens"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlGardenSelect)).WithArgs(taskToUpdate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormTaskStore_UpdateTask_ValidationError_BedNotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskID := "t1"
	updatedBedID := "nonexistent_b_update"
	taskToUpdate := &models.Task{ID: taskID, GardenID: "g1", BedID: &updatedBedID, Description: "Valid Desc"}

	// 1. Mock GetTaskByID (succeeds)
	existingTaskRows := sqlmock.NewRows([]string{
		"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at",
	}).AddRow(
		taskID, "g1", nil, "Original Description",
		time.Now().Add(-24*time.Hour), "Todo", "Medium",
		time.Now().Add(-48*time.Hour), time.Now().Add(-48*time.Hour),
	)
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskID, 1).WillReturnRows(existingTaskRows)

	// 2. Mock Bed lookup (fails) for BedID "nonexistent_b_update" in Garden "g1"
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 AND garden_id = $2 ORDER BY "beds"."id" LIMIT $3`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(*taskToUpdate.BedID, taskToUpdate.GardenID, 1).WillReturnError(gorm.ErrRecordNotFound)

	// No garden lookup should happen as validation fails on bed lookup

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrValidation)
}

func TestGormTaskStore_UpdateTask_DBErrorOnSelect(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskToUpdate := &models.Task{ID: "t1", GardenID: "g1", Description: "Update Fail Select"}
	dbErr := errors.New("select task for update failed")

	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskToUpdate.ID, 1).WillReturnError(dbErr)

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormTaskStore_UpdateTask_DBErrorOnUpdate(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskID := "t1"
	// This is the 'taskToUpdate' struct, containing the new values for the update.
	updatedBedID := "b1" // Renamed from 'bedID' in original test to avoid confusion
	taskToUpdate := &models.Task{
		ID:          taskID,
		GardenID:    "g1",
		BedID:       &updatedBedID,
		Description: "Update Fail SQL Update",
		DueDate:     time.Now(), // Ensure all fields that might be updated are set
		Status:      "Pending",
		Priority:    "Low",
	}
	dbUpdateErr := errors.New("db error on actual update")

	// 1. Mock Task Lookup (succeeds)
	// This represents the state of the task in the DB *before* the update.
	// It must provide all columns expected by 'SELECT *' and the models.Task struct.
	originalTaskBedID := "b_original_for_db_error_update_test" // An original bed ID
	originalTaskDescription := "Original Description before DB error update test"
	originalTaskDueDate := time.Now().Add(-24 * time.Hour)
	originalTaskStatus := "OriginalStatus"
	originalTaskPriority := "OriginalPriority"
	originalTaskCreatedAt := time.Now().Add(-48 * time.Hour)
	originalTaskUpdatedAt := time.Now().Add(-48 * time.Hour)

	existingTaskRows := sqlmock.NewRows([]string{
		"id", "garden_id", "bed_id", "description", "due_date", "status", "priority", "created_at", "updated_at",
	}).AddRow(
		taskID, taskToUpdate.GardenID, &originalTaskBedID, originalTaskDescription,
		originalTaskDueDate, originalTaskStatus, originalTaskPriority, originalTaskCreatedAt, originalTaskUpdatedAt,
	)
	// NOTE: GORM for PostgreSQL doesn't add table name to the column name in the WHERE clause
	sqlSelectTask := `SELECT * FROM "tasks" WHERE id = $1 ORDER BY "tasks"."id" LIMIT $2`
	mock.ExpectQuery(regexp.QuoteMeta(sqlSelectTask)).WithArgs(taskID, 1).WillReturnRows(existingTaskRows)

	// 2a. Mock Bed Lookup (succeeds, since BedID is provided in taskToUpdate) - for validating *taskToUpdate.BedID
	// Note: This is moved before garden lookup based on the implementation sequence
	bedRows := sqlmock.NewRows([]string{"id", "garden_id"}).AddRow(*taskToUpdate.BedID, taskToUpdate.GardenID)
	sqlBedSelect := `SELECT * FROM "beds" WHERE id = $1 AND garden_id = $2 ORDER BY "beds"."id" LIMIT $3`
	mock.ExpectQuery(regexp.QuoteMeta(sqlBedSelect)).WithArgs(*taskToUpdate.BedID, taskToUpdate.GardenID, 1).WillReturnRows(bedRows)

	// 3. Mock Task UPDATE (fails)
	mock.ExpectBegin()
	sqlTaskUpdate := `UPDATE "tasks" SET "bed_id"=$1,"description"=$2,"due_date"=$3,"garden_id"=$4,"priority"=$5,"status"=$6,"updated_at"=$7 WHERE id = $8`
	mock.ExpectExec(regexp.QuoteMeta(sqlTaskUpdate)).
		WithArgs(taskToUpdate.BedID, taskToUpdate.Description, taskToUpdate.DueDate, taskToUpdate.GardenID, taskToUpdate.Priority, taskToUpdate.Status, sqlmock.AnyArg(), taskToUpdate.ID).
		WillReturnError(dbUpdateErr)
	mock.ExpectRollback()

	err := store.UpdateTask(taskToUpdate)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

func TestGormTaskStore_DeleteTask_Success(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskIDToDelete := "t1_delete"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "tasks" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(taskIDToDelete).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.DeleteTask(taskIDToDelete)
	assert.NoError(t, err)
}

func TestGormTaskStore_DeleteTask_NotFound(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskIDToDelete := "nonexistent_task_delete"

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "tasks" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(taskIDToDelete).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := store.DeleteTask(taskIDToDelete)
	assert.ErrorIs(t, err, storage.ErrRecordNotFound)
}

func TestGormTaskStore_DeleteTask_DBError(t *testing.T) {
	db, mock := newMockDBForStorageTest(t)
	store := storage.NewGormTaskStore(db)
	rawSqlDB, errDB := db.DB()
	require.NoError(t, errDB)
	defer rawSqlDB.Close()
	defer func() { assert.NoError(t, mock.ExpectationsWereMet()) }()

	taskIDToDelete := "t1_err_delete"
	dbErr := errors.New("DB delete error")

	mock.ExpectBegin()
	sqlDelete := `DELETE FROM "tasks" WHERE id = $1`
	mock.ExpectExec(regexp.QuoteMeta(sqlDelete)).WithArgs(taskIDToDelete).WillReturnError(dbErr)
	mock.ExpectRollback()

	err := store.DeleteTask(taskIDToDelete)
	assert.ErrorIs(t, err, storage.ErrDatabase)
}

// Test for GetTasksByBedID - if you add this method to TaskStorer
/*
func TestGormTaskStore_GetTasksByBedID_Success(t *testing.T) {
    // ... setup ...
    bedID := "b1"
    // ... mock expectations for SELECT * FROM "tasks" WHERE bed_id = $1 ...
    // ... assertions ...
}
*/

// Test for GetTasksByGardenID - if you add this method to TaskStorer
/*
func TestGormTaskStore_GetTasksByGardenID_Success(t *testing.T) {
    // ... setup ...
    gardenID := "g1"
    // ... mock expectations for SELECT * FROM "tasks" WHERE garden_id = $1 ...
    // ... assertions ...
}
*/
