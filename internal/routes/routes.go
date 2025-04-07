package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/zjpiazza/plantastic/internal/handlers"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Garden Routes
	r.GET("/gardens", func(c *gin.Context) {
		handlers.ListGardensHandler(db, c.Writer, c.Request)
	})
	r.POST("/gardens", func(c *gin.Context) {
		handlers.CreateGardenHandler(db, c.Writer, c.Request)
	})
	r.GET("/gardens/:garden_id", func(c *gin.Context) {
		handlers.GetGardenHandler(db, c.Writer, c.Request)
	})
	r.PUT("/gardens/:garden_id", func(c *gin.Context) {
		handlers.UpdateGardenHandler(db, c.Writer, c.Request)
	})
	r.DELETE("/gardens/:garden_id", func(c *gin.Context) {
		handlers.DeleteGardenHandler(db, c.Writer, c.Request)
	})

	// Bed Routes
	r.GET("/gardens/:garden_id/beds", func(c *gin.Context) {
		handlers.ListBedsHandler(db, c.Writer, c.Request)
	})
	r.POST("/gardens/:garden_id/beds", func(c *gin.Context) {
		handlers.CreateBedHandler(db, c.Writer, c.Request)
	})
	r.GET("/beds/:bed_id", func(c *gin.Context) {
		handlers.GetBedHandler(db, c.Writer, c.Request)
	})
	r.PUT("/beds/:bed_id", func(c *gin.Context) {
		handlers.UpdateBedHandler(db, c.Writer, c.Request)
	})
	r.DELETE("/beds/:bed_id", func(c *gin.Context) {
		handlers.DeleteBedHandler(db, c.Writer, c.Request)
	})

	// Task Routes
	r.GET("/gardens/:garden_id/tasks", func(c *gin.Context) {
		handlers.ListTasksHandler(db, c.Writer, c.Request)
	})
	r.POST("/gardens/:garden_id/tasks", func(c *gin.Context) {
		handlers.CreateTaskHandler(db, c.Writer, c.Request)
	})
	r.GET("/tasks/:task_id", func(c *gin.Context) {
		handlers.GetTaskHandler(db, c.Writer, c.Request)
	})
	r.PUT("/tasks/:task_id", func(c *gin.Context) {
		handlers.UpdateTaskHandler(db, c.Writer, c.Request)
	})
	r.DELETE("/tasks/:task_id", func(c *gin.Context) {
		handlers.DeleteTaskHandler(db, c.Writer, c.Request)
	})

	return r
}
