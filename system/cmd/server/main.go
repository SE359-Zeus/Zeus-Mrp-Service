package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"os"

	"zeus-system-service/internal/config"
	"zeus-system-service/internal/handler"
	"zeus-system-service/internal/handler/middleware"
	"zeus-system-service/internal/repository/sqlite"
	"zeus-system-service/internal/service"

	openapiui "github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func loadPrivateKey(path string) *rsa.PrivateKey {
	if path == "" {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("failed to generate dev RSA key: %v", err)
		}
		log.Println("using ephemeral RSA key (dev mode)")
		return key
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read private key: %v", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		log.Fatalf("failed to decode PEM block from %s", path)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatalf("failed to parse private key: %v", err)
		}
	}
	return key.(*rsa.PrivateKey)
}

func main() {
	cfg := config.Load()

	db, err := sqlite.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := sqlite.AutoMigrate(db); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	userRepo := sqlite.NewUserRepository(db)
	refreshTokenRepo := sqlite.NewRefreshTokenRepository(db)
	auditRepo := sqlite.NewAuditRepository(db)

	userSvc := service.NewUserService(userRepo)
	privateKey := loadPrivateKey(cfg.JWTKeyPath)
	authSvc := service.NewAuthService(userSvc, refreshTokenRepo, privateKey)
	auditSvc := service.NewAuditService(auditRepo)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	auditH := handler.NewAuditHandler(auditSvc)

	r := gin.Default()

	public := r.Group("/")
	{
		public.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
			Title: "Zeus System API",
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
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
		public.POST("/api/v1/auth/login", authH.Login)
		public.POST("/api/v1/auth/refresh", authH.Refresh)
	}

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(authSvc))
	{

		api.POST("/users", userH.Create)
		api.GET("/users", userH.List)
		api.GET("/users/:id", userH.GetByID)
		api.PUT("/users/:id", userH.Update)
		api.PATCH("/users/:id/status", userH.SetStatus)

		api.POST("/logs/ingest", auditH.Ingest)
		api.GET("/logs", auditH.Query)
		api.GET("/logs/metrics", auditH.GetMetrics)
	}

	log.Printf("Zeus System service starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
