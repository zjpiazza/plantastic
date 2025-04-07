package storage

import (
	"github.com/zjpiazza/plantastic/internal/models"
	"gorm.io/gorm"
)

func GetAllTasks(db *gorm.DB) ([]models.Task, error) {
	var tasks []models.Task
	result := db.Find(&tasks)
	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}

func CreateTask(db *gorm.DB, task *models.Task) error {
	result := db.Create(&task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetTaskByID(db *gorm.DB, taskID string) (models.Task, error) {
	var task models.Task
	result := db.Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		return models.Task{}, result.Error
	}
	return task, nil
}

func UpdateTask(db *gorm.DB, task *models.Task) error {
	result := db.Save(&task)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteTask(db *gorm.DB, taskID string) error {
	result := db.Delete(&models.Task{}, taskID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
