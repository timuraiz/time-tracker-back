package routes

import (
	"time-tracker/handlers"
	"time-tracker/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// Disable automatic redirect of trailing slash
	r.RedirectTrailingSlash = false

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")

	// Time entry routes (requires authentication)
	timeEntries := api.Group("/time-entries")
	timeEntries.Use(middleware.SupabaseAuth()) // Apply authentication middleware
	{
		timeEntries.POST("", handlers.CreateTimeEntry)
		timeEntries.GET("", handlers.GetTimeEntries)
		timeEntries.GET("/:id", handlers.GetTimeEntry)
		timeEntries.PUT("/:id", handlers.UpdateTimeEntry)
		timeEntries.POST("/:id/stop", handlers.StopTimeEntry)
		timeEntries.DELETE("/:id", handlers.DeleteTimeEntry)
	}

	// Project routes (requires authentication)
	projects := api.Group("/projects")
	projects.Use(middleware.SupabaseAuth()) // Apply authentication middleware
	{
		projects.POST("", handlers.CreateProject)
		projects.GET("", handlers.GetProjects)
		projects.GET("/:id", handlers.GetProject)
		projects.PUT("/:id", handlers.UpdateProject)
		projects.DELETE("/:id", handlers.DeleteProject)
	}

	// Profile routes (requires authentication)
	profile := api.Group("/profile")
	profile.Use(middleware.SupabaseAuth()) // Apply authentication middleware
	{
		profile.POST("", handlers.CreateProfile)
		profile.GET("", handlers.GetProfile)
		profile.POST("/picture", handlers.UploadProfilePicture)
		profile.DELETE("/picture", handlers.DeleteProfilePicture)
	}

	return r
}
