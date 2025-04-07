package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zjpiazza/plantastic/internal/models"
	"github.com/zjpiazza/plantastic/internal/storage"
	"gorm.io/gorm"
)

func ListGardensHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	gardens, err := storage.GetAllGardens(db)
	if err != nil {
		http.Error(w, "Failed to fetch gardens", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gardens)
}

func CreateGardenHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var garden models.Garden
	if err := json.NewDecoder(r.Body).Decode(&garden); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := storage.CreateGarden(db, &garden); err != nil {
		http.Error(w, "Failed to create garden", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Garden created"})

}

func GetGardenHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gardenID := vars["garden_id"]

	garden, err := storage.GetGardenByID(db, gardenID)
	if err != nil {
		http.Error(w, "Garden not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(garden)
}

func UpdateGardenHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var garden models.Garden

	err := storage.UpdateGarden(db, &garden)

	if err != nil {
		http.Error(w, "Unable to update garden", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteGardenHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gardenID := vars["garden_id"]

	err := storage.DeleteGarden(db, gardenID)

	if err != nil {
		http.Error(w, "Unable to delete garden", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
