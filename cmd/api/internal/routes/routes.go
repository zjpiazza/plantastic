package routes

import (
	"github.com/gin-gonic/gin"
	devicehandlers "github.com/zjpiazza/plantastic/cmd/api/handlers"
	"github.com/zjpiazza/plantastic/cmd/api/internal/handlers"
	"github.com/zjpiazza/plantastic/cmd/api/internal/storage"
)

func SetupRoutes(gardenStore storage.GardenStorer, bedStore storage.BedStorer, taskStore storage.TaskStorer) *gin.Engine {
	r := gin.Default()

	// Garden Routes
	r.GET("/gardens", func(c *gin.Context) {
		handlers.ListGardensHandler(gardenStore, c)
	})
	r.POST("/gardens", func(c *gin.Context) {
		handlers.CreateGardenHandler(gardenStore, c)
	})
	r.GET("/gardens/:garden_id", func(c *gin.Context) {
		handlers.GetGardenHandler(gardenStore, c)
	})
	r.PUT("/gardens/:garden_id", func(c *gin.Context) {
		handlers.UpdateGardenHandler(gardenStore, c)
	})
	r.DELETE("/gardens/:garden_id", func(c *gin.Context) {
		handlers.DeleteGardenHandler(gardenStore, c)
	})

	// Bed Routes
	r.GET("/beds", func(c *gin.Context) {
		handlers.ListBedsHandler(bedStore, c)
	})
	r.POST("/beds", func(c *gin.Context) {
		handlers.CreateBedHandler(bedStore, c)
	})
	r.GET("/beds/:bed_id", func(c *gin.Context) {
		handlers.GetBedHandler(bedStore, c)
	})
	r.PUT("/beds/:bed_id", func(c *gin.Context) {
		handlers.UpdateBedHandler(bedStore, c)
	})
	r.DELETE("/beds/:bed_id", func(c *gin.Context) {
		handlers.DeleteBedHandler(bedStore, c)
	})

	// Task Routes
	r.GET("/tasks", func(c *gin.Context) {
		handlers.ListTasksHandler(taskStore, c)
	})
	r.POST("/tasks", func(c *gin.Context) {
		handlers.CreateTaskHandler(taskStore, c)
	})
	r.GET("/tasks/:task_id", func(c *gin.Context) {
		handlers.GetTaskHandler(taskStore, c)
	})
	r.PUT("/tasks/:task_id", func(c *gin.Context) {
		handlers.UpdateTaskHandler(taskStore, c)
	})
	r.DELETE("/tasks/:task_id", func(c *gin.Context) {
		handlers.DeleteTaskHandler(taskStore, c)
	})

	return r
}

func SetupProtectedRoutes(rg *gin.RouterGroup, gardenStore storage.GardenStorer, bedStore storage.BedStorer, taskStore storage.TaskStorer, deviceHandler *devicehandlers.DeviceHandler) {
	// Garden Routes
	rg.GET("/gardens", func(c *gin.Context) {
		handlers.ListGardensHandler(gardenStore, c)
	})
	rg.POST("/gardens", func(c *gin.Context) {
		handlers.CreateGardenHandler(gardenStore, c)
	})
	rg.GET("/gardens/:garden_id", func(c *gin.Context) {
		handlers.GetGardenHandler(gardenStore, c)
	})
	rg.PUT("/gardens/:garden_id", func(c *gin.Context) {
		handlers.UpdateGardenHandler(gardenStore, c)
	})
	rg.DELETE("/gardens/:garden_id", func(c *gin.Context) {
		handlers.DeleteGardenHandler(gardenStore, c)
	})

	// Bed Routes
	rg.GET("/beds", func(c *gin.Context) {
		handlers.ListBedsHandler(bedStore, c)
	})
	rg.POST("/beds", func(c *gin.Context) {
		handlers.CreateBedHandler(bedStore, c)
	})
	rg.GET("/beds/:bed_id", func(c *gin.Context) {
		handlers.GetBedHandler(bedStore, c)
	})
	rg.PUT("/beds/:bed_id", func(c *gin.Context) {
		handlers.UpdateBedHandler(bedStore, c)
	})
	rg.DELETE("/beds/:bed_id", func(c *gin.Context) {
		handlers.DeleteBedHandler(bedStore, c)
	})

	// Task Routes
	rg.GET("/tasks", func(c *gin.Context) {
		handlers.ListTasksHandler(taskStore, c)
	})
	rg.POST("/tasks", func(c *gin.Context) {
		handlers.CreateTaskHandler(taskStore, c)
	})
	rg.GET("/tasks/:task_id", func(c *gin.Context) {
		handlers.GetTaskHandler(taskStore, c)
	})
	rg.PUT("/tasks/:task_id", func(c *gin.Context) {
		handlers.UpdateTaskHandler(taskStore, c)
	})
	rg.DELETE("/tasks/:task_id", func(c *gin.Context) {
		handlers.DeleteTaskHandler(taskStore, c)
	})

	// Device Authentication Routes
	deviceAuthGroup := rg.Group("/device") // Prefixing with /device
	{
		deviceAuthGroup.POST("/link", deviceHandler.AuthenticateDevice) // This remains protected
	}
}
