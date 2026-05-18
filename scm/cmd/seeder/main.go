package main

import (
	"log"
	"zeus-scm-service/internal/repository/sqlite"
	"zeus-scm-service/seeder"
)

func main() {
	// Initialize DB (adjust path if needed, usually scm.db or relative)
	db, err := sqlite.NewDB("scm.db")
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := seeder.SeedAll(db); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding complete.")
}
