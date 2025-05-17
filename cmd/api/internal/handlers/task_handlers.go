package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
)

func ListTasksHandler(storer storage.TaskStorer, c *gin.Context) {
	tasks, err := storer.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func CreateTaskHandler(storer storage.TaskStorer, c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storer.CreateTask(&task); err != nil {
		if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		} else if err == storage.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}
	c.JSON(http.StatusCreated, task)
}

func GetTaskHandler(storer storage.TaskStorer, c *gin.Context) {
	taskID := c.Param("task_id")
	task, err := storer.GetTaskByID(taskID)
	if err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func UpdateTaskHandler(storer storage.TaskStorer, c *gin.Context) {
	taskID := c.Param("task_id")
	var taskUpdates models.Task
	if err := c.ShouldBindJSON(&taskUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	taskUpdates.ID = taskID // Ensure ID from path is used

	if err := storer.UpdateTask(&taskUpdates); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		} else if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func DeleteTaskHandler(storer storage.TaskStorer, c *gin.Context) {
	taskID := c.Param("task_id")
	if err := storer.DeleteTask(taskID); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete task"})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
