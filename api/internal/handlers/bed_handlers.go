package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/api/internal/storage"
	"github.com/zjpiazza/plantastic/pkg/models"
	"gorm.io/gorm"
)

func ListBedsHandler(db *gorm.DB, c *gin.Context) {
	beds, err := storage.GetAllBeds(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch beds"})
		return
	}

	c.JSON(http.StatusOK, beds)
}

func CreateBedHandler(db *gorm.DB, c *gin.Context) {
	var bed models.Bed
	if err := c.ShouldBindJSON(&bed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storage.CreateBed(db, &bed); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Bed created"})
}

func GetBedHandler(db *gorm.DB, c *gin.Context) {
	bedID := c.Param("bed_id")

	bed, err := storage.GetBedByID(db, bedID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bed not found"})
		return
	}

	c.JSON(http.StatusOK, bed)
}

func UpdateBedHandler(db *gorm.DB, c *gin.Context) {
	var bed models.Bed
	if err := c.ShouldBindJSON(&bed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := storage.UpdateBed(db, &bed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update bed"})
		return
	}

	c.Status(http.StatusNoContent)
}

func DeleteBedHandler(db *gorm.DB, c *gin.Context) {
	bedID := c.Param("bed_id")

	err := storage.DeleteBed(db, bedID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete bed"})
		return
	}

	c.Status(http.StatusNoContent)
}
