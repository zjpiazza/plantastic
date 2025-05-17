package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
)

// ListGardensHandler uses GardenStorer to fetch and return gardens.
func ListGardensHandler(storer storage.GardenStorer, c *gin.Context) {
	gardens, err := storer.GetAllGardens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gardens"})
		return
	}
	c.JSON(http.StatusOK, gardens)
}

// CreateGardenHandler uses GardenStorer to create a new garden.
func CreateGardenHandler(storer storage.GardenStorer, c *gin.Context) {
	var garden models.Garden
	if err := c.ShouldBindJSON(&garden); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storer.CreateGarden(&garden); err != nil {
		// Check for specific storage errors to return more appropriate HTTP status codes
		if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		} else if err == storage.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create garden"})
		return
	}
	c.JSON(http.StatusCreated, garden) // Return the created garden
}

// GetGardenHandler uses GardenStorer to fetch a specific garden by ID.
func GetGardenHandler(storer storage.GardenStorer, c *gin.Context) {
	gardenID := c.Param("garden_id")
	garden, err := storer.GetGardenByID(gardenID)
	if err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}
	c.JSON(http.StatusOK, garden)
}

// UpdateGardenHandler uses GardenStorer to update an existing garden.
func UpdateGardenHandler(storer storage.GardenStorer, c *gin.Context) {
	gardenID := c.Param("garden_id")
	var gardenUpdates models.Garden

	if err := c.ShouldBindJSON(&gardenUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// It's important that the ID from the path is used, not from the body, to prevent misuse.
	gardenUpdates.ID = gardenID

	if err := storer.UpdateGarden(&gardenUpdates); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		} else if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update garden"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Garden updated successfully"}) // Or return the updated garden
}

// DeleteGardenHandler uses GardenStorer to delete a garden by ID.
func DeleteGardenHandler(storer storage.GardenStorer, c *gin.Context) {
	gardenID := c.Param("garden_id")
	if err := storer.DeleteGarden(gardenID); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete garden"})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
