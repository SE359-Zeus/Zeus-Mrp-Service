package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"zeus-sales-service/internal/controllers"
	"zeus-sales-service/internal/middlewares"
	"zeus-sales-service/internal/repository/sqlite"
	"zeus-sales-service/internal/repository/valkey"
	"zeus-sales-service/internal/service"

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

	services := service.NewServices(sqliteRepo, valkeyRepo)
	router := middlewares.ErrorHandler(controllers.NewMux(services))

	log.Printf("Zeus Sales Service running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
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
