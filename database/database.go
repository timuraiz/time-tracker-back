package database

import (
	"log"
	"os"
	"time"
	"time-tracker/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to PostgreSQL database
	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Printf("Database connection failed. URL: %s", databaseURL)
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool and timeouts
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
}

func Migrate() {
	// Note: For development, you can use GORM AutoMigrate
	// For production, use versioned migrations with: make migrate-up

	// Enable UUID extension for PostgreSQL
	err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		log.Fatal("Failed to create UUID extension:", err)
	}

	// Option 1: Use GORM AutoMigrate (for development)
	err = DB.AutoMigrate(&models.TimeEntry{}, &models.Project{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Option 2: Use versioned migrations (recommended for production)
	// Uncomment the lines below and comment out AutoMigrate above
	// err = RunMigrations("up")
	// if err != nil {
	//     log.Fatal("Failed to run migrations:", err)
	// }

	log.Println("Database migrated successfully")
}
