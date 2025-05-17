package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
	"github.com/zjpiazza/plantastic/internal/models"
)

func ListBedsHandler(storer storage.BedStorer, c *gin.Context) {
	beds, err := storer.GetAllBeds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch beds"})
		return
	}
	c.JSON(http.StatusOK, beds)
}

func CreateBedHandler(storer storage.BedStorer, c *gin.Context) {
	var bed models.Bed
	if err := c.ShouldBindJSON(&bed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := storer.CreateBed(&bed); err != nil {
		if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		} else if err == storage.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bed"})
		return
	}
	c.JSON(http.StatusCreated, bed)
}

func GetBedHandler(storer storage.BedStorer, c *gin.Context) {
	bedID := c.Param("bed_id")
	bed, err := storer.GetBedByID(bedID)
	if err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bed not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bed"})
		return
	}
	c.JSON(http.StatusOK, bed)
}

func UpdateBedHandler(storer storage.BedStorer, c *gin.Context) {
	bedID := c.Param("bed_id")
	var bedUpdates models.Bed
	if err := c.ShouldBindJSON(&bedUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	bedUpdates.ID = bedID // Ensure ID from path is used

	if err := storer.UpdateBed(&bedUpdates); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bed not found"})
			return
		} else if err == storage.ErrValidation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update bed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bed updated successfully"})
}

func DeleteBedHandler(storer storage.BedStorer, c *gin.Context) {
	bedID := c.Param("bed_id")
	if err := storer.DeleteBed(bedID); err != nil {
		if err == storage.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Bed not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete bed"})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
