package main

import (
	"log"

	"zeus-system-service/internal/repository/sqlite"
	"zeus-system-service/seeder"
)

func main() {
	db, err := sqlite.NewDB("system.db")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	log.Println("Running AutoMigrate...")
	if err := sqlite.AutoMigrate(db); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := seeder.SeedAll(db); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seed complete.")
}
