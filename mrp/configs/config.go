package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	BaseURL         string
	Env             string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() *Config {
	return &Config{
		Port:            getEnv("MRP_PORT", "8081"),
		BaseURL:         getEnv("MRP_BASE_URL", "http://localhost:8081"),
		Env:             getEnv("MRP_ENV", "development"),
		ReadTimeout:     time.Duration(getEnvInt("MRP_READ_TIMEOUT_SEC", 15)) * time.Second,
		WriteTimeout:    time.Duration(getEnvInt("MRP_WRITE_TIMEOUT_SEC", 15)) * time.Second,
		ShutdownTimeout: time.Duration(getEnvInt("MRP_SHUTDOWN_TIMEOUT_SEC", 10)) * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}

	return parsed
}
