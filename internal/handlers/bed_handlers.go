package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zjpiazza/plantastic/internal/models"
	"github.com/zjpiazza/plantastic/internal/storage"
	"gorm.io/gorm"
)

func ListBedsHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	beds, err := storage.GetAllBeds(db)
	if err != nil {
		http.Error(w, "Failed to fetch beds", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(beds)
}

func CreateBedHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var bed models.Bed
	if err := json.NewDecoder(r.Body).Decode(&bed); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := storage.CreateBed(db, &bed); err != nil {
		http.Error(w, "Failed to create bed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Bed created"})

}

func GetBedHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bedID := vars["bed_id"]

	bed, err := storage.GetBedByID(db, bedID)
	if err != nil {
		http.Error(w, "Bed not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bed)
}

func UpdateBedHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var bed models.Bed

	err := storage.UpdateBed(db, &bed)

	if err != nil {
		http.Error(w, "Unable to update bed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteBedHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bedID := vars["bed_id"]

	err := storage.DeleteBed(db, bedID)

	if err != nil {
		http.Error(w, "Unable to delete bed", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}
