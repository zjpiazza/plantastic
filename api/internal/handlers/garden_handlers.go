package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/api/internal/storage"
	"github.com/zjpiazza/plantastic/pkg/models"
	"gorm.io/gorm"
)

func ListGardensHandler(db *gorm.DB, c *gin.Context) {
	gardens, err := storage.GetAllGardens(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gardens"})
		return
	}

	c.JSON(http.StatusOK, gardens)
}

func CreateGardenHandler(db *gorm.DB, c *gin.Context) {
	var garden models.Garden
	if err := c.ShouldBindJSON(&garden); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storage.CreateGarden(db, &garden); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create garden"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Garden created"})
}

func GetGardenHandler(db *gorm.DB, c *gin.Context) {
	gardenID := c.Param("garden_id")

	garden, err := storage.GetGardenByID(db, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
		return
	}

	c.JSON(http.StatusOK, garden)
}

func UpdateGardenHandler(db *gorm.DB, c *gin.Context) {
	var garden models.Garden
	gardenID := c.Param("garden_id")

	// Get the garden updates from the request body
	if err := c.ShouldBindJSON(&garden); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get the garden from the database
	garden, err := storage.GetGardenByID(db, gardenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
		return
	}

	// Update the garden
	err = storage.UpdateGarden(db, &garden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update garden"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteGardenHandler(db *gorm.DB, c *gin.Context) {
	gardenID := c.Param("garden_id")

	err := storage.DeleteGarden(db, gardenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete garden"})
		return
	}

	c.Status(http.StatusNoContent)
}
