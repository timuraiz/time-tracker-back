package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time-tracker/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Define command line flags
	var (
		direction = flag.String("direction", "", "Migration direction: up, down")
		version   = flag.String("version", "", "Migrate to specific version")
		status    = flag.Bool("status", false, "Show current migration status")
	)
	flag.Parse()

	// Show current status
	if *status {
		showStatus()
		return
	}

	// Migrate to specific version
	if *version != "" {
		v, err := strconv.ParseUint(*version, 10, 32)
		if err != nil {
			log.Fatal("Invalid version number:", err)
		}

		err = database.MigrateToVersion(uint(v))
		if err != nil {
			log.Fatal("Migration failed:", err)
		}
		return
	}

	// Run directional migration
	if *direction == "" {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go -direction=up     # Run all pending migrations")
		fmt.Println("  go run cmd/migrate/main.go -direction=down   # Rollback one migration")
		fmt.Println("  go run cmd/migrate/main.go -version=1        # Migrate to specific version")
		fmt.Println("  go run cmd/migrate/main.go -status           # Show current status")
		os.Exit(1)
	}

	err = database.RunMigrations(*direction)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}

func showStatus() {
	version, dirty, err := database.GetCurrentVersion()
	if err != nil {
		log.Fatal("Failed to get migration status:", err)
	}

	fmt.Printf("Current migration version: %d\n", version)
	if dirty {
		fmt.Println("Status: DIRTY (migration failed, needs manual intervention)")
	} else {
		fmt.Println("Status: CLEAN")
	}
}