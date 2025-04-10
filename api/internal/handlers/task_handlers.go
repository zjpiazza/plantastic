package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/api/internal/storage"
	"github.com/zjpiazza/plantastic/pkg/models"
	"gorm.io/gorm"
)

func ListTasksHandler(db *gorm.DB, c *gin.Context) {
	tasks, err := storage.GetAllTasks(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func CreateTaskHandler(db *gorm.DB, c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storage.CreateTask(db, &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Task created"})
}

func GetTaskHandler(db *gorm.DB, c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := storage.GetTaskByID(db, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func UpdateTaskHandler(db *gorm.DB, c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := storage.UpdateTask(db, &task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update task"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteTaskHandler(db *gorm.DB, c *gin.Context) {
	taskID := c.Param("task_id")

	err := storage.DeleteTask(db, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete task"})
		return
	}

	c.Status(http.StatusNoContent)
}
