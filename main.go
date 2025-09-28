package main

import (
	"log"
	"os"
	"time-tracker/database"
	"time-tracker/routes"
	"time-tracker/supabase"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize Supabase client
	supabase.InitClient()

	// Connect to database
	database.Connect()

	// Run migrations
	database.Migrate()

	// Setup routes
	r := routes.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
