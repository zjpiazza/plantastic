package handlers_test

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zjpiazza/plantastic/cmd/api/internal/handlers"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockGardenStore is a mock implementation of storage.GardenStorer
type MockGardenStore struct {
	mock.Mock
}

func (m *MockGardenStore) GetAllGardens() ([]models.Garden, error) {
	args := m.Called()
	// Need to type assert carefully, as Called() returns []interface{}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Garden), args.Error(1)
}

func (m *MockGardenStore) CreateGarden(garden *models.Garden) error {
	args := m.Called(garden)
	return args.Error(0)
}

func (m *MockGardenStore) GetGardenByID(gardenID string) (models.Garden, error) {
	args := m.Called(gardenID)
	if args.Get(0) == nil {
		// Return zero models.Garden and the error if the first arg is nil (indicating error path)
		return models.Garden{}, args.Error(1)
	}
	return args.Get(0).(models.Garden), args.Error(1)
}

func (m *MockGardenStore) UpdateGarden(garden *models.Garden) error {
	args := m.Called(garden)
	return args.Error(0)
}

func (m *MockGardenStore) DeleteGarden(gardenID string) error {
	args := m.Called(gardenID)
	return args.Error(0)
}

func (m *MockGardenStore) CreateGardenWithTransaction(garden *models.Garden, beds []models.Bed) error {
	args := m.Called(garden, beds)
	return args.Error(0)
}

func (m *MockGardenStore) GetAllGardensWithTimeout(timeout time.Duration) ([]models.Garden, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Garden), args.Error(1)
}

func (m *MockGardenStore) GetGardensByQuery(params map[string]string) ([]models.Garden, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Garden), args.Error(1)
}

// newMockDB creates a new GORM DB instance with sqlmock.
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, error) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB, // Use the mock sql.DB
		PreferSimpleProtocol: true,  // Recommended for sqlmock
	}), &gorm.Config{})
	if err != nil {
		sqlDB.Close() // Close the underlying mock sql.DB if gorm.Open fails
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func TestListGardensHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)

	now := time.Now().Truncate(time.Second)
	expectedGardens := []models.Garden{
		{ID: "g1", Name: "Test Garden 1", Location: "Loc1", Description: "Desc1", CreatedAt: now, UpdatedAt: now},
		{ID: "g2", Name: "Test Garden 2", Location: "Loc2", Description: "Desc2", CreatedAt: now, UpdatedAt: now},
	}

	mockStore.On("GetAllGardens").Return(expectedGardens, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.ListGardensHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualGardens []models.Garden
	err := json.Unmarshal(w.Body.Bytes(), &actualGardens)
	assert.NoError(t, err)
	for i := range actualGardens {
		actualGardens[i].CreatedAt = actualGardens[i].CreatedAt.Truncate(time.Second)
		actualGardens[i].UpdatedAt = actualGardens[i].UpdatedAt.Truncate(time.Second)
	}
	assert.ElementsMatch(t, expectedGardens, actualGardens)
	mockStore.AssertExpectations(t)
}

func TestListGardensHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)

	dbError := errors.New("simulated database error")
	mockStore.On("GetAllGardens").Return(nil, dbError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.ListGardensHandler(mockStore, c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	expectedResponse := gin.H{"error": "Failed to fetch gardens"}
	var actualResponse gin.H
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
	mockStore.AssertExpectations(t)
}

func TestListGardensHandler_EmptyResult(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)

	mockStore.On("GetAllGardens").Return([]models.Garden{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	handlers.ListGardensHandler(mockStore, c)

	t.Logf("Status Code: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	var actualGardens []models.Garden
	err := json.Unmarshal(w.Body.Bytes(), &actualGardens)
	assert.NoError(t, err)
	assert.Empty(t, actualGardens)
	mockStore.AssertExpectations(t)
}

// AnyTime is a helper for sqlmock to match any time.Time value.
// This is useful because comparing time.Time instances directly can be tricky due to nanosecond precision.
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface.
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestGetGardenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	gardenID := "g1"
	now := time.Now().Truncate(time.Second)
	expectedGarden := models.Garden{ID: gardenID, Name: "Test Garden", CreatedAt: now, UpdatedAt: now}

	mockStore.On("GetGardenByID", gardenID).Return(expectedGarden, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "garden_id", Value: gardenID}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/gardens/"+gardenID, nil)

	handlers.GetGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualGarden models.Garden
	err := json.Unmarshal(w.Body.Bytes(), &actualGarden)
	assert.NoError(t, err)
	actualGarden.CreatedAt = actualGarden.CreatedAt.Truncate(time.Second)
	actualGarden.UpdatedAt = actualGarden.UpdatedAt.Truncate(time.Second)
	assert.Equal(t, expectedGarden, actualGarden)
	mockStore.AssertExpectations(t)
}

func TestGetGardenHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	gardenID := "nonexistent"

	mockStore.On("GetGardenByID", gardenID).Return(models.Garden{}, storage.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "garden_id", Value: gardenID}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/gardens/"+gardenID, nil)

	handlers.GetGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStore.AssertExpectations(t)
}

func TestCreateGardenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	now := time.Now().Truncate(time.Second)
	newGarden := models.Garden{Name: "New Garden", Location: "New Loc", Description: "New Desc"}
	createdGarden := newGarden // Assume CreateGarden populates ID, CreatedAt, UpdatedAt
	createdGarden.ID = "g3"
	createdGarden.CreatedAt = now
	createdGarden.UpdatedAt = now

	// Mock CreateGarden to return nil (success)
	// The handler relies on the input 'garden' struct being populated by CreateGarden if it modifies it.
	// If CreateGarden doesn't modify the input struct but the handler expects it to, this test might need adjustment.
	mockStore.On("CreateGarden", &newGarden).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Garden)
		arg.ID = createdGarden.ID
		arg.CreatedAt = createdGarden.CreatedAt
		arg.UpdatedAt = createdGarden.UpdatedAt
	})

	jsonBody, _ := json.Marshal(newGarden)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/gardens", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var actualResponse models.Garden
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	assert.NoError(t, err)
	actualResponse.CreatedAt = actualResponse.CreatedAt.Truncate(time.Second)
	actualResponse.UpdatedAt = actualResponse.UpdatedAt.Truncate(time.Second)
	assert.Equal(t, createdGarden, actualResponse)
	mockStore.AssertExpectations(t)
}

func TestCreateGardenHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	invalidGarden := models.Garden{Name: ""} // Invalid: empty name

	mockStore.On("CreateGarden", &invalidGarden).Return(storage.ErrValidation)

	jsonBody, _ := json.Marshal(invalidGarden)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/gardens", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockStore.AssertExpectations(t)
}

func TestUpdateGardenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	gardenID := "g1"
	updates := models.Garden{Name: "Updated Name"} // ID will be set by handler from path param

	// We need to match on a pointer to a models.Garden struct where ID is gardenID
	mockStore.On("UpdateGarden", mock.MatchedBy(func(g *models.Garden) bool {
		return g.ID == gardenID && g.Name == updates.Name
	})).Return(nil)

	jsonBody, _ := json.Marshal(updates)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "garden_id", Value: gardenID}}
	c.Request, _ = http.NewRequest(http.MethodPut, "/gardens/"+gardenID, bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.UpdateGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code) // Changed from NoContent to OK to match handler change
	mockStore.AssertExpectations(t)
}

func TestDeleteGardenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockGardenStore)
	gardenID := "g1"

	mockStore.On("DeleteGarden", gardenID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "garden_id", Value: gardenID}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/gardens/"+gardenID, nil)

	handlers.DeleteGardenHandler(mockStore, c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}

// Note: Test for CreateGardenWithTransaction, GetAllGardensWithTimeout, GetGardensByQuery handlers are omitted for brevity
// but would follow similar patterns, mocking the respective GardenStorer interface methods.
