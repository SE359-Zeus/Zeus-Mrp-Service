package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"zeus-sales-service/internal/repository/sqlite"
	"zeus-sales-service/internal/repository/valkey"
	"zeus-sales-service/seeder"

	"github.com/redis/go-redis/v9"
)

func main() {
	dbPath := getenv("SALES_SQLITE_DB", filepath.Join("configs", "sales.db"))
	valkeyAddr := getenv("SALES_VALKEY_ADDR", "localhost:6379")

	sqliteRepo, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer sqliteRepo.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: valkeyAddr})
	defer redisClient.Close()
	valkeyRepo := valkey.New(redisClient)

	if err := seeder.SeedAll(context.Background(), sqliteRepo, valkeyRepo); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
