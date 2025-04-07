package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zjpiazza/plantastic/internal/models"
	"github.com/zjpiazza/plantastic/internal/storage"
	"gorm.io/gorm"
)

func ListTasksHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	tasks, err := storage.GetAllTasks(db)
	if err != nil {
		http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func CreateTaskHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := storage.CreateTask(db, &task); err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task created"})

}

func GetTaskHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["task_id"]

	task, err := storage.GetTaskByID(db, taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func UpdateTaskHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := storage.UpdateTask(db, &task)

	if err != nil {
		http.Error(w, "Unable to update task", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteTaskHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["task_id"]

	err := storage.DeleteTask(db, taskID)

	if err != nil {
		http.Error(w, "Unable to delete task", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
