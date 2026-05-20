package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"zeus-sales-service/internal/controllers"
	"zeus-sales-service/internal/middlewares"
	"zeus-sales-service/internal/repository/sqlite"
	"zeus-sales-service/internal/repository/valkey"
	"zeus-sales-service/internal/service"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
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

	// Create a small gin router to serve OpenAPI UI (dark theme) similar to other services
	docsRouter := gin.New()
	docsRouter.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
		Title: "Zeus Sales API",
		SpecProvider: func() ([]byte, error) {
			data, err := os.ReadFile("docs/openapi.yaml")
			if err != nil {
				return nil, err
			}
			var parsed any
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				return nil, err
			}
			return json.Marshal(parsed)
		},
		Theme: "dark",
	}))

	// Mount docs router and the API router under a main mux
	mainMux := http.NewServeMux()
	mainMux.Handle("/", router)
	mainMux.Handle("/docs/", docsRouter)
	mainMux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusFound)
	})

	log.Printf("Zeus Sales Service running on :8080")
	if err := http.ListenAndServe(":8080", mainMux); err != nil {
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
