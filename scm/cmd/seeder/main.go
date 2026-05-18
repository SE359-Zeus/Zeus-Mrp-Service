package main

import (
	"log"
	"zeus-scm-service/internal/repository/sqlite"
	"zeus-scm-service/seeder"
)

func main() {
	// 1. Initialize DB
	db, err := sqlite.NewDB("scm.db")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 2. Run GORM AutoMigrate for basic structures
	log.Println("Running AutoMigrate...")
	if err := sqlite.AutoMigrate(db); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// 3. Run SQL Migrations (000001 - 000005) for FKs and LUTs
	log.Println("Running SQL Migrations...")
	if err := sqlite.RunMigrations(db, "internal/migration"); err != nil {
		log.Printf("Migration warning (might already be up to date): %v", err)
	}

	// 4. Seed Data
	if err := seeder.SeedAll(db); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Process complete.")
}
