package main

import (
	"encoding/csv"
	"os"
	"time"

	"github.com/zjpiazza/plantastic/pkg/models"
	"gorm.io/gorm"
)

func loadGardens(db *gorm.DB, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, record := range records {
		if i == 0 {
			continue // Skip header row
		}

		createdAt, _ := time.Parse(time.RFC3339, record[4])
		updatedAt, _ := time.Parse(time.RFC3339, record[5])

		garden := models.Garden{
			ID:          record[0],
			Name:        record[1],
			Location:    record[2],
			Description: record[3],
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
		db.Create(&garden)
	}
}

func loadBeds(db *gorm.DB, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, record := range records {
		if i == 0 {
			continue // Skip header row
		}

		createdAt, _ := time.Parse(time.RFC3339, record[7])
		updatedAt, _ := time.Parse(time.RFC3339, record[8])

		bed := models.Bed{
			ID:        record[0],
			GardenID:  record[1],
			Name:      record[2],
			Type:      record[3],
			Size:      record[4],
			SoilType:  record[5],
			Notes:     record[6],
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		db.Create(&bed)
	}
}

func loadTasks(db *gorm.DB, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, record := range records {
		if i == 0 {
			continue // Skip header row
		}

		dueDate, _ := time.Parse("2006-01-02", record[4])
		createdAt, _ := time.Parse(time.RFC3339, record[7])
		updatedAt, _ := time.Parse(time.RFC3339, record[8])

		task := models.Task{
			ID:          record[0],
			GardenID:    record[1],
			BedID:       &record[2], // Nullable field
			Description: record[3],
			DueDate:     dueDate,
			Status:      record[5],
			Priority:    record[6],
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}
		db.Create(&task)
	}
}
