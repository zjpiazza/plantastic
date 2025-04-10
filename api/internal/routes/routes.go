package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/api/internal/handlers"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Garden Routes
	r.GET("/gardens", func(c *gin.Context) {
		handlers.ListGardensHandler(db, c)
	})
	r.POST("/gardens", func(c *gin.Context) {
		handlers.CreateGardenHandler(db, c)
	})
	r.GET("/gardens/:garden_id", func(c *gin.Context) {
		handlers.GetGardenHandler(db, c)
	})
	r.PUT("/gardens/:garden_id", func(c *gin.Context) {
		handlers.UpdateGardenHandler(db, c)
	})
	r.DELETE("/gardens/:garden_id", func(c *gin.Context) {
		handlers.DeleteGardenHandler(db, c)
	})

	// Bed Routes
	r.GET("/beds", func(c *gin.Context) {
		handlers.ListBedsHandler(db, c)
	})
	r.POST("/beds", func(c *gin.Context) {
		handlers.CreateBedHandler(db, c)
	})
	r.GET("/beds/:bed_id", func(c *gin.Context) {
		handlers.GetBedHandler(db, c)
	})
	r.PUT("/beds/:bed_id", func(c *gin.Context) {
		handlers.UpdateBedHandler(db, c)
	})
	r.DELETE("/beds/:bed_id", func(c *gin.Context) {
		handlers.DeleteBedHandler(db, c)
	})

	// Task Routes
	r.GET("/tasks", func(c *gin.Context) {
		handlers.ListTasksHandler(db, c)
	})
	r.POST("/tasks", func(c *gin.Context) {
		handlers.CreateTaskHandler(db, c)
	})
	r.GET("/tasks/:task_id", func(c *gin.Context) {
		handlers.GetTaskHandler(db, c)
	})
	r.PUT("/tasks/:task_id", func(c *gin.Context) {
		handlers.UpdateTaskHandler(db, c)
	})
	r.DELETE("/tasks/:task_id", func(c *gin.Context) {
		handlers.DeleteTaskHandler(db, c)
	})

	return r
}
