// cmd/api/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zjpiazza/plantastic/api/internal/routes"
	"github.com/zjpiazza/plantastic/pkg/models"
)

func main() {
	// Initialize database connection here
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=plantastic port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// AutoMigrate
	db.AutoMigrate(&models.Garden{}, &models.Bed{}, &models.Task{})
	fmt.Println("Database migration complete")

	// Load Gardens
	// loadGardens(db, "data/gardens.csv")

	// Load Beds
	// loadBeds(db, "data/beds.csv")

	// Load Tasks
	// loadTasks(db, "data/tasks.csv")

	// Initialize handlers, passing the db connection
	router := routes.SetupRoutes(db) // Pass the db to SetupRoutes

	router.Static("/static", "./static")

	fmt.Println("Starting server on :8080")
	log.Fatal(router.Run(":8080")) // Use router.Run for Gin
}
