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

// MockTaskStore is a mock implementation of storage.TaskStorer
type MockTaskStore struct {
	mock.Mock
}

func (m *MockTaskStore) GetAllTasks() ([]models.Task, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskStore) GetTaskByID(taskID string) (models.Task, error) {
	args := m.Called(taskID)
	var task models.Task
	if args.Get(0) != nil {
		task = args.Get(0).(models.Task)
	}
	return task, args.Error(1)
}

func (m *MockTaskStore) CreateTask(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskStore) UpdateTask(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskStore) DeleteTask(taskID string) error {
	args := m.Called(taskID)
	return args.Error(0)
}

func TestListTasksHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	now := time.Now().Truncate(time.Second)
	expectedTasks := []models.Task{
		{ID: "t1", Description: "Test Task 1", GardenID: "g1", DueDate: now, CreatedAt: now, UpdatedAt: now},
		{ID: "t2", Description: "Test Task 2", GardenID: "g1", DueDate: now, CreatedAt: now, UpdatedAt: now},
	}
	mockStore.On("GetAllTasks").Return(expectedTasks, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handlers.ListTasksHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualTasks []models.Task
	err := json.Unmarshal(w.Body.Bytes(), &actualTasks)
	require.NoError(t, err)
	for i := range actualTasks {
		actualTasks[i].CreatedAt = actualTasks[i].CreatedAt.Truncate(time.Second)
		actualTasks[i].UpdatedAt = actualTasks[i].UpdatedAt.Truncate(time.Second)
		actualTasks[i].DueDate = actualTasks[i].DueDate.Truncate(time.Second)
	}
	assert.ElementsMatch(t, expectedTasks, actualTasks)
	mockStore.AssertExpectations(t)
}

func TestListTasksHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	mockStore.On("GetAllTasks").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handlers.ListTasksHandler(mockStore, c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockStore.AssertExpectations(t)
}

func TestGetTaskHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	taskID := "t1"
	now := time.Now().Truncate(time.Second)
	expectedTask := models.Task{ID: taskID, Description: "Test Task", GardenID: "g1", DueDate: now, CreatedAt: now, UpdatedAt: now}
	mockStore.On("GetTaskByID", taskID).Return(expectedTask, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "task_id", Value: taskID}}

	handlers.GetTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	var actualTask models.Task
	err := json.Unmarshal(w.Body.Bytes(), &actualTask)
	require.NoError(t, err)
	actualTask.CreatedAt = actualTask.CreatedAt.Truncate(time.Second)
	actualTask.UpdatedAt = actualTask.UpdatedAt.Truncate(time.Second)
	actualTask.DueDate = actualTask.DueDate.Truncate(time.Second)
	assert.Equal(t, expectedTask, actualTask)
	mockStore.AssertExpectations(t)
}

func TestGetTaskHandler_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	taskID := "t_nonexistent"
	mockStore.On("GetTaskByID", taskID).Return(models.Task{}, storage.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "task_id", Value: taskID}}

	handlers.GetTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockStore.AssertExpectations(t)
}

func TestCreateTaskHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	now := time.Now().Truncate(time.Second)
	newTask := models.Task{Description: "New Task", GardenID: "g1", DueDate: now}
	createdTask := newTask
	createdTask.ID = "t3"
	createdTask.CreatedAt = now
	createdTask.UpdatedAt = now

	mockStore.On("CreateTask", &newTask).Return(nil).Run(func(args mock.Arguments) {
		arg := args.Get(0).(*models.Task)
		arg.ID = createdTask.ID
		arg.CreatedAt = createdTask.CreatedAt
		arg.UpdatedAt = createdTask.UpdatedAt
	})

	jsonBody, _ := json.Marshal(newTask)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusCreated, w.Code)
	var actualResponse models.Task
	err := json.Unmarshal(w.Body.Bytes(), &actualResponse)
	require.NoError(t, err)
	actualResponse.CreatedAt = actualResponse.CreatedAt.Truncate(time.Second)
	actualResponse.UpdatedAt = actualResponse.UpdatedAt.Truncate(time.Second)
	actualResponse.DueDate = actualResponse.DueDate.Truncate(time.Second)
	assert.Equal(t, createdTask, actualResponse)
	mockStore.AssertExpectations(t)
}

func TestCreateTaskHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	invalidTask := models.Task{Description: ""} // Missing GardenID too
	mockStore.On("CreateTask", &invalidTask).Return(storage.ErrValidation)

	jsonBody, _ := json.Marshal(invalidTask)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.CreateTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockStore.AssertExpectations(t)
}

func TestUpdateTaskHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	taskID := "t1"
	updates := models.Task{Description: "Updated Task Desc", GardenID: "g1"}

	mockStore.On("UpdateTask", mock.MatchedBy(func(tsk *models.Task) bool {
		return tsk.ID == taskID && tsk.Description == updates.Description && tsk.GardenID == updates.GardenID
	})).Return(nil)

	jsonBody, _ := json.Marshal(updates)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "task_id", Value: taskID}}
	c.Request, _ = http.NewRequest(http.MethodPut, "/tasks/"+taskID, bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handlers.UpdateTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockStore.AssertExpectations(t)
}

func TestDeleteTaskHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockStore := new(MockTaskStore)
	taskID := "t1"
	mockStore.On("DeleteTask", taskID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "task_id", Value: taskID}}

	handlers.DeleteTaskHandler(mockStore, c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockStore.AssertExpectations(t)
}
