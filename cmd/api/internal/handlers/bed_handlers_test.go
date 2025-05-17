package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zjpiazza/plantastic/cmd/api/internal/handlers"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage" // For storage error types
	"github.com/zjpiazza/plantastic/internal/models"
)

// MockBedStore is a mock implementation of storage.BedStorer
type MockBedStore struct {
	mock.Mock
}

func (m *MockBedStore) GetAllBeds() ([]models.Bed, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Bed), args.Error(1)
}

func (m *MockBedStore) GetBedByID(bedID string) (models.Bed, error) {
	args := m.Called(bedID)
	if args.Get(0) == nil {
		return models.Bed{}, args.Error(1)
	}
	return args.Get(0).(models.Bed), args.Error(1)
}

func (m *MockBedStore) CreateBed(bed *models.Bed) error {
	args := m.Called(bed)
	return args.Error(0)
}

func (m *MockBedStore) UpdateBed(bed *models.Bed) error {
	args := m.Called(bed)
	return args.Error(0)
}

func (m *MockBedStore) DeleteBed(bedID string) error {
	args := m.Called(bedID)
	return args.Error(0)
}

// Add GetBedsByGardenID to satisfy the BedStorer interface
func (m *MockBedStore) GetBedsByGardenID(gardenID string) ([]models.Bed, error) {
	args := m.Called(gardenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Bed), args.Error(1)
}

func TestListBedsHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	now := time.Now().Truncate(time.Second)
	expectedBeds := []models.Bed{
		{ID: "b1", Name: "Test Bed 1", GardenID: "g1", CreatedAt: now, UpdatedAt: now},
		{ID: "b2", Name: "Test Bed 2", GardenID: "g1", CreatedAt: now, UpdatedAt: now},
	}
	mockStore.On("GetAllBeds").Return(expectedBeds, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handlers.ListBedsHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualBeds []models.Bed
	err := json.Unmarshal(w.Body.Bytes(), &actualBeds)
	require.NoError(t, err)
	for i := range actualBeds {
		actualBeds[i].CreatedAt = actualBeds[i].CreatedAt.Truncate(time.Second)
		actualBeds[i].UpdatedAt = actualBeds[i].UpdatedAt.Truncate(time.Second)
	}
	assert.ElementsMatch(t, expectedBeds, actualBeds)
	mockStore.AssertExpectations(t)
}

func TestListBedsHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	mockStore.On("GetAllBeds").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handlers.ListBedsHandler(mockStore, c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestGetBedHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	bedID := "b1"
	now := time.Now().Truncate(time.Second)
	expectedBed := models.Bed{ID: bedID, Name: "Test Bed", GardenID: "g1", CreatedAt: now, UpdatedAt: now}
	mockStore.On("GetBedByID", bedID).Return(expectedBed, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "bed_id", Value: bedID}}

	handlers.GetBedHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualBed models.Bed
	err := json.Unmarshal(w.Body.Bytes(), &actualBed)
	require.NoError(t, err)
	actualBed.CreatedAt = actualBed.CreatedAt.Truncate(time.Second)
	actualBed.UpdatedAt = actualBed.UpdatedAt.Truncate(time.Second)
	assert.Equal(t, expectedBed, actualBed)
	mockStore.AssertExpectations(t)
}

func TestGetBedHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	bedID := "b_nonexistent"
	mockStore.On("GetBedByID", bedID).Return(models.Bed{}, storage.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "bed_id", Value: bedID}}

	handlers.GetBedHandler(mockStore, c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStore.AssertExpectations(t)
}

func TestCreateBedHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	now := time.Now().Truncate(time.Second)
	newBed := models.Bed{Name: "New Bed", GardenID: "g1"}
	createdBed := newBed
	createdBed.ID = "b3"
	createdBed.CreatedAt = now
	createdBed.UpdatedAt = now

	mockStore.On("CreateBed", &newBed).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Bed)
		arg.ID = createdBed.ID
		arg.CreatedAt = createdBed.CreatedAt
		arg.UpdatedAt = createdBed.UpdatedAt
	})

	jsonBody, _ := json.Marshal(newBed)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/beds", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateBedHandler(mockStore, c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var actualResponse models.Bed
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	actualResponse.CreatedAt = actualResponse.CreatedAt.Truncate(time.Second)
	actualResponse.UpdatedAt = actualResponse.UpdatedAt.Truncate(time.Second)
	assert.Equal(t, createdBed, actualResponse)
	mockStore.AssertExpectations(t)
}

func TestCreateBedHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	invalidBed := models.Bed{Name: ""} // Missing GardenID too
	mockStore.On("CreateBed", &invalidBed).Return(storage.ErrValidation)

	jsonBody, _ := json.Marshal(invalidBed)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/beds", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateBedHandler(mockStore, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockStore.AssertExpectations(t)
}

func TestUpdateBedHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	bedID := "b1"
	updates := models.Bed{Name: "Updated Bed Name", GardenID: "g1"}

	mockStore.On("UpdateBed", mock.MatchedBy(func(b *models.Bed) bool {
		return b.ID == bedID && b.Name == updates.Name && b.GardenID == updates.GardenID
	})).Return(nil)

	jsonBody, _ := json.Marshal(updates)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "bed_id", Value: bedID}}
	c.Request, _ = http.NewRequest(http.MethodPut, "/beds/"+bedID, bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.UpdateBedHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestDeleteBedHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockBedStore)
	bedID := "b1"
	mockStore.On("DeleteBed", bedID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "bed_id", Value: bedID}}

	handlers.DeleteBedHandler(mockStore, c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}
